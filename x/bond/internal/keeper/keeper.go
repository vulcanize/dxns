//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/vulcanize/dxns/x/bond/internal/types"
)

// prefixIDToBondIndex is the prefix for ID -> Bond index in the KVStore.
// Note: This is the primary index in the system.
// Note: Golang doesn't support const arrays.
var prefixIDToBondIndex = []byte{0x00}

// prefixOwnerToBondsIndex is the prefix for the Owner -> [Bond] index in the KVStore.
var prefixOwnerToBondsIndex = []byte{0x01}

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper

	// Track bond usage in other cosmos-sdk modules (more like a usage tracker).
	usageKeepers []types.BondUsageKeeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.

	paramSubspace subspace.Subspace
}

// BondClientKeeper is the subset of functionality exposed to other modules.
type BondClientKeeper interface {
	HasBond(ctx sdk.Context, id types.ID) bool
	GetBond(ctx sdk.Context, id types.ID) types.Bond
	MatchBonds(ctx sdk.Context, matchFn func(*types.Bond) bool) []*types.Bond
	TransferCoinsToModuleAccount(ctx sdk.Context, id types.ID, moduleAccount string, coins sdk.Coins) error
	TranserCoinsToAccount(ctx sdk.Context, id types.ID, account sdk.AccAddress, coins sdk.Coins) error
}

var _ BondClientKeeper = (*Keeper)(nil)

// NewKeeper creates new instances of the bond Keeper
func NewKeeper(accountKeeper auth.AccountKeeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper,
	usageKeepers []types.BondUsageKeeper, storeKey sdk.StoreKey, cdc *codec.Codec, paramstore params.Subspace) Keeper {
	return Keeper{
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		supplyKeeper:  supplyKeeper,
		usageKeepers:  usageKeepers,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}

// Generates Bond ID -> Bond index key.
func getBondIndexKey(id types.ID) []byte {
	return append(prefixIDToBondIndex, []byte(id)...)
}

// Generates Owner -> Bonds index key.
func getOwnerToBondsIndexKey(owner string, bondID types.ID) []byte {
	return append(append(prefixOwnerToBondsIndex, []byte(owner)...), []byte(bondID)...)
}

// SaveBond - saves a bond to the store.
func (k Keeper) SaveBond(ctx sdk.Context, bond types.Bond) {
	store := ctx.KVStore(k.storeKey)

	// Bond ID -> Bond index.
	store.Set(getBondIndexKey(bond.ID), k.cdc.MustMarshalBinaryBare(bond))

	// Owner -> [Bond] index.
	store.Set(getOwnerToBondsIndexKey(bond.Owner, bond.ID), []byte{})
}

// HasBond - checks if a bond by the given ID exists.
func (k Keeper) HasBond(ctx sdk.Context, id types.ID) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(getBondIndexKey(id))
}

// DeleteBond - deletes the bond.
func (k Keeper) DeleteBond(ctx sdk.Context, bond types.Bond) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getBondIndexKey(bond.ID))
	store.Delete(getOwnerToBondsIndexKey(bond.Owner, bond.ID))
}

// GetBond - gets a record from the store.
func (k Keeper) GetBond(ctx sdk.Context, id types.ID) types.Bond {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(getBondIndexKey(id))
	var obj types.Bond
	k.cdc.MustUnmarshalBinaryBare(bz, &obj)

	return obj
}

// ListBonds - get all bonds.
func (k Keeper) ListBonds(ctx sdk.Context) []types.Bond {
	var bonds []types.Bond

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, prefixIDToBondIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.Bond
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			bonds = append(bonds, obj)
		}
	}

	return bonds
}

// QueryBondsByOwner - query bonds by owner.
func (k Keeper) QueryBondsByOwner(ctx sdk.Context, ownerAddress string) []types.Bond {
	var bonds []types.Bond

	ownerPrefix := append(prefixOwnerToBondsIndex, []byte(ownerAddress)...)
	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, ownerPrefix)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bondID := itr.Key()[len(ownerPrefix):]
		bz := store.Get(append(prefixIDToBondIndex, bondID...))
		if bz != nil {
			var obj types.Bond
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			bonds = append(bonds, obj)
		}
	}

	return bonds
}

// MatchBonds - get all matching bonds.
func (k Keeper) MatchBonds(ctx sdk.Context, matchFn func(*types.Bond) bool) []*types.Bond {
	var bonds []*types.Bond

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, prefixIDToBondIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.Bond
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			if matchFn(&obj) {
				bonds = append(bonds, &obj)
			}
		}
	}

	return bonds
}

// CreateBond creates a new bond.
func (k Keeper) CreateBond(ctx sdk.Context, ownerAddress sdk.AccAddress, coins sdk.Coins) (*types.Bond, error) {
	// Check if account has funds.
	if !k.bankKeeper.HasCoins(ctx, ownerAddress, coins) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient funds.")
	}

	// Generate bond ID.
	account := k.accountKeeper.GetAccount(ctx, ownerAddress)
	bondID := types.BondID{
		Address:  ownerAddress,
		AccNum:   account.GetAccountNumber(),
		Sequence: account.GetSequence(),
	}.Generate()

	maxBondAmount, err := k.getMaxBondAmount(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid max bond amount.")
	}

	bond := types.Bond{ID: types.ID(bondID), Owner: ownerAddress.String(), Balance: coins}
	if bond.Balance.IsAnyGT(maxBondAmount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Max bond amount exceeded.")
	}

	// Move funds into the bond account module.
	sdkErr := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, ownerAddress, types.ModuleName, bond.Balance)
	if err != nil {
		return nil, sdkErr
	}

	// Save bond in store.
	k.SaveBond(ctx, bond)

	return &bond, nil
}

// RefillBond refills an existing bond.
func (k Keeper) RefillBond(ctx sdk.Context, id types.ID, ownerAddress sdk.AccAddress, coins sdk.Coins) (*types.Bond, error) {
	if !k.HasBond(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	bond := k.GetBond(ctx, id)
	if bond.Owner != ownerAddress.String() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	// Check if account has funds.
	if !k.bankKeeper.HasCoins(ctx, ownerAddress, coins) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient funds.")
	}

	maxBondAmount, err := k.getMaxBondAmount(ctx)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid max bond amount.")
	}

	updatedBalance := bond.Balance.Add(coins...)
	if updatedBalance.IsAnyGT(maxBondAmount) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Max bond amount exceeded.")
	}

	// Move funds into the bond account module.
	sdkErr := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, ownerAddress, types.ModuleName, coins)
	if err != nil {
		return nil, sdkErr
	}

	// Update bond balance and save.
	bond.Balance = updatedBalance
	k.SaveBond(ctx, bond)

	return &bond, nil
}

// WithdrawBond withdraws funds from a bond.
func (k Keeper) WithdrawBond(ctx sdk.Context, id types.ID, ownerAddress sdk.AccAddress, coins sdk.Coins) (*types.Bond, error) {
	if !k.HasBond(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	bond := k.GetBond(ctx, id)
	if bond.Owner != ownerAddress.String() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	updatedBalance, isNeg := bond.Balance.SafeSub(coins)
	if isNeg {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient bond balance.")
	}

	// Move funds from the bond into the account.
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, ownerAddress, coins)
	if err != nil {
		return nil, err
	}

	// Update bond balance and save.
	bond.Balance = updatedBalance
	k.SaveBond(ctx, bond)

	return &bond, nil
}

// CancelBond cancels a bond, returning funds to the owner.
func (k Keeper) CancelBond(ctx sdk.Context, id types.ID, ownerAddress sdk.AccAddress) (*types.Bond, error) {
	if !k.HasBond(ctx, id) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	bond := k.GetBond(ctx, id)
	if bond.Owner != ownerAddress.String() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	// Check if bond is used in other modules.
	for _, usageKeeper := range k.usageKeepers {
		if usageKeeper.UsesBond(ctx, id) {
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("Bond in use by the '%s' module.", usageKeeper.ModuleName()))
		}
	}

	// Move funds from the bond into the account.
	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, ownerAddress, bond.Balance)
	if err != nil {
		return nil, err
	}

	k.DeleteBond(ctx, bond)

	return &bond, nil
}

// GetBondModuleBalances gets the bond module account(s) balances.
func (k Keeper) GetBondModuleBalances(ctx sdk.Context) map[string]sdk.Coins {
	balances := map[string]sdk.Coins{}
	accountNames := []string{types.ModuleName}

	for _, accountName := range accountNames {
		moduleAddress := k.supplyKeeper.GetModuleAddress(accountName)
		moduleAccount := k.accountKeeper.GetAccount(ctx, moduleAddress)
		if moduleAccount != nil {
			balances[accountName] = moduleAccount.GetCoins()
		}
	}

	return balances
}

// TransferCoinsToModuleAccount noves funds from the bonds module account to another module account.
func (k Keeper) TransferCoinsToModuleAccount(ctx sdk.Context, id types.ID, moduleAccount string, coins sdk.Coins) error {
	if !k.HasBond(ctx, id) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond not found.")
	}

	bondObj := k.GetBond(ctx, id)

	// Deduct rent from bond.
	updatedBalance, isNeg := bondObj.Balance.SafeSub(coins)
	if isNeg {
		// Check if bond has sufficient funds.
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "Insufficient funds.")
	}

	// Move funds from bond module to record rent module.
	err := k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, moduleAccount, coins)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Error transfering funds.")
	}

	// Update bond balance.
	bondObj.Balance = updatedBalance
	k.SaveBond(ctx, bondObj)

	return nil
}

// TranserCoinsToAccount moves coins from the bond to an account.
func (k Keeper) TranserCoinsToAccount(ctx sdk.Context, id types.ID, account sdk.AccAddress, coins sdk.Coins) error {
	return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Not implemented.")
}

func (k Keeper) getMaxBondAmount(ctx sdk.Context) (sdk.Coins, error) {
	params := k.GetParams(ctx)
	maxBondAmount, err := sdk.ParseCoins(params.MaxBondAmount)
	if err != nil {
		return nil, err
	}

	return maxBondAmount, nil
}
