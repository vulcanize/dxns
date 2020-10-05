//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/wirelineio/dxns/x/bond/internal/types"
)

// Default parameter namespace.
const (
	DefaultParamspace = types.ModuleName
)

// ParamKeyTable - ParamTable for bond module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// MaxBondAmount - get the max bond amount.
func (k Keeper) MaxBondAmount(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyMaxBondAmount, &res)
	return
}

// GetParams - Get all parameteras as types.Params.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxBondAmount(ctx),
	)
}

// SetParams - set the params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
