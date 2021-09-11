package nv12

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/rt"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	states3 "github.com/filecoin-project/specs-actors/v3/actors/states"
	builtin4 "github.com/filecoin-project/specs-actors/v4/actors/builtin"
	states4 "github.com/filecoin-project/specs-actors/v4/actors/states"
	adt4 "github.com/filecoin-project/specs-actors/v4/actors/util/adt"

	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

// Config parameterizes a state tree migration
type Config struct {
	// Number of migration worker goroutines to run.
	// More workers enables higher CPU utilization doing migration computations (including state encoding)
	MaxWorkers uint
	// Capacity of the queue of jobs available to workers (zero for unbuffered).
	// A queue length of hundreds to thousands improves throughput at the cost of memory.
	JobQueueSize uint
	// Capacity of the queue receiving migration results from workers, for persisting (zero for unbuffered).
	// A queue length of tens to hundreds improves throughput at the cost of memory.
	ResultQueueSize uint
	// Time between progress logs to emit.
	// Zero (the default) results in no progress logs.
	ProgressLogPeriod time.Duration
}

type Logger interface {
	// This is the same logging interface provided by the Runtime
	Log(level rt.LogLevel, msg string, args ...interface{})
}

func ActorHeadKey(addr address.Address, head cid.Cid) string {
	return addr.String() + "-h-" + head.String()
}

// Migrates from v11 to v12
//
// This migration updates all miner actor states to record their deadline cron status.
// In v12 miner actors defer scheduling deadline cron until beginning to prove storage
// and stop running deadline cron when they stop proving storage so this value can be false.
// All miner actors in v11 have active deadline crons so this field is set to true in migration.
// MigrationCache stores and loads cached data. Its implementation must be threadsafe
type MigrationCache interface {
	Write(key string, newCid cid.Cid) error
	Read(key string) (bool, cid.Cid, error)
	Load(key string, loadFunc func() (cid.Cid, error)) (cid.Cid, error)
}

// Migrates the filecoin state tree starting from the global state tree and upgrading all actor state.
// The store must support concurrent writes (even if the configured worker count is 1).
func MigrateStateTree(ctx context.Context, store cbor.IpldStore, actorsRootIn cid.Cid, priorEpoch abi.ChainEpoch, cfg Config, log Logger, cache MigrationCache) (cid.Cid, error) {
	if cfg.MaxWorkers <= 0 {
		return cid.Undef, xerrors.Errorf("invalid migration config with %d workers", cfg.MaxWorkers)
	}

	// Maps prior version code CIDs to migration functions.
	var migrations = map[cid.Cid]actorMigration{
		builtin3.AccountActorCodeID:          nilMigrator{builtin4.AccountActorCodeID},
		builtin3.CronActorCodeID:             nilMigrator{builtin4.CronActorCodeID},
		builtin3.InitActorCodeID:             nilMigrator{builtin4.InitActorCodeID},
		builtin3.MultisigActorCodeID:         nilMigrator{builtin4.MultisigActorCodeID},
		builtin3.PaymentChannelActorCodeID:   nilMigrator{builtin4.PaymentChannelActorCodeID},
		builtin3.RewardActorCodeID:           nilMigrator{builtin4.RewardActorCodeID},
		builtin3.StorageMarketActorCodeID:    nilMigrator{builtin4.StorageMarketActorCodeID},
		builtin3.StorageMinerActorCodeID:     cachedMigration(cache, minerMigrator{}),
		builtin3.StoragePowerActorCodeID:     cachedMigration(cache, powerMigrator{}),
		builtin3.SystemActorCodeID:           nilMigrator{builtin4.SystemActorCodeID},
		builtin3.VerifiedRegistryActorCodeID: nilMigrator{builtin4.VerifiedRegistryActorCodeID},
	}

	// Set of prior version code CIDs for actors to defer during iteration, for explicit migration afterwards.
	var deferredCodeIDs = map[cid.Cid]struct{}{
		// None
	}

	if len(migrations)+len(deferredCodeIDs) != 11 {
		panic(fmt.Sprintf("incomplete migration specification with %d code CIDs", len(migrations)))
	}
	startTime := time.Now()

	// Load input and output state trees
	adtStore := adt4.WrapStore(ctx, store)
	actorsIn, err := states3.LoadTree(adtStore, actorsRootIn)
	if err != nil {
		return cid.Undef, err
	}
	actorsOut, err := states4.NewTree(adtStore)
	if err != nil {
		return cid.Undef, err
	}

	// Setup synchronization
	grp, ctx := errgroup.WithContext(ctx)
	// Input and output queues for workers.
	jobCh := make(chan *migrationJob, cfg.JobQueueSize)
	jobResultCh := make(chan *migrationJobResult, cfg.ResultQueueSize)
	// Atomically-modified counters for logging progress
	var jobCount uint32
	var doneCount uint32

	// Iterate all actors in old state root to create migration jobs for each non-deferred actor.
	grp.Go(func() error {
		defer close(jobCh)
		log.Log(rt.INFO, "Creating migration jobs for tree %s", actorsRootIn)
		if err = actorsIn.ForEach(func(addr address.Address, actorIn *states3.Actor) error {
			if _, ok := deferredCodeIDs[actorIn.Code]; ok {
				return nil // Deferred for explicit migration later.
			}
			nextInput := &migrationJob{
				Address:        addr,
				Actor:          *actorIn, // Must take a copy, the pointer is not stable.
				cache:          cache,
				actorMigration: migrations[actorIn.Code],
			}
			select {
			case jobCh <- nextInput:
			case <-ctx.Done():
				return ctx.Err()
			}
			atomic.AddUint32(&jobCount, 1)
			return nil
		}); err != nil {
			return err
		}
		log.Log(rt.INFO, "Done creating %d migration jobs for tree %s after %v", jobCount, actorsRootIn, time.Since(startTime))
		return nil
	})

	// Worker threads run jobs.
	var workerWg sync.WaitGroup
	for i := uint(0); i < cfg.MaxWorkers; i++ {
		workerWg.Add(1)
		workerId := i
		grp.Go(func() error {
			defer workerWg.Done()
			for job := range jobCh {
				result, err := job.run(ctx, store, priorEpoch)
				if err != nil {
					return err
				}
				select {
				case jobResultCh <- result:
				case <-ctx.Done():
					return ctx.Err()
				}
				atomic.AddUint32(&doneCount, 1)
			}
			log.Log(rt.INFO, "Worker %d done", workerId)
			return nil
		})
	}
	log.Log(rt.INFO, "Started %d workers", cfg.MaxWorkers)

	// Monitor the job queue. This non-critical goroutine is outside the errgroup and exits when
	// workersFinished is closed, or the context done.
	workersFinished := make(chan struct{}) // Closed when waitgroup is emptied.
	if cfg.ProgressLogPeriod > 0 {
		go func() {
			defer log.Log(rt.DEBUG, "Job queue monitor done")
			for {
				select {
				case <-time.After(cfg.ProgressLogPeriod):
					jobsNow := jobCount // Snapshot values to avoid incorrect-looking arithmetic if they change.
					doneNow := doneCount
					pendingNow := jobsNow - doneNow
					elapsed := time.Since(startTime)
					rate := float64(doneNow) / elapsed.Seconds()
					log.Log(rt.INFO, "%d jobs created, %d done, %d pending after %v (%.0f/s)",
						jobsNow, doneNow, pendingNow, elapsed, rate)
				case <-workersFinished:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Close result channel when workers are done sending to it.
	grp.Go(func() error {
		workerWg.Wait()
		close(jobResultCh)
		close(workersFinished)
		log.Log(rt.INFO, "All workers done after %v", time.Since(startTime))
		return nil
	})

	// Insert migrated records in output state tree and accumulators.
	grp.Go(func() error {
		log.Log(rt.INFO, "Result writer started")
		resultCount := 0
		for result := range jobResultCh {
			if err := actorsOut.SetActor(result.Address, &result.Actor); err != nil {
				return err
			}
			resultCount++
		}
		log.Log(rt.INFO, "Result writer wrote %d results to state tree after %v", resultCount, time.Since(startTime))
		return nil
	})

	if err := grp.Wait(); err != nil {
		return cid.Undef, err
	}

	elapsed := time.Since(startTime)
	rate := float64(doneCount) / elapsed.Seconds()
	log.Log(rt.INFO, "All %d done after %v (%.0f/s). Flushing state tree root.", doneCount, elapsed, rate)
	return actorsOut.Flush()
}

type actorMigrationInput struct {
	address    address.Address // actor's address
	balance    abi.TokenAmount // actor's balance
	head       cid.Cid         // actor's state head CID
	priorEpoch abi.ChainEpoch  // epoch of last state transition prior to migration
	cache      MigrationCache  // cache of existing cid -> cid migrations for this actor
}

type actorMigrationResult struct {
	newCodeCID cid.Cid
	newHead    cid.Cid
}

type actorMigration interface {
	// Loads an actor's state from an input store and writes new state to an output store.
	// Returns the new state head CID.
	migrateState(ctx context.Context, store cbor.IpldStore, input actorMigrationInput) (result *actorMigrationResult, err error)
	migratedCodeCID() cid.Cid
}

type migrationJob struct {
	address.Address
	states3.Actor
	actorMigration
	cache MigrationCache
}
type migrationJobResult struct {
	address.Address
	states4.Actor
}

func (job *migrationJob) run(ctx context.Context, store cbor.IpldStore, priorEpoch abi.ChainEpoch) (*migrationJobResult, error) {
	result, err := job.migrateState(ctx, store, actorMigrationInput{
		address:    job.Address,
		balance:    job.Actor.Balance,
		head:       job.Actor.Head,
		priorEpoch: priorEpoch,
		cache:      job.cache,
	})
	if err != nil {
		return nil, xerrors.Errorf("state migration failed for %s actor, addr %s: %w",
			builtin3.ActorNameByCode(job.Actor.Code), job.Address, err)
	}

	// Set up new actor record with the migrated state.
	return &migrationJobResult{
		job.Address, // Unchanged
		states4.Actor{
			Code:       result.newCodeCID,
			Head:       result.newHead,
			CallSeqNum: job.Actor.CallSeqNum, // Unchanged
			Balance:    job.Actor.Balance,    // Unchanged
		},
	}, nil
}

// Migrator which preserves the head CID and provides a fixed result code CID.
type nilMigrator struct {
	OutCodeCID cid.Cid
}

func (n nilMigrator) migrateState(_ context.Context, _ cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	return &actorMigrationResult{
		newCodeCID: n.OutCodeCID,
		newHead:    in.head,
	}, nil
}

func (n nilMigrator) migratedCodeCID() cid.Cid {
	return n.OutCodeCID
}

// Migrator that uses cached transformation if it exists
type cachedMigrator struct {
	cache MigrationCache
	actorMigration
}

func (c cachedMigrator) migrateState(ctx context.Context, store cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	newHead, err := c.cache.Load(ActorHeadKey(in.address, in.head), func() (cid.Cid, error) {
		result, err := c.actorMigration.migrateState(ctx, store, in)
		if err != nil {
			return cid.Undef, err
		}
		return result.newHead, nil
	})
	if err != nil {
		return nil, err
	}
	return &actorMigrationResult{
		newCodeCID: c.migratedCodeCID(),
		newHead:    newHead,
	}, nil
}

func cachedMigration(cache MigrationCache, m actorMigration) actorMigration {
	return cachedMigrator{
		actorMigration: m,
		cache:          cache,
	}
}
