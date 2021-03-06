package reward

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/go-state-types/exitcode"
	rtt "github.com/filecoin-project/go-state-types/rt"
	reward0 "github.com/filecoin-project/specs-actors/actors/builtin/reward"
	"github.com/filecoin-project/specs-actors/v6/actors/builtin"
	"github.com/filecoin-project/specs-actors/v6/actors/runtime"
	"github.com/filecoin-project/specs-actors/v6/actors/util/smoothing"
	"github.com/ipfs/go-cid"
)

// PenaltyMultiplier is the factor miner penaltys are scaled up by
const PenaltyMultiplier = 3

type Actor struct{}

func (a Actor) Exports() []interface{} {
	return []interface{}{
		builtin.MethodConstructor: a.Constructor,
		2:                         a.AwardBlockReward,
		3:                         a.ThisEpochReward,
		4:                         a.UpdateNetworkKPI,
		5:                         a.StorageReward,
	}
}

func (a Actor) Code() cid.Cid {
	return builtin.RewardActorCodeID
}

func (a Actor) IsSingleton() bool {
	return true
}

func (a Actor) State() cbor.Er {
	return new(State)
}

var _ runtime.VMActor = Actor{}

func (a Actor) Constructor(rt runtime.Runtime, currRealizedPower *abi.StoragePower) *abi.EmptyValue {
	rt.ValidateImmediateCallerIs(builtin.SystemActorAddr)

	if currRealizedPower == nil {
		rt.Abortf(exitcode.ErrIllegalArgument, "argument should not be nil")
		return nil // linter does not understand abort exiting
	}
	st := ConstructState(*currRealizedPower)
	rt.StateCreate(st)
	return nil
}

//type AwardBlockRewardParams struct {
//	Miner     address.Address
//	Penalty   abi.TokenAmount // penalty for including bad messages in a block, >= 0
//	GasReward abi.TokenAmount // gas reward from all gas fees in a block, >= 0
//	WinCount  int64           // number of reward units won, > 0
//}
type AwardBlockRewardParams = reward0.AwardBlockRewardParams

type AwardStorageRewardParams struct {
	Miner  address.Address
	MinerSize int64
	NetSize int64
}

// Awards a reward to a block producer.
// This method is called only by the system actor, implicitly, as the last message in the evaluation of a block.
// The system actor thus computes the parameters and attached value.
//
// The reward includes two components:
// - the epoch block reward, computed and paid from the reward actor's balance,
// - the block gas reward, expected to be transferred to the reward actor with this invocation.
//
// The reward is reduced before the residual is credited to the block producer, by:
// - a penalty amount, provided as a parameter, which is burnt,
func (a Actor) AwardBlockReward(rt runtime.Runtime, params *AwardBlockRewardParams) *abi.EmptyValue {
	rt.ValidateImmediateCallerIs(builtin.SystemActorAddr)
	priorBalance := rt.CurrentBalance()
	if params.Penalty.LessThan(big.Zero()) {
		rt.Abortf(exitcode.ErrIllegalArgument, "negative penalty %v", params.Penalty)
	}
	if params.GasReward.LessThan(big.Zero()) {
		rt.Abortf(exitcode.ErrIllegalArgument, "negative gas reward %v", params.GasReward)
	}
	if priorBalance.LessThan(params.GasReward) {
		rt.Abortf(exitcode.ErrIllegalState, "actor current balance %v insufficient to pay gas reward %v",
			priorBalance, params.GasReward)
	}
	if params.WinCount <= 0 {
		rt.Abortf(exitcode.ErrIllegalArgument, "invalid win count %d", params.WinCount)
	}

	minerAddr, ok := rt.ResolveAddress(params.Miner)
	if !ok {
		rt.Abortf(exitcode.ErrNotFound, "failed to resolve given owner address")
	}
	// The miner penalty is scaled up by a factor of PenaltyMultiplier
	penalty := big.Mul(big.NewInt(PenaltyMultiplier), params.Penalty)
	totalReward := big.Zero()
	var st State
	rt.StateTransaction(&st, func() {
		blockReward := big.Mul(st.ThisEpochReward, big.NewInt(params.WinCount))
		blockReward = big.Div(blockReward, big.NewInt(builtin.ExpectedLeadersPerEpoch))
		totalReward = big.Add(blockReward, params.GasReward)
		currBalance := rt.CurrentBalance()
		fmt.Println("this epoch reward:", st.ThisEpochReward, "win count???", params.WinCount, "currBalance", currBalance)
		if totalReward.GreaterThan(currBalance) {
			rt.Log(rtt.WARN, "reward actor balance %d below totalReward expected %d, paying out rest of balance", currBalance, totalReward)
			totalReward = currBalance

			blockReward = big.Sub(totalReward, params.GasReward)
			// Since we have already asserted the balance is greater than gas reward blockReward is >= 0
			builtin.RequireState(rt, blockReward.GreaterThanEqual(big.Zero()), "programming error, block reward %v below zero", blockReward)
		}
		st.TotalStoragePowerReward = big.Add(st.TotalStoragePowerReward, blockReward)
	})

	builtin.RequireState(rt, totalReward.LessThanEqual(priorBalance), "reward %v exceeds balance %v", totalReward, priorBalance)

	fmt.Println("totalReward:", totalReward, "to???", minerAddr.String())
	fmt.Printf("reward state------------ %+v\n", st)

	// if this fails, we can assume the miner is responsible and avoid failing here.
	rewardParams := builtin.ApplyRewardParams{
		Reward:  totalReward,
		Penalty: penalty,
	}
	code := rt.Send(minerAddr, builtin.MethodsMiner.ApplyRewards, &rewardParams, totalReward, &builtin.Discard{})
	if !code.IsSuccess() {
		rt.Log(rtt.ERROR, "failed to send ApplyRewards call to the miner actor with funds: %v, code: %v", totalReward, code)
		code := rt.Send(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, totalReward, &builtin.Discard{})
		if !code.IsSuccess() {
			rt.Log(rtt.ERROR, "failed to send unsent reward to the burnt funds actor, code: %v", code)
		}
	}

	return nil
}

// Changed since v0:
// - removed ThisEpochReward (unsmoothed)
type ThisEpochRewardReturn struct {
	ThisEpochRewardSmoothed smoothing.FilterEstimate
	ThisEpochBaselinePower  abi.StoragePower
}

// The award value used for the current epoch, updated at the end of an epoch
// through cron tick.  In the case previous epochs were null blocks this
// is the reward value as calculated at the last non-null epoch.
func (a Actor) ThisEpochReward(rt runtime.Runtime, _ *abi.EmptyValue) *ThisEpochRewardReturn {
	rt.ValidateImmediateCallerAcceptAny()

	var st State
	rt.StateReadonly(&st)
	return &ThisEpochRewardReturn{
		ThisEpochRewardSmoothed: st.ThisEpochRewardSmoothed,
		ThisEpochBaselinePower:  st.ThisEpochBaselinePower,
	}
}

// Called at the end of each epoch by the power actor (in turn by its cron hook).
// This is only invoked for non-empty tipsets, but catches up any number of null
// epochs to compute the next epoch reward.
func (a Actor) UpdateNetworkKPI(rt runtime.Runtime, currRealizedPower *abi.StoragePower) *abi.EmptyValue {
	rt.ValidateImmediateCallerIs(builtin.StoragePowerActorAddr)
	if currRealizedPower == nil {
		rt.Abortf(exitcode.ErrIllegalArgument, "argument should not be nil")
	}

	var st State
	rt.StateTransaction(&st, func() {
		prev := st.Epoch
		// if there were null runs catch up the computation until
		// st.Epoch == rt.CurrEpoch()
		for st.Epoch < rt.CurrEpoch() {
			// Update to next epoch to process null rounds
			st.updateToNextEpoch(*currRealizedPower)
		}

		st.updateToNextEpochWithReward(*currRealizedPower)
		// only update smoothed estimates after updating reward and epoch
		st.updateSmoothedEstimates(st.Epoch - prev)
	})
	return nil
}

// Awards a reward to a block producer.
// This method is called only by the system actor, implicitly, as the last message in the evaluation of a block.
// The system actor thus computes the parameters and attached value.
//
// The reward includes two components:
// - the epoch block reward, computed and paid from the reward actor's balance,
// - the block gas reward, expected to be transferred to the reward actor with this invocation.
//
// The reward is reduced before the residual is credited to the block producer, by:
// - a penalty amount, provided as a parameter, which is burnt,
func (a Actor) StorageReward(rt runtime.Runtime, params *AwardStorageRewardParams) *abi.EmptyValue {

	// get miner list
	rt.ValidateImmediateCallerAcceptAny()
	minerAddr, ok := rt.ResolveAddress(params.Miner)
	if !ok {
		rt.Abortf(exitcode.ErrNotFound, "failed to resolve given owner address")
	}

	var st State
	rt.StateReadonly(&st)

	reward:= getKBReward(params.NetSize/1024) // reward per KB
	//fmt.Println("sssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss reward1", reward)
	reward = big.Mul(reward,big.NewInt(params.MinerSize/1024))
	fmt.Println("sssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss reward", reward)
	// The miner penalty is scaled up by a factor of PenaltyMultiplier
	// if this fails, we can assume the miner is responsible and avoid failing here.
	rewardParams := builtin.ApplyRewardParams{
		Reward:  reward,
		Penalty: big.Zero(),
	}

	code := rt.Send(minerAddr, builtin.MethodsMiner.ApplyRewards, &rewardParams, reward, &builtin.Discard{})
	//code := rt.Send(minerAddr, builtin.MethodSend, nil, rewardParams.Reward, &builtin.Discard{})
	if !code.IsSuccess() {
		rt.Log(rtt.ERROR, "failed to send ApplyRewards call to the miner actor with funds: %v, code: %v", params, code)
		code := rt.Send(builtin.BurntFundsActorAddr, builtin.MethodSend, nil, reward, &builtin.Discard{})
		if !code.IsSuccess() {
			rt.Log(rtt.ERROR, "failed to send unsent reward to the burnt funds actor, code: %v", code)
		}
	}

	return nil
}

// test function
func tGetKBReward(){
	fmt.Println(2 * 1024 * 1024 * 1024 * 1024)  // k m g t p ???k?????????????????? 2199023255552
	fmt.Println(6 * 1024 * 1024 * 1024 * 1024)  // k m g t p 6597069766656
	fmt.Println(12 * 1024 * 1024 * 1024 * 1024) // k m g t p 13194139533312
	fmt.Println(20 * 1024 * 1024 * 1024 * 1024) // k m g t p 21990232555520
	rw := getKBReward(1024)
	fmt.Println(rw.String())
	rw = getKBReward(2199023255553)
	fmt.Println(rw.String())
	rw = getKBReward(6597069766657)
	fmt.Println(rw.String())
	rw = getKBReward(13194139533313)
	fmt.Println(rw.String())
}
const FilecoinPrecision = 1_000_000_000_000_000_000

// ?????????tipset???kb????????????
// netSize???kb????????????????????????????????????????????????1024
func getKBReward(netSize int64) big.Int {
	var r512 float64
	switch {
	case netSize <= 2199023255552: // 0-2P
		//return big.NewInt(1616610)
		r512 = 0.002441 //?????? ?????????k ???512M???????????????
	case netSize <= 6597069766656:
		//return big.NewInt(970231)
		r512 = 0.001465 //?????? ?????????k ???512M???????????????
	case netSize <= 13194139533312:
		//return big.NewInt(647041)
		r512 = 0.000977 //?????? ?????????k ???512M???????????????
	default:
		//return big.NewInt(323189)
		r512 = 0.000488 //?????? ?????????k ???512M???????????????
	}
	r512 = r512 * FilecoinPrecision // ?????????????????????nanok
	rw := big.NewInt(int64(r512))
	rw = big.Div(rw, big.NewInt(512*1024)) // ????????????K?????????
	rw = big.Div(rw, big.NewInt(2880)) // ???????????????tipset??????????????????????????????tipset???kb?????????
	return  rw
}
