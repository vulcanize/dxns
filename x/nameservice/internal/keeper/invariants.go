//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/vulcanize/dxns/x/nameservice/internal/types"
)

// RegisterInvariants registers all nameservice module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "record", RecordInvariants(k))
}

// ModuleAccountInvariant checks that the 'bond' module account balance is non-negative.
func ModuleAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		moduleAccount := k.supplyKeeper.GetModuleAccount(ctx, types.RecordRentModuleAccountName)
		if moduleAccount.GetCoins().IsAnyNegative() {
			return sdk.FormatInvariant(
					types.ModuleName,
					"module-account",
					fmt.Sprintf("Module account '%s' has negative balance.", types.RecordRentModuleAccountName)),
				true
		}

		return "", false
	}
}

// RecordInvariants checks that every record:
// (1) has a corresponding naming record &
// (2) associated bond exists, if bondID is not null.
func RecordInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		itr := sdk.KVStorePrefixIterator(store, PrefixCIDToRecordIndex)
		defer itr.Close()
		for ; itr.Valid(); itr.Next() {
			bz := store.Get(itr.Key())
			if bz != nil {
				var obj types.RecordObj
				k.cdc.MustUnmarshalBinaryBare(bz, &obj)
				record := obj.ToRecord()

				if record.BondID != "" && !k.bondKeeper.HasBond(ctx, record.BondID) {
					return sdk.FormatInvariant(types.ModuleName, "record-bond", fmt.Sprintf("Bond not found for record ID: '%s'.", record.ID)), true
				}
			}
		}

		return "", false
	}
}

// AllInvariants runs all invariants of the nameservice module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := RecordInvariants(k)(ctx)
		if stop {
			return res, stop
		}

		return ModuleAccountInvariant(k)(ctx)
	}
}
