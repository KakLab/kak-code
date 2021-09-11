package nv12

import (
	"context"
	builtin4 "github.com/filecoin-project/specs-actors/v4/actors/builtin"

	power3 "github.com/filecoin-project/specs-actors/v3/actors/builtin/power"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"

	power4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/power"
	smoothing4 "github.com/filecoin-project/specs-actors/v4/actors/util/smoothing"
)

type powerMigrator struct{}

func (m powerMigrator) migrateState(ctx context.Context, store cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	var inState power3.State
	if err := store.Get(ctx, in.head, &inState); err != nil {
		return nil, err
	}

	outState := power4.State{
		TotalRawBytePower:         inState.TotalRawBytePower,
		TotalBytesCommitted:       inState.TotalBytesCommitted,
		TotalQualityAdjPower:      inState.TotalQualityAdjPower,
		TotalQABytesCommitted:     inState.TotalQABytesCommitted,
		TotalPledgeCollateral:     inState.TotalPledgeCollateral,
		ThisEpochRawBytePower:     inState.ThisEpochRawBytePower,
		ThisEpochQualityAdjPower:  inState.ThisEpochQualityAdjPower,
		ThisEpochPledgeCollateral: inState.ThisEpochPledgeCollateral,
		ThisEpochQAPowerSmoothed:  smoothing4.FilterEstimate(inState.ThisEpochQAPowerSmoothed),
		MinerCount:                inState.MinerCount,
		MinerAboveMinPowerCount:   inState.MinerAboveMinPowerCount,
		CronEventQueue:            inState.CronEventQueue,
		FirstCronEpoch:            inState.FirstCronEpoch,
		Claims:                    inState.Claims,
		ProofValidationBatch:      inState.ProofValidationBatch,
		TotalPos:                  inState.TotalPos,
	}
	newHead, err := store.Put(ctx, &outState)
	return &actorMigrationResult{
		newCodeCID: m.migratedCodeCID(),
		newHead:    newHead,
	}, err
}

func (m powerMigrator) migratedCodeCID() cid.Cid {
	return builtin4.StoragePowerActorCodeID
}