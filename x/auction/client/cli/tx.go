//
// Copyright 2020 Wireline, Inc.
//

package cli

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	wnsUtils "github.com/wirelineio/dxns/utils"
	"github.com/wirelineio/dxns/x/auction/internal/types"
)

// GetTxCmd returns transaction commands for this module.
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// TODO(ashwin): Add Tx commands.
	auctionTxCmd.AddCommand(flags.PostCommands(
		GetCmdCommitBid(cdc),
		GetCmdRevealBid(cdc),
	)...)

	return auctionTxCmd
}

// GetCmdCommitBid is the CLI command for committing a bid.
func GetCmdCommitBid(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-bid [auction-id] [bid-amount]",
		Short: "Commit sealed bid.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Validate bid amount.
			bidAmount, err := sdk.ParseCoin(args[1])
			if err != nil {
				return err
			}

			mnemonic, err := wnsUtils.GenerateMnemonic()
			if err != nil {
				return err
			}

			chainID := viper.GetString("chain-id")
			auctionID := args[0]

			reveal := map[string]interface{}{
				"chainId":       chainID,
				"auctionId":     auctionID,
				"bidderAddress": cliCtx.GetFromAddress().String(),
				"bidAmount":     bidAmount.String(),
				"noise":         mnemonic,
			}

			commitHash, content, err := wnsUtils.GenerateHash(reveal)
			if err != nil {
				return err
			}

			// Save reveal file.
			ioutil.WriteFile(fmt.Sprintf("%s-%s.json", cliCtx.GetFromName(), commitHash), content, 0600)

			msg := types.NewMsgCommitBid(auctionID, commitHash, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdRevealBid is the CLI command for revealing a bid.
func GetCmdRevealBid(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal-bid [auction-id] [reveal-file-path]",
		Short: "Reveal bid.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			auctionID := args[0]
			revealFilePath := args[1]

			revealBytes, err := ioutil.ReadFile(revealFilePath)
			if err != nil {
				return err
			}

			// TODO(ashwin): Before revealing, check if auction is in reveal phase.

			msg := types.NewMsgRevealBid(auctionID, hex.EncodeToString(revealBytes), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
