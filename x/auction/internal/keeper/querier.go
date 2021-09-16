//
// Copyright 2020 Wireline, Inc.
//

package keeper

import (
	"encoding/json"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/vulcanize/dxns/x/auction/internal/types"
)

// query endpoints supported by the auction Querier
const (
	QueryListAuctions = "list"
	QueryGetAuction   = "get"
	QueryGetBid       = "get-bid"
	QueryGetBids      = "get-bids"
	QueryByOwner      = "query-by-owner"
	QueryParameters   = "parameters"
	QueryBalance      = "balance"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryListAuctions:
			return listAuctions(ctx, path[1:], req, keeper)
		case QueryGetAuction:
			return getAuction(ctx, path[1:], req, keeper)
		case QueryGetBid:
			return getBid(ctx, path[1:], req, keeper)
		case QueryGetBids:
			return getBids(ctx, path[1:], req, keeper)
		case QueryByOwner:
			return queryAuctionsByOwner(ctx, path[1:], req, keeper)
		case QueryParameters:
			return queryParameters(ctx, path[1:], req, keeper)
		case QueryBalance:
			return queryBalance(ctx, path[1:], req, keeper)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown auction query endpoint")
		}
	}
}

// nolint: unparam
func listAuctions(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	auctions := keeper.ListAuctions(ctx)

	bz, err2 := json.MarshalIndent(auctions, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func getAuction(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {

	id := types.ID(strings.Join(path, "/"))
	if !keeper.HasAuction(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Auction not found.")
	}

	auction := keeper.GetAuction(ctx, id)

	bz, err2 := json.MarshalIndent(auction, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func getBid(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	id := types.ID(path[0])
	bidder := path[1]

	if !keeper.HasBid(ctx, types.ID(id), bidder) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Bid not found.")
	}

	bid := keeper.GetBid(ctx, types.ID(id), bidder)

	bz, err2 := json.MarshalIndent(bid, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func getBids(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	id := types.ID(path[0])

	bids := keeper.GetBids(ctx, types.ID(id))

	bz, err2 := json.MarshalIndent(bids, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func queryAuctionsByOwner(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	auctions := keeper.QueryAuctionsByOwner(ctx, path[0])

	bz, err2 := json.MarshalIndent(auctions, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

func queryParameters(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	params := keeper.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}

func queryBalance(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	balances := keeper.GetAuctionModuleBalances(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, balances)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}
