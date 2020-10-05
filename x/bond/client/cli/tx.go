//
// Copyright 2019 Wireline, Inc.
//

package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/wirelineio/dxns/x/bond/internal/types"
)

// GetTxCmd returns transaction commands for this module.
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	bondTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	bondTxCmd.AddCommand(flags.PostCommands(
		GetCmdCreateBond(cdc),
		GetCmdRefillBond(cdc),
		GetCmdWithdrawFromBond(cdc),
		GetCmdCancelBond(cdc),
	)...)

	return bondTxCmd
}

// GetCmdCreateBond is the CLI command for creating a bond.
func GetCmdCreateBond(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [amount]",
		Short: "Create bond.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			coin, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateBond(sdk.NewCoins(coin), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdRefillBond is the CLI command for creating a bond.
func GetCmdRefillBond(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refill [bond ID] [amount]",
		Short: "Refill bond.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			bondID := args[0]
			coin, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgRefillBond(bondID, coin.Denom, coin.Amount.Int64(), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdCancelBond is the CLI command for cancelling a bond.
func GetCmdCancelBond(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [bond ID]",
		Short: "Cancel bond.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			msg := types.NewMsgCancelBond(args[0], cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdWithdrawFromBond is the CLI command for withdrawing funds from a bond.
func GetCmdWithdrawFromBond(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [bond ID] [amount]",
		Short: "Withdraw funds from bond.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			bondID := args[0]
			coin, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawBond(bondID, coin.Denom, coin.Amount.Int64(), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
