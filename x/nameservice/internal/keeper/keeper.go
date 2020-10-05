//
// Copyright 2019 Wireline, Inc.
//

package keeper

import (
	"bytes"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/go-amino"
	"github.com/wirelineio/dxns/x/auction"
	"github.com/wirelineio/dxns/x/bond"
	"github.com/wirelineio/dxns/x/nameservice/internal/types"
)

// PrefixCIDToRecordIndex is the prefix for CID -> Record index.
// Note: This is the primary index in the system.
// Note: Golang doesn't support const arrays.
var PrefixCIDToRecordIndex = []byte{0x00}

// PrefixNameAuthorityRecordIndex is the prefix for the name -> NameAuthority index.
var PrefixNameAuthorityRecordIndex = []byte{0x01}

// PrefixWRNToNameRecordIndex is the prefix for the WRN -> NamingRecord index.
var PrefixWRNToNameRecordIndex = []byte{0x02}

// PrefixBondIDToRecordsIndex is the prefix for the Bond ID -> [Record] index.
var PrefixBondIDToRecordsIndex = []byte{0x03}

// PrefixBlockChangesetIndex is the prefix for the block changeset index.
var PrefixBlockChangesetIndex = []byte{0x04}

// PrefixAuctionToAuthorityNameIndex is the prefix for the auction ID -> authority name index.
var PrefixAuctionToAuthorityNameIndex = []byte{0x05}

// PrefixBondIDToAuthoritiesIndex is the prefix for the Bond ID -> [Authority] index.
var PrefixBondIDToAuthoritiesIndex = []byte{0x06}

// PrefixExpiryTimeToRecordsIndex is the prefix for the Expiry Time -> [Record] index.
var PrefixExpiryTimeToRecordsIndex = []byte{0x10}

// PrefixExpiryTimeToAuthoritiesIndex is the prefix for the Expiry Time -> [Authority] index.
var PrefixExpiryTimeToAuthoritiesIndex = []byte{0x11}

// KeySyncStatus is the key for the sync status record.
// Only used by WNS lite but defined here to prevent conflicts with existing prefixes.
var KeySyncStatus = []byte{0xff}

// PrefixCIDToNamesIndex the the reverse index for naming, i.e. maps CID -> []Names.
// TODO(ashwin): Move out of WNS once we have an indexing service.
var PrefixCIDToNamesIndex = []byte{0xe0}

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	accountKeeper auth.AccountKeeper
	supplyKeeper  supply.Keeper
	recordKeeper  RecordKeeper
	bondKeeper    bond.BondClientKeeper
	auctionKeeper auction.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.

	paramstore params.Subspace
}

// NewKeeper creates new instances of the nameservice Keeper
func NewKeeper(accountKeeper auth.AccountKeeper, supplyKeeper supply.Keeper, recordKeeper RecordKeeper, bondKeeper bond.BondClientKeeper, auctionKeeper auction.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec, paramstore params.Subspace) Keeper {
	return Keeper{
		accountKeeper: accountKeeper,
		supplyKeeper:  supplyKeeper,
		recordKeeper:  recordKeeper,
		bondKeeper:    bondKeeper,
		auctionKeeper: auctionKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramstore:    paramstore.WithKeyTable(ParamKeyTable()),
	}
}

// RecordKeeper exposes the bare minimal read-only API for other modules.
type RecordKeeper struct {
	auctionKeeper auction.Keeper
	storeKey      sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc           *codec.Codec // The wire codec for binary encoding/decoding.
}

// Record keeper implements the bond usage keeper interface.
var _ bond.BondUsageKeeper = (*RecordKeeper)(nil)
var _ auction.AuctionUsageKeeper = (*RecordKeeper)(nil)

// NewRecordKeeper creates new instances of the nameservice RecordKeeper
func NewRecordKeeper(auctionKeeper auction.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) RecordKeeper {
	return RecordKeeper{
		auctionKeeper: auctionKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
	}
}

// PutRecord - saves a record to the store and updates ID -> Record index.
func (k Keeper) PutRecord(ctx sdk.Context, record types.Record) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetRecordIndexKey(record.ID), k.cdc.MustMarshalBinaryBare(record.ToRecordObj()))
	k.updateBlockChangesetForRecord(ctx, record.ID)
}

// Generates Bond ID -> Bond index key.
func GetRecordIndexKey(id types.ID) []byte {
	return append(PrefixCIDToRecordIndex, []byte(id)...)
}

// Generates Bond ID -> Records index key.
func getBondIDToRecordsIndexKey(bondID bond.ID, id types.ID) []byte {
	return append(append(PrefixBondIDToRecordsIndex, []byte(bondID)...), []byte(id)...)
}

// AddBondToRecordIndexEntry adds the Bond ID -> [Record] index entry.
func (k Keeper) AddBondToRecordIndexEntry(ctx sdk.Context, bondID bond.ID, id types.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Set(getBondIDToRecordsIndexKey(bondID, id), []byte{})
}

// RemoveBondToRecordIndexEntry removes the Bond ID -> [Record] index entry.
func (k Keeper) RemoveBondToRecordIndexEntry(ctx sdk.Context, bondID bond.ID, id types.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getBondIDToRecordsIndexKey(bondID, id))
}

// HasRecord - checks if a record by the given ID exists.
func (k Keeper) HasRecord(ctx sdk.Context, id types.ID) bool {
	return HasRecord(ctx.KVStore(k.storeKey), id)
}

// HasRecord - checks if a record by the given ID exists.
func HasRecord(store sdk.KVStore, id types.ID) bool {
	return store.Has(GetRecordIndexKey(id))
}

// GetRecord - gets a record from the store.
func (k Keeper) GetRecord(ctx sdk.Context, id types.ID) types.Record {
	return GetRecord(ctx.KVStore(k.storeKey), k.cdc, id)
}

// GetRecord - gets a record from the store.
func GetRecord(store sdk.KVStore, codec *amino.Codec, id types.ID) types.Record {
	bz := store.Get(GetRecordIndexKey(id))
	var obj types.RecordObj
	codec.MustUnmarshalBinaryBare(bz, &obj)

	return recordObjToRecord(store, codec, obj)
}

// ListRecords - get all records.
func (k Keeper) ListRecords(ctx sdk.Context) []types.Record {
	var records []types.Record

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixCIDToRecordIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.RecordObj
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			records = append(records, recordObjToRecord(store, k.cdc, obj))
		}
	}

	return records
}

// MatchRecords - get all matching records.
func (k Keeper) MatchRecords(ctx sdk.Context, matchFn func(*types.Record) bool) []*types.Record {
	return MatchRecords(ctx.KVStore(k.storeKey), k.cdc, matchFn)
}

// MatchRecords - get all matching records.
func MatchRecords(store sdk.KVStore, codec *amino.Codec, matchFn func(*types.Record) bool) []*types.Record {
	var records []*types.Record

	itr := sdk.KVStorePrefixIterator(store, PrefixCIDToRecordIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.RecordObj
			codec.MustUnmarshalBinaryBare(bz, &obj)
			record := recordObjToRecord(store, codec, obj)
			if matchFn(&record) {
				records = append(records, &record)
			}
		}
	}

	return records
}

// QueryRecordsByBond - get all records for the given bond.
func (k RecordKeeper) QueryRecordsByBond(ctx sdk.Context, bondID bond.ID) []types.Record {
	var records []types.Record

	bondIDPrefix := append(PrefixBondIDToRecordsIndex, []byte(bondID)...)
	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, bondIDPrefix)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		cid := itr.Key()[len(bondIDPrefix):]
		bz := store.Get(append(PrefixCIDToRecordIndex, cid...))
		if bz != nil {
			var obj types.RecordObj
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			records = append(records, recordObjToRecord(store, k.cdc, obj))
		}
	}

	return records
}

// ModuleName returns the module name.
func (k RecordKeeper) ModuleName() string {
	return types.ModuleName
}

// UsesBond returns true if the bond has associated records.
func (k RecordKeeper) UsesBond(ctx sdk.Context, bondID bond.ID) bool {
	bondIDPrefix := append(PrefixBondIDToRecordsIndex, []byte(bondID)...)
	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, bondIDPrefix)
	defer itr.Close()
	return itr.Valid()
}

func bondUsedInRecord(store sdk.KVStore, bondID bond.ID) bool {
	bondIDPrefix := append(PrefixBondIDToRecordsIndex, []byte(bondID)...)
	itr := sdk.KVStorePrefixIterator(store, bondIDPrefix)
	defer itr.Close()
	return itr.Valid()
}

func bondUsedInAuthority(store sdk.KVStore, bondID bond.ID) bool {
	bondIDPrefix := append(PrefixBondIDToRecordsIndex, []byte(bondID)...)
	itr := sdk.KVStorePrefixIterator(store, bondIDPrefix)
	defer itr.Close()
	return itr.Valid()
}

// getRecordExpiryQueueTimeKey gets the prefix for the record expiry queue.
func getRecordExpiryQueueTimeKey(timestamp time.Time) []byte {
	timeBytes := sdk.FormatTimeBytes(timestamp)
	return append(PrefixExpiryTimeToRecordsIndex, timeBytes...)
}

// GetRecordExpiryQueueTimeSlice gets a specific record queue timeslice.
// A timeslice is a slice of CIDs corresponding to records that expire at a certain time.
func (k Keeper) GetRecordExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (cids []types.ID) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(getRecordExpiryQueueTimeKey(timestamp))
	if bz == nil {
		return []types.ID{}
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cids)
	return cids
}

// SetRecordExpiryQueueTimeSlice sets a specific record expiry queue timeslice.
func (k Keeper) SetRecordExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time, cids []types.ID) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cids)
	store.Set(getRecordExpiryQueueTimeKey(timestamp), bz)
}

// DeleteRecordExpiryQueueTimeSlice deletes a specific record expiry queue timeslice.
func (k Keeper) DeleteRecordExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getRecordExpiryQueueTimeKey(timestamp))
}

// InsertRecordExpiryQueue inserts a record CID to the appropriate timeslice in the record expiry queue.
func (k Keeper) InsertRecordExpiryQueue(ctx sdk.Context, val types.Record) {
	timeSlice := k.GetRecordExpiryQueueTimeSlice(ctx, val.ExpiryTime)
	timeSlice = append(timeSlice, val.ID)
	k.SetRecordExpiryQueueTimeSlice(ctx, val.ExpiryTime, timeSlice)
}

// DeleteRecordExpiryQueue deletes a record CID from the record expiry queue.
func (k Keeper) DeleteRecordExpiryQueue(ctx sdk.Context, record types.Record) {
	timeSlice := k.GetRecordExpiryQueueTimeSlice(ctx, record.ExpiryTime)
	newTimeSlice := []types.ID{}

	for _, cid := range timeSlice {
		if !bytes.Equal([]byte(cid), []byte(record.ID)) {
			newTimeSlice = append(newTimeSlice, cid)
		}
	}

	if len(newTimeSlice) == 0 {
		k.DeleteRecordExpiryQueueTimeSlice(ctx, record.ExpiryTime)
	} else {
		k.SetRecordExpiryQueueTimeSlice(ctx, record.ExpiryTime, newTimeSlice)
	}
}

// RecordExpiryQueueIterator returns all the record expiry queue timeslices from time 0 until endTime.
func (k Keeper) RecordExpiryQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	rangeEndBytes := sdk.InclusiveEndBytes(getRecordExpiryQueueTimeKey(endTime))
	return store.Iterator(PrefixExpiryTimeToRecordsIndex, rangeEndBytes)
}

// GetAllExpiredRecords returns a concatenated list of all the timeslices before currTime.
func (k Keeper) GetAllExpiredRecords(ctx sdk.Context, currTime time.Time) (expiredRecordCIDs []types.ID) {
	// Gets an iterator for all timeslices from time 0 until the current block header time.
	itr := k.RecordExpiryQueueIterator(ctx, ctx.BlockHeader().Time)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		timeslice := []types.ID{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(itr.Value(), &timeslice)
		expiredRecordCIDs = append(expiredRecordCIDs, timeslice...)
	}

	return expiredRecordCIDs
}

// ProcessRecordExpiryQueue tries to renew expiring records (by collecting rent) else marks them as deleted.
func (k Keeper) ProcessRecordExpiryQueue(ctx sdk.Context) {
	cids := k.GetAllExpiredRecords(ctx, ctx.BlockHeader().Time)
	for _, cid := range cids {
		record := k.GetRecord(ctx, cid)

		// If record doesn't have an associated bond or if bond no longer exists, mark it deleted.
		if record.BondID == "" || !k.bondKeeper.HasBond(ctx, record.BondID) {
			record.Deleted = true
			k.PutRecord(ctx, record)
			k.DeleteRecordExpiryQueue(ctx, record)

			return
		}

		// Try to renew the record by taking rent.
		k.TryTakeRecordRent(ctx, record)
	}
}

func (k Keeper) GetRecordExpiryQueue(ctx sdk.Context) (expired map[string][]types.ID) {
	records := make(map[string][]types.ID)

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixExpiryTimeToRecordsIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		var record []types.ID
		k.cdc.MustUnmarshalBinaryLengthPrefixed(itr.Value(), &record)
		records[string(itr.Key()[len(PrefixExpiryTimeToRecordsIndex):])] = record
	}

	return records
}

// TryTakeRecordRent tries to take rent from the record bond.
func (k Keeper) TryTakeRecordRent(ctx sdk.Context, record types.Record) {
	rent, err := sdk.ParseCoins(k.RecordRent(ctx))
	if err != nil {
		panic("Invalid record rent.")
	}

	sdkErr := k.bondKeeper.TransferCoinsToModuleAccount(ctx, record.BondID, types.RecordRentModuleAccountName, rent)
	if sdkErr != nil {
		// Insufficient funds, mark record as deleted.
		record.Deleted = true
		k.PutRecord(ctx, record)
		k.DeleteRecordExpiryQueue(ctx, record)

		return
	}

	// Delete old expiry queue entry, create new one.
	k.DeleteRecordExpiryQueue(ctx, record)
	record.ExpiryTime = ctx.BlockHeader().Time.Add(k.RecordRentDuration(ctx))
	k.InsertRecordExpiryQueue(ctx, record)

	// Save record.
	record.Deleted = false
	k.PutRecord(ctx, record)
	k.AddBondToRecordIndexEntry(ctx, record.BondID, record.ID)
}

func recordObjToRecord(store sdk.KVStore, codec *amino.Codec, obj types.RecordObj) types.Record {
	record := obj.ToRecord()

	reverseNameIndexKey := GetCIDToNamesIndexKey(obj.ID)
	if store.Has(reverseNameIndexKey) {
		var names []string
		codec.MustUnmarshalBinaryBare(store.Get(reverseNameIndexKey), &names)
		record.Names = names
	}

	return record
}

// GetModuleBalances gets the nameservice module account(s) balances.
func (k Keeper) GetModuleBalances(ctx sdk.Context) map[string]sdk.Coins {
	balances := map[string]sdk.Coins{}
	accountNames := []string{types.RecordRentModuleAccountName, types.AuthorityRentModuleAccountName}

	for _, accountName := range accountNames {
		moduleAddress := k.supplyKeeper.GetModuleAddress(accountName)
		moduleAccount := k.accountKeeper.GetAccount(ctx, moduleAddress)
		if moduleAccount != nil {
			balances[accountName] = moduleAccount.GetCoins()
		}
	}

	return balances
}
