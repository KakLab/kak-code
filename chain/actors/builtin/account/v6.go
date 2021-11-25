package account

import (
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/lotus/chain/actors/adt"

	account6 "github.com/filecoin-project/specs-actors/v6/actors/builtin/account"
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
	account6.State
	store adt.Store
}

func (s *state6) PubkeyAddress() (address.Address, error) {
	return s.Address, nil
}
