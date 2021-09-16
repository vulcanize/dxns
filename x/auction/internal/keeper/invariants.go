//
// Copyright 2020 Wireline, Inc.
//

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/vulcanize/dxns/x/auction/internal/types"
)

// RegisterInvariants registers all auction module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-account", ModuleAccountInvariant(k))
}

// ModuleAccountInvariant checks that the 'auction' module account balance is non-negative.
func ModuleAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		moduleAccount := k.supplyKeeper.GetModuleAccount(ctx, types.ModuleName)
		if moduleAccount.GetCoins().IsAnyNegative() {
			return sdk.FormatInvariant(
					types.ModuleName,
					"module-account",
					fmt.Sprintf("Module account '%s' has negative balance.", types.ModuleName)),
				true
		}

		return "", false
	}
}

// AllInvariants runs all invariants of the auctions module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		return ModuleAccountInvariant(k)(ctx)
	}
}
