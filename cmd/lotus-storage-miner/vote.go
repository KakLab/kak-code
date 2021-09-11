package main

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	cliutil "github.com/filecoin-project/lotus/cli/util"
	miner2 "github.com/filecoin-project/specs-actors/v2/actors/builtin/miner"
	miner4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/miner"
	cbor "github.com/ipfs/go-ipld-cbor"
	"golang.org/x/xerrors"
	"strconv"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/specs-actors/v4/actors/builtin"
	"github.com/urfave/cli/v2"
)

var GetFullNodeAPI = cliutil.GetFullNodeAPI
var ReqContext = cliutil.ReqContext

var voteCmd = &cli.Command{
	Name:  "vote",
	Usage: "interact with vote",
	Subcommands: []*cli.Command{
		voteStatusCmd,
		voteSendCmd,
		voteWithdrawCmd,
	},
}

var voteSendCmd = &cli.Command{
	Name:      "send",
	Usage:     "add vote",
	ArgsUsage: "amount (KAK)",
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := ReqContext(cctx)

		nodeApi, mcloser, err := lcli.GetStorageMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer mcloser()

		// get argument
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify amount of KAK to vote")
		}
		amount, err := strconv.ParseUint(cctx.Args().First(), 10, 64)
		if err != nil {
			return err
		}

		fromAddr, err := api.WalletDefaultAddress(ctx)
		if err != nil {
			return err
		}

		maddr, err := nodeApi.ActorAddress(ctx)
		if err != nil {
			return err
		}

		// build msg
		params := miner4.AddPosParams{
			Pos: big.Mul(types.NewInt(amount), big.NewInt(int64(build.FilecoinPrecision))),
		}
		enc, err := actors.SerializeParams(&params)
		if err != nil {
			return err
		}
		msg := &types.Message{
			To:     maddr,    // miner
			From:   fromAddr, // wallet
			Value:  params.Pos,
			Method: builtin.MethodsMiner.AddPos,
			Params: enc,
		}

		// send msg
		smsg, err := api.MpoolPushMessage(ctx, msg, nil)
		if err != nil {
			return err
		}
		fmt.Println(smsg.Cid())

		return nil
	},
}

var voteWithdrawCmd = &cli.Command{
	Name:      "withdraw",
	Usage:     "withdraw vote",
	ArgsUsage: "amount (KAK)",
	Action: func(cctx *cli.Context) error {
		api, closer, err := GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := ReqContext(cctx)

		nodeApi, mcloser, err := lcli.GetStorageMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer mcloser()

		maddr, err := nodeApi.ActorAddress(ctx)
		if err != nil {
			return err
		}

		mi, err := api.StateMinerInfo(ctx, maddr, types.EmptyTSK)
		if err != nil {
			return err
		}

		mact, err := api.StateGetActor(ctx, maddr, types.EmptyTSK)
		if err != nil {
			return err
		}

		tbs := blockstore.NewTieredBstore(blockstore.NewAPIBlockstore(api), blockstore.NewMemory())
		mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mact)
		if err != nil {
			return err
		}

		lockedFunds, err := mas.LockedFunds()
		if err != nil {
			return xerrors.Errorf("getting locked funds: %w", err)
		}
		available:= lockedFunds.PosDeposits
		amount := available
		if cctx.Args().Present() {
			f, err := types.ParseFIL(cctx.Args().First())
			if err != nil {
				return xerrors.Errorf("parsing 'amount' argument: %w", err)
			}

			amount = abi.TokenAmount(f)

			if amount.GreaterThan(available) {
				return xerrors.Errorf("can't withdraw more funds than available; requested: %s; available: %s", amount, available)
			}
		}

		// build msg
		params, err := actors.SerializeParams(&miner2.WithdrawBalanceParams{
			AmountRequested: amount, // Default to attempting to withdraw all the extra funds in the miner actor
		})
		if err != nil {
			return err
		}
		msg := &types.Message{
			To:     maddr,    // miner
			From:   mi.Owner, // owner
			Value:  types.NewInt(0),
			Method: builtin.MethodsMiner.WithDrawPos,
			Params: params,
		}

		// send msg
		smsg, err := api.MpoolPushMessage(ctx, msg, nil)
		if err != nil {
			return err
		}
		fmt.Println(smsg.Cid())

		return nil
	},
}

var voteStatusCmd = &cli.Command{
	Name:      "status",
	Usage:     "Get the vote status of kak system",
	ArgsUsage: "<sectorNum>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "log",
			Usage: "display event log",
		},
		&cli.BoolFlag{
			Name:  "on-chain-info",
			Usage: "show sector on chain info",
		},
	},
	Action: func(cctx *cli.Context) error {
		nodeApi, closer, err := lcli.GetStorageMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		if !cctx.Args().Present() {
			return fmt.Errorf("must specify sector number to get status of")
		}

		id, err := strconv.ParseUint(cctx.Args().First(), 10, 64)
		if err != nil {
			return err
		}

		onChainInfo := cctx.Bool("on-chain-info")
		status, err := nodeApi.SectorsStatus(ctx, abi.SectorNumber(id), onChainInfo)
		if err != nil {
			return err
		}

		fmt.Printf("SectorID:\t%d\n", status.SectorID)
		fmt.Printf("Status:\t\t%s\n", status.State)
		fmt.Printf("CIDcommD:\t%s\n", status.CommD)
		fmt.Printf("CIDcommR:\t%s\n", status.CommR)
		fmt.Printf("Ticket:\t\t%x\n", status.Ticket.Value)
		fmt.Printf("TicketH:\t%d\n", status.Ticket.Epoch)
		fmt.Printf("Seed:\t\t%x\n", status.Seed.Value)
		fmt.Printf("SeedH:\t\t%d\n", status.Seed.Epoch)
		fmt.Printf("Precommit:\t%s\n", status.PreCommitMsg)
		fmt.Printf("Commit:\t\t%s\n", status.CommitMsg)
		fmt.Printf("Proof:\t\t%x\n", status.Proof)
		fmt.Printf("Deals:\t\t%v\n", status.Deals)
		fmt.Printf("Retries:\t%d\n", status.Retries)
		if status.LastErr != "" {
			fmt.Printf("Last Error:\t\t%s\n", status.LastErr)
		}

		if onChainInfo {
			fmt.Printf("\nSector On Chain Info\n")
			fmt.Printf("SealProof:\t\t%x\n", status.SealProof)
			fmt.Printf("Activation:\t\t%v\n", status.Activation)
			fmt.Printf("Expiration:\t\t%v\n", status.Expiration)
			fmt.Printf("DealWeight:\t\t%v\n", status.DealWeight)
			fmt.Printf("VerifiedDealWeight:\t\t%v\n", status.VerifiedDealWeight)
			fmt.Printf("InitialPledge:\t\t%v\n", status.InitialPledge)
			fmt.Printf("\nExpiration Info\n")
			fmt.Printf("OnTime:\t\t%v\n", status.OnTime)
			fmt.Printf("Early:\t\t%v\n", status.Early)
		}

		if cctx.Bool("log") {
			fmt.Printf("--------\nEvent Log:\n")

			for i, l := range status.Log {
				fmt.Printf("%d.\t%s:\t[%s]\t%s\n", i, time.Unix(int64(l.Timestamp), 0), l.Kind, l.Message)
				if l.Trace != "" {
					fmt.Printf("\t%s\n", l.Trace)
				}
			}
		}
		return nil
	},
}
