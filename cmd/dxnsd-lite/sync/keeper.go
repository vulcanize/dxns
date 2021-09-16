//
// Copyright 2020 Wireline, Inc.
//

package sync

import (
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/tendermint/go-amino"
	"github.com/vulcanize/dxns/x/auction"
	ns "github.com/vulcanize/dxns/x/nameservice"
)

// Keeper is an impl. of an interface similar to the nameservice Keeper.
type Keeper struct {
	config *Config
	codec  *amino.Codec
	store  store.KVStore
}

// NewKeeper creates a new keeper.
func NewKeeper(ctx *Context) *Keeper {
	return &Keeper{config: ctx.config, codec: ctx.codec, store: ctx.store}
}

// Status represents the sync status of the node.
type Status struct {
	LastSyncedHeight int64
	CatchingUp       bool
}

// GetChainID gets the chain ID.
func (k Keeper) GetChainID() string {
	return k.config.ChainID
}

// HasStatusRecord checks if the store has a status record.
func (k Keeper) HasStatusRecord() bool {
	return k.store.Has(ns.KeySyncStatus)
}

// GetStatusRecord gets the sync status record.
func (k Keeper) GetStatusRecord() Status {
	bz := k.store.Get(ns.KeySyncStatus)
	var status Status
	k.codec.MustUnmarshalBinaryBare(bz, &status)

	return status
}

// SaveStatus saves the sync status record.
func (k Keeper) SaveStatus(status Status) {
	bz := k.codec.MustMarshalBinaryBare(status)
	k.store.Set(ns.KeySyncStatus, bz)
}

// HasRecord - checks if a record by the given ID exists.
func (k Keeper) HasRecord(id ns.ID) bool {
	return ns.HasRecord(k.store, id)
}

// GetRecord - gets a record from the store.
func (k Keeper) GetRecord(id ns.ID) ns.Record {
	return ns.GetRecord(k.store, k.codec, id)
}

// PutRecord - saves a record to the store and updates ID -> Record index.
func (k Keeper) PutRecord(record ns.RecordObj) {
	k.store.Set(ns.GetRecordIndexKey(record.ID), k.codec.MustMarshalBinaryBare(record))
}

// SetNameAuthorityRecord - sets a name authority record.
func (k Keeper) SetNameAuthorityRecord(name string, nameAuthority ns.NameAuthority) {
	k.store.Set(ns.GetNameAuthorityIndexKey(name), k.codec.MustMarshalBinaryBare(nameAuthority))
}

// SetNameRecordRaw - sets a name record (used during intial sync).
func (k Keeper) SetNameRecordRaw(wrn string, nameRecord ns.NameRecord) {
	k.store.Set(ns.GetNameRecordIndexKey(wrn), k.codec.MustMarshalBinaryBare(nameRecord))
}

// SetNameRecord - sets a name record.
func (k Keeper) SetNameRecord(wrn string, nameRecord ns.NameRecord) {
	ns.SetNameRecord(k.store, k.codec, wrn, nameRecord.ID, nameRecord.Height)
}

// ResolveWRN resolves a WRN to a record.
func (k Keeper) ResolveWRN(wrn string) *ns.Record {
	record, _ := ns.ResolveWRN(k.store, k.codec, wrn)

	return record
}

// GetNameAuthority get the name authority record for an authority name.
func (k Keeper) GetNameAuthority(name string) *ns.NameAuthority {
	return ns.GetNameAuthority(k.store, k.codec, name)
}

// GetNameRecord get the name record for a name/WRN.
func (k Keeper) GetNameRecord(name string) *ns.NameRecord {
	return ns.GetNameRecord(k.store, k.codec, name)
}

// MatchRecords - get all matching records.
func (k Keeper) MatchRecords(matchFn func(*ns.Record) bool) []*ns.Record {
	return ns.MatchRecords(k.store, k.codec, matchFn)
}

// GetAuction get the auction record.
func (k Keeper) GetAuction(id auction.ID) *auction.Auction {
	return auction.GetAuction(k.store, k.codec, id)
}

// GetBids get the auction bids.
func (k Keeper) GetBids(id auction.ID) []*auction.Bid {
	return auction.GetBids(k.store, k.codec, id)
}

// SaveAuction - saves an auction record.
func (k Keeper) SaveAuction(auctionObj auction.Auction) {
	// Auction ID -> Auction index.
	k.store.Set(auction.GetAuctionIndexKey(auctionObj.ID), k.codec.MustMarshalBinaryBare(auctionObj))

	// Owner -> [Auction] index.
	k.store.Set(auction.GetOwnerToAuctionsIndexKey(auctionObj.OwnerAddress, auctionObj.ID), []byte{})
}

// SaveBid - saves an auction bid.
func (k Keeper) SaveBid(bid auction.Bid) {
	k.store.Set(auction.GetBidIndexKey(bid.AuctionID, bid.BidderAddress), k.codec.MustMarshalBinaryBare(bid))
}
