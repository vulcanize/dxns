//
// Copyright 2020 Wireline, Inc.
//

package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wirelineio/dxns/x/auction/internal/types"
)

// GetQueryCmd returns query commands.
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	auctionQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	auctionQueryCmd.AddCommand(flags.GetCommands(
		GetCmdList(storeKey, cdc),
		GetCmdGetAuction(storeKey, cdc),
		GetCmdGetBid(storeKey, cdc),
		GetCmdGetBids(storeKey, cdc),
		GetCmdListByBidder(storeKey, cdc),
		GetCmdQueryParams(storeKey, cdc),
		GetCmdBalance(storeKey, cdc),
	)...)
	return auctionQueryCmd
}

// GetCmdList queries all auctions.
func GetCmdList(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List auctions.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/list", queryRoute), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

// GetCmdGetBid queries an auction bid.
func GetCmdGetBid(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-bid [auction-id] [bidder]",
		Short: "Get auction bid.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := args[0]
			bidder := args[1]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get-bid/%s/%s", queryRoute, id, bidder), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

// GetCmdGetBids queries all auction bids.
func GetCmdGetBids(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-bids [auction-id]",
		Short: "Get all auction bids.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get-bids/%s", queryRoute, id), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

// GetCmdGetAuction queries an auction.
func GetCmdGetAuction(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get [ID]",
		Short: "Get auction.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get/%s", queryRoute, id), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

// GetCmdListByBidder queries auctions by bidder.
func GetCmdListByBidder(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "query-by-owner [address]",
		Short: "Query auctions by owner/creator.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			address := args[0]

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/query-by-owner/%s", queryRoute, address), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current auction parameters information.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as auction parameters.

Example:
$ %s query auction params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/parameters", queryRoute)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			cdc.MustUnmarshalJSON(bz, &params)
			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCmdBalance queries the auction module account balance.
func GetCmdBalance(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Get auction module account balance.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set("trust-node", true)

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balance", queryRoute), nil)
			if err != nil {
				return err
			}

			fmt.Println(string(res))

			return nil
		},
	}
}
