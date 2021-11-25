package nv13

import (
	"context"
	miner4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/miner"
	builtin5 "github.com/filecoin-project/specs-actors/v5/actors/builtin"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	"golang.org/x/xerrors"

	miner5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
)

type minerMigrator struct{}

func (m minerMigrator) migrateState(ctx context.Context, store cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	var inState miner4.State
	if err := store.Get(ctx, in.head, &inState); err != nil {
		return nil, err
	}

	initPosVestingFundsCid, err := store.Put(ctx, miner5.ConstructPosVestingFunds(inState.PosDeposits))
	if err != nil {
		return nil, xerrors.Errorf("failed to construct init pos vesting funds: %w", err)
	}

	outState := miner5.State{
		// No change
		Info:                      inState.Info,
		PreCommitDeposits:         inState.PreCommitDeposits,
		LockedFunds:               inState.LockedFunds,
		VestingFunds:              inState.VestingFunds,
		FeeDebt:                   inState.FeeDebt,
		InitialPledge:             inState.InitialPledge,
		PreCommittedSectors:       inState.PreCommittedSectors,
		PreCommittedSectorsExpiry: inState.PreCommittedSectorsExpiry,
		AllocatedSectors:          inState.AllocatedSectors,
		Sectors:                   inState.Sectors,
		ProvingPeriodStart:        inState.ProvingPeriodStart,
		CurrentDeadline:           inState.CurrentDeadline,
		Deadlines:                 inState.Deadlines,
		EarlyTerminations:         inState.EarlyTerminations,
		PosDeposits:               inState.PosDeposits,
		PosVestingFunds:           initPosVestingFundsCid,
		// Changed field
		DeadlineCronActive: true,
	}
	newHead, err := store.Put(ctx, &outState)
	return &actorMigrationResult{
		newCodeCID: m.migratedCodeCID(),
		newHead:    newHead,
	}, err
}

func (m minerMigrator) migratedCodeCID() cid.Cid {
	return builtin5.StorageMinerActorCodeID
}
