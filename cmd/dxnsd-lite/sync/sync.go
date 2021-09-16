//
// Copyright 2020 Wireline, Inc.
//

package sync

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/vulcanize/dxns/x/auction"
	ns "github.com/vulcanize/dxns/x/nameservice"
)

// AggressiveSyncIntervalInMillis is the interval for aggressive sync, to catch up quickly to the current height.
const AggressiveSyncIntervalInMillis = 250

// SyncIntervalInMillis is the interval for initiating incremental sync, when already caught up to current height.
const SyncIntervalInMillis = 5 * 1000

// ErrorWaitDurationMillis is the wait duration in case of errors.
const ErrorWaitDurationMillis = 1 * 1000

// SyncLaggingMinHeightDiff is the min. difference in height to consider a lite node as lagging the full node.
const SyncLaggingMinHeightDiff = 5

// DumpRPCNodeStatsFrequencyMillis controls frequency to dump RPC node stats.
const DumpRPCNodeStatsFrequencyMillis = 60 * 1000

// DiscoverRPCNodesFrequencyMillis controls frequency to discover new RPC endpoints.
const DiscoverRPCNodesFrequencyMillis = 60 * 1000

// Init sets up the lite node.
func Init(ctx *Context, height int64) {
	// If sync record exists, abort with error.
	if ctx.keeper.HasStatusRecord() {
		ctx.log.Fatalln("Node already initialized, aborting.")
	}

	if !ctx.config.InitFromNode && !ctx.config.InitFromGenesisFile {
		ctx.log.Fatalln("Must pass one of `--from-node` and `--from-genesis-file`.")
	}

	if ctx.config.InitFromNode {
		initFromNode(ctx)
	} else if ctx.config.InitFromGenesisFile {
		initFromGenesisFile(ctx, height)
	}
}

// Start initiates the sync process.
func Start(ctx *Context) {
	// Fail if node has no sync status record.
	if !ctx.keeper.HasStatusRecord() {
		ctx.log.Fatalln("Node not initialized, aborting.")
	}

	go dumpConnectionStatsOnTimer(ctx)

	if ctx.config.SyncTimeoutMins > 0 {
		ctx.log.Infoln("Sync timeout ON:", ctx.config.SyncTimeoutMins)
		go exitOnSyncTimeout(ctx)
	} else {
		ctx.log.Infoln("Sync timeout OFF.")
	}

	if ctx.config.Endpoint != "" {
		go discoverRPCNodesOnTimer(ctx)
		ctx.log.Infoln("RPC endpoint discovery ON:", ctx.config.Endpoint)
	} else {
		ctx.log.Infoln("RPC endpoint discovery OFF.")
	}

	syncStatus := ctx.keeper.GetStatusRecord()
	lastSyncedHeight := syncStatus.LastSyncedHeight

	for {
		chainCurrentHeight, err := ctx.PrimaryNode.getCurrentHeight()
		if err != nil {
			logErrorAndWait(ctx, err)
			continue
		}

		if lastSyncedHeight > chainCurrentHeight {
			// Maybe we've connected to a new primary node (after restart) and that isn't fully caught up, yet. Just wait.
			logErrorAndWait(ctx, errors.New("last synced height greater than current chain height"))
			continue
		}

		newSyncHeight := lastSyncedHeight + 1
		if newSyncHeight > chainCurrentHeight {
			// Can't sync beyond chain height, just wait.
			waitAfterSync(chainCurrentHeight, chainCurrentHeight)
			continue
		}

		err = syncAtHeight(ctx, newSyncHeight)
		if err != nil {
			logErrorAndWait(ctx, err)
			continue
		}

		// Saved last synced height in db.
		lastSyncedHeight = newSyncHeight
		catchingUp := (chainCurrentHeight - lastSyncedHeight) > SyncLaggingMinHeightDiff

		ctx.keeper.SaveStatus(Status{
			LastSyncedHeight: lastSyncedHeight,
			CatchingUp:       catchingUp,
		})

		waitAfterSync(chainCurrentHeight, lastSyncedHeight)
	}
}

// syncAtHeight runs a sync cycle for the given height.
func syncAtHeight(ctx *Context, height int64) error {
	rpc := getRandomRPCNodeHandler(ctx)

	ctx.log.Infoln("Syncing from", rpc.Address, "at height:", height)

	changeset, err := rpc.getBlockChangeset(ctx, height)
	if err != nil {
		return err
	}

	if changeset.Height <= 0 {
		// No changeset for this block, ignore.
		return nil
	}

	ctx.log.Debugln("Syncing changeset:", changeset)

	// Sync records.
	err = rpc.syncRecords(ctx, height, changeset.Records)
	if err != nil {
		return err
	}

	// Sync auctions.
	err = rpc.syncAuctions(ctx, height, changeset.Auctions)
	if err != nil {
		return err
	}

	// Sync auction bids.
	err = rpc.syncAuctionBids(ctx, height, changeset.AuctionBids)
	if err != nil {
		return err
	}

	// Sync name authority records.
	err = rpc.syncNameAuthorityRecords(ctx, height, changeset.NameAuthorities)
	if err != nil {
		return err
	}

	// Sync name records.
	err = rpc.syncNameRecords(ctx, height, changeset.Names)
	if err != nil {
		return err
	}

	// Flush cache changes to underlying store.
	ctx.cache.Write()

	return nil
}

func (rpc *RPCNodeHandler) syncRecords(ctx *Context, height int64, records []ns.ID) error {
	for _, id := range records {
		recordKey := ns.GetRecordIndexKey(id)
		value, err := rpc.getStoreValue(ctx, NameStorePath, recordKey, height)
		if err != nil {
			return err
		}

		ctx.cache.Set(recordKey, value)
	}

	return nil
}

func (rpc *RPCNodeHandler) syncAuctions(ctx *Context, height int64, auctions []auction.ID) error {
	for _, id := range auctions {
		auctionKey := auction.GetAuctionIndexKey(id)
		value, err := rpc.getStoreValue(ctx, AuctionStorePath, auctionKey, height)
		if err != nil {
			return err
		}

		ctx.cache.Set(auctionKey, value)
	}

	return nil
}

func (rpc *RPCNodeHandler) syncAuctionBids(ctx *Context, height int64, bids []auction.AuctionBidInfo) error {
	for _, bid := range bids {
		bidKey := auction.GetBidIndexKey(bid.AuctionID, bid.BidderAddress)
		value, err := rpc.getStoreValue(ctx, AuctionStorePath, bidKey, height)
		if err != nil {
			return err
		}

		ctx.cache.Set(bidKey, value)
	}

	return nil
}

func (rpc *RPCNodeHandler) syncNameAuthorityRecords(ctx *Context, height int64, nameAuthorities []string) error {
	for _, name := range nameAuthorities {
		nameAuhorityRecordKey := ns.GetNameAuthorityIndexKey(name)
		value, err := rpc.getStoreValue(ctx, NameStorePath, nameAuhorityRecordKey, height)
		if err != nil {
			return err
		}

		ctx.cache.Set(nameAuhorityRecordKey, value)
	}

	return nil
}

func (rpc *RPCNodeHandler) syncNameRecords(ctx *Context, height int64, names []string) error {
	for _, name := range names {
		nameRecordKey := ns.GetNameRecordIndexKey(name)
		value, err := rpc.getStoreValue(ctx, NameStorePath, nameRecordKey, height)
		if err != nil {
			return err
		}

		ctx.cache.Set(nameRecordKey, value)

		// Update Record ID -> []Names index.
		nameRecord := ns.GetNameRecord(ctx.cache, ctx.codec, name)
		if nameRecord.ID != "" {
			// Same name might have pointed to another record earlier, should be in history.
			// Delete that mapping.
			removeOldNameMapping(ctx, name, nameRecord)

			// Set name.
			ns.AddRecordToNameMapping(ctx.cache, ctx.codec, nameRecord.ID, name)
		} else {
			// Delete name. ID of old record should be in history.
			removeOldNameMapping(ctx, name, nameRecord)
		}
	}

	return nil
}

func removeOldNameMapping(ctx *Context, name string, nameRecord *ns.NameRecord) {
	historyCount := len(nameRecord.History)
	if historyCount > 0 {
		oldNameEntry := nameRecord.History[historyCount-1]
		if oldNameEntry.ID != "" {
			ns.RemoveRecordToNameMapping(ctx.cache, ctx.codec, oldNameEntry.ID, name)
		}
	}
}

func waitAfterSync(chainCurrentHeight int64, lastSyncedHeight int64) {
	if chainCurrentHeight == lastSyncedHeight {
		// Caught up to current chain height, don't have to poll aggressively now.
		time.Sleep(SyncIntervalInMillis * time.Millisecond)
	} else {
		// Still catching up to current height, poll more aggressively.
		time.Sleep(AggressiveSyncIntervalInMillis * time.Millisecond)
	}
}

func logErrorAndWait(ctx *Context, err error) {
	ctx.log.Errorln(err)

	// TODO(ashwin): Exponential backoff logic.
	time.Sleep(ErrorWaitDurationMillis * time.Millisecond)
}

func initFromNode(ctx *Context) {
	height, err := ctx.PrimaryNode.getCurrentHeight()
	if err != nil {
		ctx.log.Fatalln("Error fetching current height:", err)
	}

	ctx.log.Debugln("Current block height:", height)

	recordKVs, err := ctx.getStoreSubspace("nameservice", ns.PrefixCIDToRecordIndex, height)
	if err != nil {
		ctx.log.Fatalln("Error fetching records", err)
	}

	for _, kv := range recordKVs {
		var record ns.RecordObj
		ctx.codec.MustUnmarshalBinaryBare(kv.Value, &record)
		ctx.log.Debugln("Importing record", record.ID)
		ctx.keeper.PutRecord(record)
	}

	auctionBidKVs, err := ctx.getStoreSubspace("auction", auction.PrefixAuctionBidsIndex, height)
	if err != nil {
		ctx.log.Fatalln("Error fetching auction bid records", err)
	}

	for _, kv := range auctionBidKVs {
		var bid auction.Bid
		ctx.codec.MustUnmarshalBinaryBare(kv.Value, &bid)
		ctx.log.Debugln("Importing auction bid", bid.BidderAddress)
		ctx.keeper.SaveBid(bid)
	}

	auctionKVs, err := ctx.getStoreSubspace("auction", auction.PrefixIDToAuctionIndex, height)
	if err != nil {
		ctx.log.Fatalln("Error fetching auction records", err)
	}

	for _, kv := range auctionKVs {
		var auctionRecord auction.Auction
		ctx.codec.MustUnmarshalBinaryBare(kv.Value, &auctionRecord)
		id := kv.Key[len(auction.PrefixIDToAuctionIndex):]
		ctx.log.Debugln("Importing auction", auction.ID(id))
		ctx.keeper.SaveAuction(auctionRecord)
	}

	authorityKVs, err := ctx.getStoreSubspace("nameservice", ns.PrefixNameAuthorityRecordIndex, height)
	if err != nil {
		ctx.log.Fatalln("Error fetching authority records", err)
	}

	for _, kv := range authorityKVs {
		var authorityRecord ns.NameAuthority
		ctx.codec.MustUnmarshalBinaryBare(kv.Value, &authorityRecord)
		name := string(kv.Key[len(ns.PrefixNameAuthorityRecordIndex):])
		ctx.log.Debugln("Importing authority", name)
		ctx.keeper.SetNameAuthorityRecord(name, authorityRecord)
	}

	namesKVs, err := ctx.getStoreSubspace("nameservice", ns.PrefixWRNToNameRecordIndex, height)
	if err != nil {
		ctx.log.Fatalln("Error fetching name records", err)
	}

	for _, kv := range namesKVs {
		var nameRecord ns.NameRecord
		ctx.codec.MustUnmarshalBinaryBare(kv.Value, &nameRecord)
		wrn := string(kv.Key[len(ns.PrefixWRNToNameRecordIndex):])
		ctx.log.Debugln("Importing name", wrn)

		ctx.keeper.SetNameRecordRaw(wrn, nameRecord)
		if nameRecord.ID != "" {
			ns.AddRecordToNameMapping(ctx.store, ctx.codec, nameRecord.ID, wrn)
		}
	}

	// Create sync status record.
	ctx.keeper.SaveStatus(Status{LastSyncedHeight: height})
}

func initFromGenesisFile(ctx *Context, height int64) {
	// Create <home>/config directory if it doesn't exist.
	configDirPath := filepath.Join(ctx.config.Home, "config")
	os.Mkdir(configDirPath, 0755)

	// Import genesis.json.
	genesisJSONPath := filepath.Join(configDirPath, "genesis.json")
	_, err := os.Stat(genesisJSONPath)
	if err != nil {
		ctx.log.Fatalln("Genesis file error:", err)
	}

	geneisState := GenesisState{}
	bytes, err := ioutil.ReadFile(genesisJSONPath)
	if err != nil {
		ctx.log.Fatalln(err)
	}

	err = ctx.codec.UnmarshalJSON(bytes, &geneisState)
	if err != nil {
		ctx.log.Fatalln(err)
	}

	// Check that chain-id matches.
	if geneisState.ChainID != ctx.config.ChainID {
		ctx.log.Fatalln("Chain ID mismatch:", genesisJSONPath)
	}

	authorities := geneisState.AppState.Nameservice.Authorities
	for _, nameAuthority := range authorities {
		ctx.keeper.SetNameAuthorityRecord(nameAuthority.Name, nameAuthority.Entry)
	}

	names := geneisState.AppState.Nameservice.Names
	for _, nameEntry := range names {
		ctx.keeper.SetNameRecord(nameEntry.Name, nameEntry.Entry)
	}

	records := geneisState.AppState.Nameservice.Records
	for _, record := range records {
		ctx.keeper.PutRecord(record)
	}

	// Create sync status record.
	ctx.keeper.SaveStatus(Status{LastSyncedHeight: height})
}

func getRandomRPCNodeHandler(ctx *Context) *RPCNodeHandler {
	ctx.nodeLock.RLock()
	defer ctx.nodeLock.RUnlock()

	// TODO(ashwin): Intelligent selection of nodes (e.g. based on QoS).
	nodes := ctx.secondaryNodes
	keys := reflect.ValueOf(nodes).MapKeys()
	address := keys[rand.Intn(len(keys))].Interface().(string)
	rpc := nodes[address]

	return rpc
}

func dumpConnectionStatsOnTimer(ctx *Context) {
	for {
		time.Sleep(DumpRPCNodeStatsFrequencyMillis * time.Millisecond)
		dumpConnectionStats(ctx)
	}
}

func exitOnSyncTimeout(ctx *Context) {
	prevHeight := ctx.keeper.GetStatusRecord().LastSyncedHeight

	for {
		time.Sleep(time.Duration(ctx.config.SyncTimeoutMins) * time.Minute)
		currentHeight := ctx.keeper.GetStatusRecord().LastSyncedHeight
		if currentHeight <= prevHeight {
			// No progress for quite some time. Quit node.
			ctx.log.Fatalln("Sync timed out at height:", currentHeight)
		}

		prevHeight = currentHeight
	}
}

func dumpConnectionStats(ctx *Context) {
	ctx.nodeLock.RLock()
	defer ctx.nodeLock.RUnlock()

	// Log RPC node stats.
	bytes, _ := json.Marshal(ctx.secondaryNodes)
	ctx.log.Debugln(string(bytes))
}

func discoverRPCNodesOnTimer(ctx *Context) {
	for {
		discoverRPCNodes(ctx)
		time.Sleep(DiscoverRPCNodesFrequencyMillis * time.Millisecond)
	}
}

// Discover new RPC nodes.
func discoverRPCNodes(ctx *Context) {
	rpcEndpoints, err := DiscoverRPCEndpoints(ctx, ctx.config.Endpoint)
	if err != nil {
		ctx.log.Errorln("Error discovering RPC endpoints", err)
		return
	}

	ctx.log.Debugln("RPC endpoints:", rpcEndpoints)

	ctx.nodeLock.Lock()
	defer ctx.nodeLock.Unlock()

	for _, rpcEndpoint := range rpcEndpoints {
		if _, exists := ctx.secondaryNodes[rpcEndpoint]; !exists {
			ctx.log.Infoln("Added new RPC endpoint:", rpcEndpoint)
			rpc := NewRPCNodeHandler(rpcEndpoint)
			ctx.secondaryNodes[rpcEndpoint] = rpc
		}
	}
}
