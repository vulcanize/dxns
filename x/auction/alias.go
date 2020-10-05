//
// Copyright 2020 Wireline, Inc.
//

package auction

import (
	"github.com/wirelineio/dxns/x/auction/internal/keeper"
	"github.com/wirelineio/dxns/x/auction/internal/types"
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey

	AuctionBurnModuleAccountName = types.AuctionBurnModuleAccountName
	AuctionStatusCompleted       = types.AuctionStatusCompleted
)

var (
	DefaultParamspace = types.DefaultParamspace
	NewKeeper         = keeper.NewKeeper
	NewQuerier        = keeper.NewQuerier
	ModuleCdc         = types.ModuleCdc
	RegisterCodec     = types.RegisterCodec

	RegisterInvariants = keeper.RegisterInvariants

	NewMsgCreateAuction = types.NewMsgCreateAuction

	PrefixIDToAuctionIndex     = keeper.PrefixIDToAuctionIndex
	PrefixAuctionBidsIndex     = keeper.PrefixAuctionBidsIndex
	GetAuctionIndexKey         = keeper.GetAuctionIndexKey
	GetAuctionBidsIndexPrefix  = keeper.GetAuctionBidsIndexPrefix
	GetBidIndexKey             = keeper.GetBidIndexKey
	GetOwnerToAuctionsIndexKey = keeper.GetOwnerToAuctionsIndexKey

	GetAuction = keeper.GetAuction
	GetBids    = keeper.GetBids
)

type (
	ID      = types.ID
	Auction = types.Auction
	Bid     = types.Bid

	// Used for block changeset.
	AuctionBidInfo = types.AuctionBidInfo

	Keeper              = keeper.Keeper
	AuctionUsageKeeper  = types.AuctionUsageKeeper
	AuctionClientKeeper = keeper.AuctionClientKeeper
	Params              = types.Params
)
