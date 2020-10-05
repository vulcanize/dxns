//
// Copyright 2020 Wireline, Inc.
//

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc is the codec for the module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateAuction{}, "auction/CreateAuction", nil)
	cdc.RegisterConcrete(MsgCommitBid{}, "auction/CommitBid", nil)
	cdc.RegisterConcrete(MsgRevealBid{}, "auction/RevealBid", nil)
}
