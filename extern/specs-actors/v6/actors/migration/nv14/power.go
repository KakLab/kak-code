package nv14

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	builtin5 "github.com/filecoin-project/specs-actors/v5/actors/builtin"
	power5 "github.com/filecoin-project/specs-actors/v5/actors/builtin/power"
	adt5 "github.com/filecoin-project/specs-actors/v5/actors/util/adt"
	adt6 "github.com/filecoin-project/specs-actors/v6/actors/util/adt"
	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"

	builtin6 "github.com/filecoin-project/specs-actors/v6/actors/builtin"
	power6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/power"
	smoothing6 "github.com/filecoin-project/specs-actors/v6/actors/util/smoothing"
)

type powerMigrator struct{}

func (m powerMigrator) migrateState(ctx context.Context, store cbor.IpldStore, in actorMigrationInput) (*actorMigrationResult, error) {
	var inState power5.State
	if err := store.Get(ctx, in.head, &inState); err != nil {
		return nil, err
	}

	claimsOut, err := m.migrateClaims(ctx, store, inState.Claims)
	if err != nil {
		return nil, err
	}

	outState := power6.State{
		TotalRawBytePower:         inState.TotalRawBytePower,
		TotalBytesCommitted:       inState.TotalBytesCommitted,
		TotalQualityAdjPower:      inState.TotalQualityAdjPower,
		TotalQABytesCommitted:     inState.TotalQABytesCommitted,
		TotalPledgeCollateral:     inState.TotalPledgeCollateral,
		ThisEpochRawBytePower:     inState.ThisEpochRawBytePower,
		ThisEpochQualityAdjPower:  inState.ThisEpochQualityAdjPower,
		ThisEpochPledgeCollateral: inState.ThisEpochPledgeCollateral,
		ThisEpochQAPowerSmoothed:  smoothing6.FilterEstimate(inState.ThisEpochQAPowerSmoothed),
		MinerCount:                inState.MinerCount,
		MinerAboveMinPowerCount:   inState.MinerAboveMinPowerCount,
		CronEventQueue:            inState.CronEventQueue,
		FirstCronEpoch:            inState.FirstCronEpoch,
		Claims:                    claimsOut,
		ProofValidationBatch:      inState.ProofValidationBatch,
		TotalPos:                  inState.TotalPos,
		TotalKakSectorSize:        abi.SectorSize(inState.TotalRawBytePower.Int64()),
	}
	newHead, err := store.Put(ctx, &outState)
	return &actorMigrationResult{
		newCodeCID: m.migratedCodeCID(),
		newHead:    newHead,
	}, err
}

func (m powerMigrator) migratedCodeCID() cid.Cid {
	return builtin6.StoragePowerActorCodeID
}

func (m powerMigrator) migrateClaims(ctx context.Context, store cbor.IpldStore, root cid.Cid) (cid.Cid, error) {
	astore := adt5.WrapStore(ctx, store)
	inClaims, err := adt5.AsMap(astore, root, builtin5.DefaultHamtBitwidth)
	if err != nil {
		return cid.Undef, err
	}
	outClaims, err := adt6.MakeEmptyMap(astore, builtin5.DefaultHamtBitwidth)
	if err != nil {
		return cid.Undef, err
	}

	var inClaim power5.Claim
	if err = inClaims.ForEach(&inClaim, func(key string) error {
		outClaim := power6.Claim{
			WindowPoStProofType: inClaim.WindowPoStProofType,
			RawBytePower:        inClaim.RawBytePower,
			QualityAdjPower:     inClaim.QualityAdjPower,
			PosPower:            inClaim.PosPower,
			KTatolSize:          0,
		}
		return outClaims.Put(StringKey(key), &outClaim)
	}); err != nil {
		return cid.Undef, err
	}
	return outClaims.Root()
}
