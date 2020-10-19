package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MakeCodec registers the necessary types and interfaces for an sdk.App. This
// codec is provided to all the modules the application depends on.
//
// NOTE: This codec will be deprecated in favor of AppCodec once all modules are
// migrated to protobuf.
func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc.Seal()
}

// MakeCodecLite is the codec used by the lite node.
func MakeCodecLite() *codec.Codec {
	return MakeCodec()
}
