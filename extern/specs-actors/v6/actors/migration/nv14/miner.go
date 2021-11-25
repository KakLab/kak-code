package nv14

import (
	"context"

	miner5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/miner"
	builtin6 "github.com/filecoin-project/specs-actors/v6/actors/builtin"
	miner6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/miner"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
)

type minerMigrator struct{}

func (m minerMigrator) migrateState(ctx context.Context, store cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	var inState miner5.State
	if err := store.Get(ctx, in.head, &inState); err != nil {
		return nil, err
	}

	outState := miner6.State{
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
		PosVestingFunds:           inState.PosVestingFunds,
		EmptyPreCommitSectors:     0,
		EmptyCommitSectors:        0,
	}
	newHead, err := store.Put(ctx, &outState)
	return &actorMigrationResult{
		newCodeCID: m.migratedCodeCID(),
		newHead:    newHead,
	}, err
}

func (m minerMigrator) migratedCodeCID() cid.Cid {
	return builtin6.StorageMinerActorCodeID
}
