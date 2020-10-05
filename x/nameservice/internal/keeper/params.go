//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/wirelineio/dxns/x/nameservice/internal/types"
)

// Default parameter namespace.
const (
	DefaultParamspace = types.ModuleName
)

// ParamKeyTable - ParamTable for nameservice module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// RecordRent - get the record periodic rent.
func (k Keeper) RecordRent(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyRecordRent, &res)
	return
}

// RecordRentDuration - get the record expiry duration.
func (k Keeper) RecordRentDuration(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyRecordRentDuration, &res)
	return
}

func (k Keeper) AuthorityAuctionsEnabled(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyAuthorityAuctions, &res)
	return
}

func (k Keeper) AuthorityAuctionCommitsDuration(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyCommitsDuration, &res)
	return
}

func (k Keeper) AuthorityAuctionRevealsDuration(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyRevealsDuration, &res)
	return
}

func (k Keeper) AuthorityAuctionCommitFee(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyCommitFee, &res)
	return
}

func (k Keeper) AuthorityAuctionRevealFee(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyRevealFee, &res)
	return
}

func (k Keeper) AuthorityAuctionMinimumBid(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyMinimumBid, &res)
	return
}

func (k Keeper) AuthorityRent(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyAuthorityRent, &res)
	return
}

func (k Keeper) AuthorityRentDuration(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyAuthorityRentDuration, &res)
	return
}

func (k Keeper) AuthorityGracePeriod(ctx sdk.Context) (res time.Duration) {
	k.paramstore.Get(ctx, types.KeyAuthorityGracePeriod, &res)
	return
}

// GetParams - Get all parameteras as types.Params.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.RecordRent(ctx),
		k.RecordRentDuration(ctx),

		k.AuthorityRent(ctx),
		k.AuthorityRentDuration(ctx),
		k.AuthorityGracePeriod(ctx),

		k.AuthorityAuctionsEnabled(ctx),
		k.AuthorityAuctionCommitsDuration(ctx),
		k.AuthorityAuctionRevealsDuration(ctx),
		k.AuthorityAuctionCommitFee(ctx),
		k.AuthorityAuctionRevealFee(ctx),
		k.AuthorityAuctionMinimumBid(ctx),
	)
}

// SetParams - set the params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
