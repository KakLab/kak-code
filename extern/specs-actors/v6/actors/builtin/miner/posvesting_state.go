package miner

import (
	"sort"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
)

// PosVestingFunds represents the pos vesting table state for the miner.
// It is a slice of (VestingEpoch, VestingAmount).
// The slice will always be sorted by the VestingEpoch.
type PosVestingFunds struct {
	Funds []PosVestingFund
}

func (v *PosVestingFunds) getUnlockVestedFunds(currEpoch abi.ChainEpoch) abi.TokenAmount {
	amountUnlocked := abi.NewTokenAmount(0)
	for _, vf := range v.Funds {
		if vf.Epoch >= currEpoch {
			break
		}
		amountUnlocked = big.Add(amountUnlocked, vf.Amount)
	}
	return amountUnlocked
}

func (v *PosVestingFunds) unlockVestedFunds(currEpoch abi.ChainEpoch, amount abi.TokenAmount) abi.TokenAmount {
	amountUnlocked := abi.NewTokenAmount(0)

	lastIndexToRemove := -1
	for i, vf := range v.Funds {
		if vf.Epoch >= currEpoch {
			break
		}

		amountUnlocked = big.Add(amountUnlocked, vf.Amount)
		if amount.LessThanEqual(amountUnlocked) {
			v.Funds[i].Amount = big.Sub(amountUnlocked, amount)
			amountUnlocked = amount
			break
		}
		lastIndexToRemove = i
	}

	// remove all entries upto and including lastIndexToRemove
	if lastIndexToRemove != -1 {
		v.Funds = v.Funds[lastIndexToRemove+1:]
	}

	return amountUnlocked
}

func (v *PosVestingFunds) addLockedFunds(currEpoch abi.ChainEpoch, amount abi.TokenAmount) {
	// maps the epochs in PosVestingFunds to their indices in the slice
	entry := PosVestingFund{Epoch: currEpoch + PosVestPeriod, Amount: amount}
	v.Funds = append(v.Funds, entry)

	// sort slice by epoch
	sort.Slice(v.Funds, func(first, second int) bool {
		return v.Funds[first].Epoch < v.Funds[second].Epoch
	})
}

// PosVestingFunds represents miner funds that will vest at the given epoch.
type PosVestingFund struct {
	Epoch  abi.ChainEpoch // end
	Amount abi.TokenAmount
}

// ConstructVestingFunds constructs empty PosVestingFunds state.
func ConstructPosVestingFunds(posVestingFund abi.TokenAmount) *PosVestingFunds {
	v := new(PosVestingFunds)
	entry := PosVestingFund{Epoch: 0 + PosVestPeriod, Amount:posVestingFund}
	v.Funds = append(v.Funds, entry)
	return v
}
