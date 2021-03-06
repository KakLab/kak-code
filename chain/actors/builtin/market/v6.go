package market

import (
	"bytes"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/types"

	market6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/market"
	adt6 "github.com/filecoin-project/specs-actors/v6/actors/util/adt"
)

var _ State = (*state6)(nil)

func load6(store adt.Store, root cid.Cid) (State, error) {
	out := state6{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type state6 struct {
	market6.State
	store adt.Store
}

func (s *state6) TotalLocked() (abi.TokenAmount, error) {
	fml := types.BigAdd(s.TotalClientLockedCollateral, s.TotalProviderLockedCollateral)
	fml = types.BigAdd(fml, s.TotalClientStorageFee)
	return fml, nil
}

func (s *state6) BalancesChanged(otherState State) (bool, error) {
	otherState2, ok := otherState.(*state6)
	if !ok {
		// there's no way to compare different versions of the state, so let's
		// just say that means the state of balances has changed
		return true, nil
	}
	return !s.State.EscrowTable.Equals(otherState2.State.EscrowTable) || !s.State.LockedTable.Equals(otherState2.State.LockedTable), nil
}

func (s *state6) StatesChanged(otherState State) (bool, error) {
	otherState2, ok := otherState.(*state6)
	if !ok {
		// there's no way to compare different versions of the state, so let's
		// just say that means the state of balances has changed
		return true, nil
	}
	return !s.State.States.Equals(otherState2.State.States), nil
}

func (s *state6) States() (DealStates, error) {
	stateArray, err := adt6.AsArray(s.store, s.State.States, market6.StatesAmtBitwidth)
	if err != nil {
		return nil, err
	}
	return &dealStates6{stateArray}, nil
}

func (s *state6) ProposalsChanged(otherState State) (bool, error) {
	otherState2, ok := otherState.(*state6)
	if !ok {
		// there's no way to compare different versions of the state, so let's
		// just say that means the state of balances has changed
		return true, nil
	}
	return !s.State.Proposals.Equals(otherState2.State.Proposals), nil
}

func (s *state6) Proposals() (DealProposals, error) {
	proposalArray, err := adt6.AsArray(s.store, s.State.Proposals, market6.ProposalsAmtBitwidth)
	if err != nil {
		return nil, err
	}
	return &dealProposals6{proposalArray}, nil
}

func (s *state6) EscrowTable() (BalanceTable, error) {
	bt, err := adt6.AsBalanceTable(s.store, s.State.EscrowTable)
	if err != nil {
		return nil, err
	}
	return &balanceTable6{bt}, nil
}

func (s *state6) LockedTable() (BalanceTable, error) {
	bt, err := adt6.AsBalanceTable(s.store, s.State.LockedTable)
	if err != nil {
		return nil, err
	}
	return &balanceTable6{bt}, nil
}

func (s *state6) VerifyDealsForActivation(
	minerAddr address.Address, deals []abi.DealID, currEpoch, sectorExpiry abi.ChainEpoch,
) (weight, verifiedWeight abi.DealWeight, err error) {
	w, vw, _, err := market6.ValidateDealsForActivation(&s.State, s.store, deals, minerAddr, sectorExpiry, currEpoch)
	return w, vw, err
}

func (s *state6) NextID() (abi.DealID, error) {
	return s.State.NextID, nil
}

type balanceTable6 struct {
	*adt6.BalanceTable
}

func (bt *balanceTable6) ForEach(cb func(address.Address, abi.TokenAmount) error) error {
	asMap := (*adt6.Map)(bt.BalanceTable)
	var ta abi.TokenAmount
	return asMap.ForEach(&ta, func(key string) error {
		a, err := address.NewFromBytes([]byte(key))
		if err != nil {
			return err
		}
		return cb(a, ta)
	})
}

type dealStates6 struct {
	adt.Array
}

func (s *dealStates6) Get(dealID abi.DealID) (*DealState, bool, error) {
	var deal2 market6.DealState
	found, err := s.Array.Get(uint64(dealID), &deal2)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	deal := fromV6DealState(deal2)
	return &deal, true, nil
}

func (s *dealStates6) ForEach(cb func(dealID abi.DealID, ds DealState) error) error {
	var ds1 market6.DealState
	return s.Array.ForEach(&ds1, func(idx int64) error {
		return cb(abi.DealID(idx), fromV6DealState(ds1))
	})
}

func (s *dealStates6) decode(val *cbg.Deferred) (*DealState, error) {
	var ds1 market6.DealState
	if err := ds1.UnmarshalCBOR(bytes.NewReader(val.Raw)); err != nil {
		return nil, err
	}
	ds := fromV6DealState(ds1)
	return &ds, nil
}

func (s *dealStates6) array() adt.Array {
	return s.Array
}

func fromV6DealState(v6 market6.DealState) DealState {
	return (DealState)(v6)
}

type dealProposals6 struct {
	adt.Array
}

func (s *dealProposals6) Get(dealID abi.DealID) (*DealProposal, bool, error) {
	var proposal2 market6.DealProposal
	found, err := s.Array.Get(uint64(dealID), &proposal2)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	proposal := fromV6DealProposal(proposal2)
	return &proposal, true, nil
}

func (s *dealProposals6) ForEach(cb func(dealID abi.DealID, dp DealProposal) error) error {
	var dp1 market6.DealProposal
	return s.Array.ForEach(&dp1, func(idx int64) error {
		return cb(abi.DealID(idx), fromV6DealProposal(dp1))
	})
}

func (s *dealProposals6) decode(val *cbg.Deferred) (*DealProposal, error) {
	var dp1 market6.DealProposal
	if err := dp1.UnmarshalCBOR(bytes.NewReader(val.Raw)); err != nil {
		return nil, err
	}
	dp := fromV6DealProposal(dp1)
	return &dp, nil
}

func (s *dealProposals6) array() adt.Array {
	return s.Array
}

func fromV6DealProposal(v6 market6.DealProposal) DealProposal {
	return (DealProposal)(v6)
}
