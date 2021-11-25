package builtin

import (
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/cbor"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	builtin4 "github.com/filecoin-project/specs-actors/v4/actors/builtin"
	builtin5 "github.com/filecoin-project/specs-actors/v5/actors/builtin"
	builtin6 "github.com/filecoin-project/specs-actors/v6/actors/builtin"

	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/types"

	miner0 "github.com/filecoin-project/specs-actors/actors/builtin/miner"
	proof0 "github.com/filecoin-project/specs-actors/actors/runtime/proof"
	smoothing0 "github.com/filecoin-project/specs-actors/actors/util/smoothing"
	smoothing2 "github.com/filecoin-project/specs-actors/v2/actors/util/smoothing"
	smoothing3 "github.com/filecoin-project/specs-actors/v3/actors/util/smoothing"
	smoothing4 "github.com/filecoin-project/specs-actors/v4/actors/util/smoothing"
	smoothing5 "github.com/filecoin-project/specs-actors/v5/actors/util/smoothing"
	smoothing6 "github.com/filecoin-project/specs-actors/v6/actors/util/smoothing"
)

var SystemActorAddr = builtin0.SystemActorAddr
var BurntFundsActorAddr = builtin0.BurntFundsActorAddr
var CronActorAddr = builtin0.CronActorAddr
var SaftAddress = makeAddress("k0122")
var ReserveAddress = makeAddress("k090")
var RootVerifierAddress = makeAddress("k080")

var (
	ExpectedLeadersPerEpoch = builtin0.ExpectedLeadersPerEpoch
)

const (
	EpochDurationSeconds = builtin0.EpochDurationSeconds
	EpochsInDay          = builtin0.EpochsInDay
	SecondsInDay         = builtin0.SecondsInDay
)

const (
	MethodSend        = builtin4.MethodSend
	MethodConstructor = builtin4.MethodConstructor
)

// These are all just type aliases across actor versions 0, 2, & 3. In the future, that might change
// and we might need to do something fancier.
type SectorInfo = proof0.SectorInfo
type PoStProof = proof0.PoStProof
type FilterEstimate = smoothing0.FilterEstimate

func FromV0FilterEstimate(v0 smoothing0.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v0) //nolint:unconvert
}

// Doesn't change between actors v0, v2, and v3.
func QAPowerForWeight(size abi.SectorSize, duration abi.ChainEpoch, dealWeight, verifiedWeight abi.DealWeight) abi.StoragePower {
	return miner0.QAPowerForWeight(size, duration, dealWeight, verifiedWeight)
}

func FromV2FilterEstimate(v2 smoothing2.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v2)
}

func FromV3FilterEstimate(v3 smoothing3.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v3)
}

func FromV4FilterEstimate(v4 smoothing4.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v4)
}

func FromV5FilterEstimate(v5 smoothing5.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v5)
}

func FromV6FilterEstimate(v6 smoothing6.FilterEstimate) FilterEstimate {
	return (FilterEstimate)(v6)
}

type ActorStateLoader func(store adt.Store, root cid.Cid) (cbor.Marshaler, error)

var ActorStateLoaders = make(map[cid.Cid]ActorStateLoader)

func RegisterActorState(code cid.Cid, loader ActorStateLoader) {
	ActorStateLoaders[code] = loader
}

func Load(store adt.Store, act *types.Actor) (cbor.Marshaler, error) {
	loader, found := ActorStateLoaders[act.Code]
	if !found {
		return nil, xerrors.Errorf("unknown actor code %s", act.Code)
	}
	return loader(store, act.Head)
}

func ActorNameByCode(c cid.Cid) string {
	switch {
	case builtin0.IsBuiltinActor(c):
		return builtin0.ActorNameByCode(c)
	case builtin2.IsBuiltinActor(c):
		return builtin2.ActorNameByCode(c)
	case builtin3.IsBuiltinActor(c):
		return builtin3.ActorNameByCode(c)
	case builtin4.IsBuiltinActor(c):
		return builtin4.ActorNameByCode(c)
	case builtin5.IsBuiltinActor(c):
		return builtin5.ActorNameByCode(c)
	case builtin6.IsBuiltinActor(c):
		return builtin6.ActorNameByCode(c)
	default:
		return "<unknown>"
	}
}

func IsBuiltinActor(c cid.Cid) bool {
	return builtin0.IsBuiltinActor(c) ||
		builtin2.IsBuiltinActor(c) ||
		builtin3.IsBuiltinActor(c) ||
		builtin4.IsBuiltinActor(c) ||
		builtin5.IsBuiltinActor(c) ||
		builtin6.IsBuiltinActor(c)
}

func IsAccountActor(c cid.Cid) bool {
	return c == builtin0.AccountActorCodeID ||
		c == builtin2.AccountActorCodeID ||
		c == builtin3.AccountActorCodeID ||
		c == builtin4.AccountActorCodeID ||
		c == builtin5.AccountActorCodeID ||
		c == builtin6.AccountActorCodeID
}

func IsStorageMinerActor(c cid.Cid) bool {
	return c == builtin0.StorageMinerActorCodeID ||
		c == builtin2.StorageMinerActorCodeID ||
		c == builtin3.StorageMinerActorCodeID ||
		c == builtin4.StorageMinerActorCodeID ||
		c == builtin5.StorageMinerActorCodeID ||
		c == builtin6.StorageMinerActorCodeID
}

func IsMultisigActor(c cid.Cid) bool {
	return c == builtin0.MultisigActorCodeID ||
		c == builtin2.MultisigActorCodeID ||
		c == builtin3.MultisigActorCodeID ||
		c == builtin4.MultisigActorCodeID ||
		c == builtin5.MultisigActorCodeID ||
		c == builtin6.MultisigActorCodeID
}

func IsPaymentChannelActor(c cid.Cid) bool {
	return c == builtin0.PaymentChannelActorCodeID ||
		c == builtin2.PaymentChannelActorCodeID ||
		c == builtin3.PaymentChannelActorCodeID ||
		c == builtin4.PaymentChannelActorCodeID ||
		c == builtin5.PaymentChannelActorCodeID ||
		c == builtin6.PaymentChannelActorCodeID
}

func makeAddress(addr string) address.Address {
	ret, err := address.NewFromString(addr)
	if err != nil {
		panic(err)
	}

	return ret
}
