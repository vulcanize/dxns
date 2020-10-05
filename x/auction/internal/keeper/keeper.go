//
// Copyright 2020 Wireline, Inc.
//

package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/go-amino"
	wnsUtils "github.com/wirelineio/dxns/utils"
	"github.com/wirelineio/dxns/x/auction/internal/types"
)

// CompletedAuctionDeleteTimeout => Completed auctions are deleted after this timeout (after reveals end time).
const CompletedAuctionDeleteTimeout time.Duration = time.Hour * 24

// PrefixIDToAuctionIndex is the prefix for ID -> Auction index in the KVStore.
// Note: This is the primary index in the system.
// Note: Golang doesn't support const arrays.
var PrefixIDToAuctionIndex = []byte{0x00}

// prefixOwnerToAuctionsIndex is the prefix for the Owner -> [Auction] index in the KVStore.
var prefixOwnerToAuctionsIndex = []byte{0x01}

// PrefixAuctionBidsIndex is the prefix for the (auction, bidder) -> Bid index in the KVStore.
var PrefixAuctionBidsIndex = []byte{0x02}

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper

	// Track auction usage in other cosmos-sdk modules (more like a usage tracker).
	usageKeepers []types.AuctionUsageKeeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.

	paramSubspace subspace.Subspace
}

// AuctionClientKeeper is the subset of functionality exposed to other modules.
type AuctionClientKeeper interface {
	HasAuction(ctx sdk.Context, id types.ID) bool
	GetAuction(ctx sdk.Context, id types.ID) types.Auction
	MatchAuctions(ctx sdk.Context, matchFn func(*types.Auction) bool) []*types.Auction
}

// NewKeeper creates new instances of the auction Keeper
func NewKeeper(accountKeeper auth.AccountKeeper, bankKeeper bank.Keeper, supplyKeeper supply.Keeper,
	storeKey sdk.StoreKey, cdc *codec.Codec, paramstore params.Subspace) Keeper {
	return Keeper{
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		supplyKeeper:  supplyKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}

func (k *Keeper) SetUsageKeepers(usageKeepers []types.AuctionUsageKeeper) {
	k.usageKeepers = usageKeepers
}

func (k Keeper) GetUsageKeepers() []types.AuctionUsageKeeper {
	return k.usageKeepers
}

// Generates Auction ID -> Auction index key.
func GetAuctionIndexKey(id types.ID) []byte {
	return append(PrefixIDToAuctionIndex, []byte(id)...)
}

// Generates Owner -> Auctions index key.
func GetOwnerToAuctionsIndexKey(owner string, auctionID types.ID) []byte {
	return append(append(prefixOwnerToAuctionsIndex, []byte(owner)...), []byte(auctionID)...)
}

func GetBidIndexKey(auctionID types.ID, bidder string) []byte {
	return append(GetAuctionBidsIndexPrefix(auctionID), []byte(bidder)...)
}

func GetAuctionBidsIndexPrefix(auctionID types.ID) []byte {
	return append(append(PrefixAuctionBidsIndex, []byte(auctionID)...))
}

// SaveAuction - saves a auction to the store.
func (k Keeper) SaveAuction(ctx sdk.Context, auction types.Auction) {
	store := ctx.KVStore(k.storeKey)

	// Auction ID -> Auction index.
	store.Set(GetAuctionIndexKey(auction.ID), k.cdc.MustMarshalBinaryBare(auction))

	// Owner -> [Auction] index.
	store.Set(GetOwnerToAuctionsIndexKey(auction.OwnerAddress, auction.ID), []byte{})

	// Notify interested parties.
	for _, keeper := range k.usageKeepers {
		keeper.OnAuction(ctx, auction.ID)
	}
}

func (k Keeper) SaveBid(ctx sdk.Context, bid types.Bid) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetBidIndexKey(bid.AuctionID, bid.BidderAddress), k.cdc.MustMarshalBinaryBare(bid))

	// Notify interested parties.
	for _, keeper := range k.usageKeepers {
		keeper.OnAuctionBid(ctx, bid.AuctionID, bid.BidderAddress)
	}
}

func (k Keeper) DeleteBid(ctx sdk.Context, bid types.Bid) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetBidIndexKey(bid.AuctionID, bid.BidderAddress))
}

// HasAuction - checks if a auction by the given ID exists.
func (k Keeper) HasAuction(ctx sdk.Context, id types.ID) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetAuctionIndexKey(id))
}

func (k Keeper) HasBid(ctx sdk.Context, id types.ID, bidder string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetBidIndexKey(id, bidder))
}

// DeleteAuction - deletes the auction.
func (k Keeper) DeleteAuction(ctx sdk.Context, auction types.Auction) {
	// Delete all bids first.
	bids := k.GetBids(ctx, auction.ID)
	for _, bid := range bids {
		k.DeleteBid(ctx, *bid)
	}

	// Delete the auction itself.
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetAuctionIndexKey(auction.ID))
	store.Delete(GetOwnerToAuctionsIndexKey(auction.OwnerAddress, auction.ID))
}

// GetAuction - gets a record from the store.
func (k Keeper) GetAuction(ctx sdk.Context, id types.ID) *types.Auction {
	return GetAuction(ctx.KVStore(k.storeKey), k.cdc, id)
}

func GetAuction(store sdk.KVStore, codec *amino.Codec, id types.ID) *types.Auction {
	auctionKey := GetAuctionIndexKey(id)
	if !store.Has(auctionKey) {
		return nil
	}

	bz := store.Get(auctionKey)
	var obj types.Auction
	codec.MustUnmarshalBinaryBare(bz, &obj)

	return &obj
}

// GetBids gets the auction bids.
func (k Keeper) GetBids(ctx sdk.Context, id types.ID) []*types.Bid {
	return GetBids(ctx.KVStore(k.storeKey), k.cdc, id)
}

func GetBids(store sdk.KVStore, codec *amino.Codec, id types.ID) []*types.Bid {
	var bids []*types.Bid

	itr := sdk.KVStorePrefixIterator(store, GetAuctionBidsIndexPrefix(id))
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.Bid
			codec.MustUnmarshalBinaryBare(bz, &obj)
			bids = append(bids, &obj)
		}
	}

	return bids
}

func (k Keeper) GetBid(ctx sdk.Context, id types.ID, bidder string) types.Bid {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(GetBidIndexKey(id, bidder))
	var obj types.Bid
	k.cdc.MustUnmarshalBinaryBare(bz, &obj)

	return obj
}

// ListAuctions - get all auctions.
func (k Keeper) ListAuctions(ctx sdk.Context) []types.Auction {
	var auctions []types.Auction

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixIDToAuctionIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.Auction
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			auctions = append(auctions, obj)
		}
	}

	return auctions
}

// QueryAuctionsByOwner - query auctions by owner.
func (k Keeper) QueryAuctionsByOwner(ctx sdk.Context, ownerAddress string) []types.Auction {
	var auctions []types.Auction

	ownerPrefix := append(prefixOwnerToAuctionsIndex, []byte(ownerAddress)...)
	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, ownerPrefix)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		auctionID := itr.Key()[len(ownerPrefix):]
		bz := store.Get(append(PrefixIDToAuctionIndex, auctionID...))
		if bz != nil {
			var obj types.Auction
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			auctions = append(auctions, obj)
		}
	}

	return auctions
}

// MatchAuctions - get all matching auctions.
func (k Keeper) MatchAuctions(ctx sdk.Context, matchFn func(*types.Auction) bool) []*types.Auction {
	var auctions []*types.Auction

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixIDToAuctionIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var obj types.Auction
			k.cdc.MustUnmarshalBinaryBare(bz, &obj)
			if matchFn(&obj) {
				auctions = append(auctions, &obj)
			}
		}
	}

	return auctions
}

// CreateAuction creates a new auction.
func (k Keeper) CreateAuction(ctx sdk.Context, msg types.MsgCreateAuction) (*types.Auction, error) {
	// Might be called from another module directly, always validate.
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// Generate auction ID.
	account := k.accountKeeper.GetAccount(ctx, msg.Signer)
	if account == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Account not found.")
	}

	auctionID := types.AuctionID{
		Address:  msg.Signer,
		AccNum:   account.GetAccountNumber(),
		Sequence: account.GetSequence(),
	}.Generate()

	// Compute timestamps.
	now := ctx.BlockTime()
	commitsEndTime := now.Add(msg.CommitsDuration)
	revealsEndTime := now.Add(msg.CommitsDuration + msg.RevealsDuration)

	auction := types.Auction{
		ID:             types.ID(auctionID),
		Status:         types.AuctionStatusCommitPhase,
		OwnerAddress:   msg.Signer.String(),
		CreateTime:     now,
		CommitsEndTime: commitsEndTime,
		RevealsEndTime: revealsEndTime,
		CommitFee:      msg.CommitFee,
		RevealFee:      msg.RevealFee,
		MinimumBid:     msg.MinimumBid,
	}

	// Save auction in store.
	k.SaveAuction(ctx, auction)

	return &auction, nil
}

// CommitBid commits a bid for an auction.
func (k Keeper) CommitBid(ctx sdk.Context, msg types.MsgCommitBid) (*types.Auction, error) {
	if !k.HasAuction(ctx, msg.AuctionID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Auction not found.")
	}

	auction := k.GetAuction(ctx, msg.AuctionID)
	if auction.Status != types.AuctionStatusCommitPhase {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Auction is not in commit phase.")
	}

	// Take auction fees from account.
	totalFee := auction.CommitFee.Add(auction.RevealFee)
	sdkErr := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Signer, types.ModuleName, sdk.NewCoins(totalFee))
	if sdkErr != nil {
		return nil, sdkErr
	}

	// Check if an old bid already exists, if so, return old bids auction fee (update bid scenario).
	bidder := msg.Signer.String()
	if k.HasBid(ctx, msg.AuctionID, bidder) {
		oldBid := k.GetBid(ctx, msg.AuctionID, bidder)
		oldTotalFee := oldBid.CommitFee.Add(oldBid.RevealFee)
		sdkErr := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, msg.Signer, sdk.NewCoins(oldTotalFee))
		if sdkErr != nil {
			return nil, sdkErr
		}
	}

	// Save new bid.
	bid := types.Bid{
		AuctionID:     msg.AuctionID,
		BidderAddress: bidder,
		Status:        types.BidStatusCommitted,
		CommitHash:    msg.CommitHash,
		CommitTime:    ctx.BlockTime(),
		CommitFee:     auction.CommitFee,
		RevealFee:     auction.RevealFee,
	}

	k.SaveBid(ctx, bid)

	return auction, nil
}

// RevealBid reeals a bid comitted earlier.
func (k Keeper) RevealBid(ctx sdk.Context, msg types.MsgRevealBid) (*types.Auction, error) {
	if !k.HasAuction(ctx, msg.AuctionID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Auction not found.")
	}

	auction := k.GetAuction(ctx, msg.AuctionID)
	if auction.Status != types.AuctionStatusRevealPhase {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Auction is not in reveal phase.")
	}

	if !k.HasBid(ctx, msg.AuctionID, msg.Signer.String()) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bid not found.")
	}

	bid := k.GetBid(ctx, auction.ID, msg.Signer.String())
	if bid.Status != types.BidStatusCommitted {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bid not in committed state.")
	}

	revealBytes, err := hex.DecodeString(msg.Reveal)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal string.")
	}

	cid, err := wnsUtils.CIDFromJSONBytes(revealBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal JSON.")
	}

	if bid.CommitHash != cid {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Commit hash mismatch.")
	}

	var reveal map[string]interface{}
	err = json.Unmarshal(revealBytes, &reveal)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Reveal JSON unmarshal error.")
	}

	chainID, err := wnsUtils.GetAttributeAsString(reveal, "chainId")
	if err != nil || chainID != ctx.ChainID() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal chainID.")
	}

	auctionID, err := wnsUtils.GetAttributeAsString(reveal, "auctionId")
	if err != nil || types.ID(auctionID) != msg.AuctionID {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal auction ID.")
	}

	bidderAddress, err := wnsUtils.GetAttributeAsString(reveal, "bidderAddress")
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid address.")
	}

	if bidderAddress != msg.Signer.String() {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Reveal bid address mismatch.")
	}

	bidAmountStr, err := wnsUtils.GetAttributeAsString(reveal, "bidAmount")
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid amount.")
	}

	bidAmount, err := sdk.ParseCoin(bidAmountStr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid reveal bid amount.")
	}

	if bidAmount.IsLT(auction.MinimumBid) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bid is lower than minimum bid.")
	}

	// Lock bid amount.
	sdkErr := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, msg.Signer, types.ModuleName, sdk.NewCoins(bidAmount))
	if sdkErr != nil {
		return nil, sdkErr
	}

	// Update bid.
	bid.BidAmount = bidAmount
	bid.RevealTime = ctx.BlockTime()
	bid.Status = types.BidStatusRevealed
	k.SaveBid(ctx, bid)

	return auction, nil
}

// GetAuctionModuleBalances gets the auction module account(s) balances.
func (k Keeper) GetAuctionModuleBalances(ctx sdk.Context) map[string]sdk.Coins {
	balances := map[string]sdk.Coins{}
	accountNames := []string{types.ModuleName, types.AuctionBurnModuleAccountName}

	for _, accountName := range accountNames {
		moduleAddress := k.supplyKeeper.GetModuleAddress(accountName)
		moduleAccount := k.accountKeeper.GetAccount(ctx, moduleAddress)
		if moduleAccount != nil {
			balances[accountName] = moduleAccount.GetCoins()
		}
	}

	return balances
}

func (k Keeper) EndBlockerProcessAuctions(ctx sdk.Context) {
	// Transition auction state (commit, reveal, expired, completed).
	k.processAuctionPhases(ctx)

	// Delete stale auctions.
	k.deleteCompletedAuctions(ctx)
}

func (k Keeper) processAuctionPhases(ctx sdk.Context) {
	auctions := k.MatchAuctions(ctx, func(_ *types.Auction) bool {
		return true
	})

	for _, auction := range auctions {
		// Commit -> Reveal state.
		if auction.Status == types.AuctionStatusCommitPhase && ctx.BlockTime().After(auction.CommitsEndTime) {
			auction.Status = types.AuctionStatusRevealPhase
			k.SaveAuction(ctx, *auction)
			ctx.Logger().Info(fmt.Sprintf("Moved auction %s to reveal phase.", auction.ID))
		}

		// Reveal -> Expired state.
		if auction.Status == types.AuctionStatusRevealPhase && ctx.BlockTime().After(auction.RevealsEndTime) {
			auction.Status = types.AuctionStatusExpired
			k.SaveAuction(ctx, *auction)
			ctx.Logger().Info(fmt.Sprintf("Moved auction %s to expired state.", auction.ID))
		}

		// If auction has expired, pick a winner from revealed bids.
		if auction.Status == types.AuctionStatusExpired {
			k.pickAuctionWinner(ctx, auction)
		}
	}
}

// Delete completed stale auctions.
func (k Keeper) deleteCompletedAuctions(ctx sdk.Context) {
	auctions := k.MatchAuctions(ctx, func(auction *types.Auction) bool {
		deleteTime := auction.RevealsEndTime.Add(CompletedAuctionDeleteTimeout)
		return auction.Status == types.AuctionStatusCompleted && ctx.BlockTime().After(deleteTime)
	})

	for _, auction := range auctions {
		ctx.Logger().Info(fmt.Sprintf("Deleting completed auction %s after timeout.", auction.ID))
		k.DeleteAuction(ctx, *auction)
	}
}

func (k Keeper) pickAuctionWinner(ctx sdk.Context, auction *types.Auction) {
	ctx.Logger().Info(fmt.Sprintf("Picking auction %s winner.", auction.ID))

	var highestBid *types.Bid
	var secondHighestBid *types.Bid

	bids := k.GetBids(ctx, auction.ID)
	for _, bid := range bids {
		ctx.Logger().Info(fmt.Sprintf("Processing bid %s %s", bid.BidderAddress, bid.BidAmount.String()))

		// Only consider revealed bids.
		if bid.Status != types.BidStatusRevealed {
			ctx.Logger().Info(fmt.Sprintf("Ignoring unrevealed bid %s %s", bid.BidderAddress, bid.BidAmount.String()))

			continue
		}

		// Init highest bid.
		if highestBid == nil {
			highestBid = bid

			ctx.Logger().Info(fmt.Sprintf("Initializing 1st bid %s %s", bid.BidderAddress, bid.BidAmount.String()))

			continue
		}

		if highestBid.BidAmount.IsLT(bid.BidAmount) {
			ctx.Logger().Info(fmt.Sprintf("New highest bid %s %s", bid.BidderAddress, bid.BidAmount.String()))

			secondHighestBid = highestBid
			highestBid = bid

			ctx.Logger().Info(fmt.Sprintf("Updated 1st bid %s %s", highestBid.BidderAddress, highestBid.BidAmount.String()))
			ctx.Logger().Info(fmt.Sprintf("Updated 2nd bid %s %s", secondHighestBid.BidderAddress, secondHighestBid.BidAmount.String()))

		} else if secondHighestBid == nil || secondHighestBid.BidAmount.IsLT(bid.BidAmount) {
			ctx.Logger().Info(fmt.Sprintf("New 2nd highest bid %s %s", bid.BidderAddress, bid.BidAmount.String()))

			secondHighestBid = bid
			ctx.Logger().Info(fmt.Sprintf("Updated 2nd bid %s %s", secondHighestBid.BidderAddress, secondHighestBid.BidAmount.String()))
		} else {
			ctx.Logger().Info(fmt.Sprintf("Ignoring bid as it doesn't affect 1st/2nd price %s %s", bid.BidderAddress, bid.BidAmount.String()))
		}
	}

	// Highest bid is the winner, but pays second highest bid price.
	auction.Status = types.AuctionStatusCompleted

	if highestBid != nil {
		auction.WinnerAddress = highestBid.BidderAddress
		auction.WinnerBid = highestBid.BidAmount

		// Winner pays 2nd price, if a 2nd price exists.
		auction.WinnerPrice = highestBid.BidAmount
		if secondHighestBid != nil {
			auction.WinnerPrice = secondHighestBid.BidAmount
		}

		ctx.Logger().Info(fmt.Sprintf("Auction %s winner %s.", auction.ID, auction.WinnerAddress))
		ctx.Logger().Info(fmt.Sprintf("Auction %s winner bid %s.", auction.ID, auction.WinnerBid.String()))
		ctx.Logger().Info(fmt.Sprintf("Auction %s winner price %s.", auction.ID, auction.WinnerPrice.String()))

	} else {
		ctx.Logger().Info(fmt.Sprintf("Auction %s has no valid revealed bids (no winner).", auction.ID))
	}

	k.SaveAuction(ctx, *auction)

	for _, bid := range bids {
		bidderAddress, err := sdk.AccAddressFromBech32(bid.BidderAddress)
		if err != nil {
			panic("Invalid bidder address.")
		}

		if bid.Status == types.BidStatusRevealed {
			// Send reveal fee back to bidders that've revealed the bid.
			sdkErr := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidderAddress, sdk.NewCoins(bid.RevealFee))
			if sdkErr != nil {
				ctx.Logger().Error(fmt.Sprintf("Auction error returning reveal fee: %v", sdkErr))
				panic(sdkErr)
			}
		}

		// Send back locked bid amount to all bidders.
		sdkErr := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, bidderAddress, sdk.NewCoins(bid.BidAmount))
		if sdkErr != nil {
			ctx.Logger().Error(fmt.Sprintf("Auction error returning bid amount: %v", sdkErr))
			panic(sdkErr)
		}
	}

	// Process winner account (if nobody bids, there won't be a winner).
	if auction.WinnerAddress != "" {
		winnerAddress, err := sdk.AccAddressFromBech32(auction.WinnerAddress)
		if err != nil {
			panic("Invalid winner address.")
		}

		// Take 2nd price from winner.
		sdkErr := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, winnerAddress, types.ModuleName, sdk.NewCoins(auction.WinnerPrice))
		if sdkErr != nil {
			ctx.Logger().Error(fmt.Sprintf("Auction error taking funds from winner: %v", sdkErr))
			panic(sdkErr)
		}

		// Burn anything over the min. bid amount.
		amountToBurn := auction.WinnerPrice.Sub(auction.MinimumBid)
		if amountToBurn.IsNegative() {
			panic("Auction coins to burn cannot be negative.")
		}

		// Use auction burn module account instead of actually burning coins to better keep track of supply.
		sdkErr = k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, types.AuctionBurnModuleAccountName, sdk.NewCoins(amountToBurn))
		if sdkErr != nil {
			ctx.Logger().Error(fmt.Sprintf("Auction error burning coins: %v", sdkErr))
			panic(sdkErr)
		}
	}

	// Notify other modules (hook).
	ctx.Logger().Info(fmt.Sprintf("Auction %s notifying %d modules.", auction.ID, len(k.usageKeepers)))
	for _, keeper := range k.usageKeepers {
		ctx.Logger().Info(fmt.Sprintf("Auction %s notifying module %s.", auction.ID, keeper.ModuleName()))
		keeper.OnAuctionWinnerSelected(ctx, auction.ID)
	}
}
