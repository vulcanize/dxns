//
// Copyright 2019 Wireline, Inc.
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
	cdc.RegisterConcrete(MsgCreateBond{}, "bond/CreateBond", nil)
	cdc.RegisterConcrete(MsgRefillBond{}, "bond/RefillBond", nil)
	cdc.RegisterConcrete(MsgWithdrawBond{}, "bond/WithdrawBond", nil)
	cdc.RegisterConcrete(MsgCancelBond{}, "bond/CancelBond", nil)
}
