//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"encoding/json"
	"strings"

	"github.com/vulcanize/dxns/x/bond"
	"github.com/vulcanize/dxns/x/nameservice/internal/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the nameservice Querier
const (
	ListRecordsPath        = "list"
	GetRecordPath          = "get"
	QueryRecordsByBondPath = "query-by-bond"
	QueryParametersPath    = "parameters"
	Balance                = "balance"

	WhoIsPath       = "whois"
	LookUpWRNPath   = "lookup"
	ListNamesPath   = "names"
	ResolveNamePath = "resolve"

	RecordExpiryQueue    = "record-expiry"
	AuthorityExpiryQueue = "authority-expiry"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case WhoIsPath:
			return whoIs(ctx, path[1:], req, keeper)
		case LookUpWRNPath:
			return lookupWRN(ctx, path[1:], req, keeper)
		case ListRecordsPath:
			return listResources(ctx, path[1:], req, keeper)
		case GetRecordPath:
			return getResource(ctx, path[1:], req, keeper)
		case ListNamesPath:
			return listNames(ctx, path[1:], req, keeper)
		case ResolveNamePath:
			return resolveName(ctx, path[1:], req, keeper)
		case QueryRecordsByBondPath:
			return queryRecordsByBond(ctx, path[1:], req, keeper)
		case QueryParametersPath:
			return queryParameters(ctx, path[1:], req, keeper)
		case Balance:
			return queryBalance(ctx, path[1:], req, keeper)
		case RecordExpiryQueue:
			return queryRecordExpiryQueue(ctx, path[1:], req, keeper)
		case AuthorityExpiryQueue:
			return queryAuthorityExpiryQueue(ctx, path[1:], req, keeper)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown nameservice query endpoint")
		}
	}
}

func whoIs(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	name := path[0]

	if !keeper.HasNameAuthority(ctx, name) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Name not found.")
	}

	nameAuthority := keeper.GetNameAuthority(ctx, name)

	bz, err2 := json.MarshalIndent(nameAuthority, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

func lookupWRN(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	wrn := strings.Join(path, "/")

	if !keeper.HasNameRecord(ctx, wrn) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "WRN not found.")
	}

	nameRecord := keeper.GetNameRecord(ctx, wrn)

	bz, err2 := json.MarshalIndent(nameRecord, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func listResources(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	records := keeper.ListRecords(ctx)

	bz, err2 := json.MarshalIndent(records, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func getResource(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {

	id := types.ID(strings.Join(path, "/"))
	if !keeper.HasRecord(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Record not found.")
	}

	record := keeper.GetRecord(ctx, id)

	bz, err2 := json.MarshalIndent(record, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func listNames(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	records := keeper.ListNameRecords(ctx)

	bz, err2 := json.MarshalIndent(records, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func resolveName(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	wrn := strings.Join(path, "/")

	record := keeper.ResolveWRN(ctx, wrn)

	bz, err2 := json.MarshalIndent(record, "", "  ")
	if err2 != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return bz, nil
}

// nolint: unparam
func queryRecordsByBond(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {

	id := bond.ID(strings.Join(path, "/"))
	records := keeper.recordKeeper.QueryRecordsByBond(ctx, id)

	bz, err2 := json.MarshalIndent(records, "", "  ")
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
	balances := keeper.GetModuleBalances(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, balances)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}

func queryRecordExpiryQueue(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	expiryQueue := keeper.GetRecordExpiryQueue(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, expiryQueue)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}

func queryAuthorityExpiryQueue(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	expiryQueue := keeper.GetAuthorityExpiryQueue(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, expiryQueue)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "could not marshal result to JSON")
	}

	return res, nil
}
