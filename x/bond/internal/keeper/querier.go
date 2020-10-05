//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"encoding/json"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/bond/internal/types"
)

// query endpoints supported by the bond Querier
const (
	ListBonds       = "list"
	GetBond         = "get"
	QueryByOwner    = "query-by-owner"
	QueryParameters = "parameters"
	Balance         = "balance"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case ListBonds:
			return listBonds(ctx, path[1:], req, keeper)
		case GetBond:
			return getBond(ctx, path[1:], req, keeper)
		case QueryByOwner:
			return queryBondsByOwner(ctx, path[1:], req, keeper)
		case QueryParameters:
			return queryParameters(ctx, path[1:], req, keeper)
		case Balance:
			return queryBalance(ctx, path[1:], req, keeper)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown bond query endpoint")
		}
	}
}

// nolint: unparam
func listBonds(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	bonds := keeper.ListBonds(ctx)

	bz, err2 := json.MarshalIndent(bonds, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func getBond(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {

	id := types.ID(strings.Join(path, "/"))
	if !keeper.HasBond(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Bond not found.")
	}

	bond := keeper.GetBond(ctx, id)

	bz, err2 := json.MarshalIndent(bond, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func queryBondsByOwner(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	bonds := keeper.QueryBondsByOwner(ctx, path[0])

	bz, err2 := json.MarshalIndent(bonds, "", "  ")
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
	balances := keeper.GetBondModuleBalances(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, balances)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}
