//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Wire      = "wire"  // 1 (base denom unit).
	MilliWire = "mwire" // 10^-3 (milli).
	MicroWire = "uwire" // 10^-6 (micro).
)

func init() {
	initNativeCoinUnits()
}

func initNativeCoinUnits() {
	_ = sdk.RegisterDenom(Wire, sdk.OneDec())
	_ = sdk.RegisterDenom(MilliWire, sdk.NewDecWithPrec(1, 3))
	_ = sdk.RegisterDenom(MicroWire, sdk.NewDecWithPrec(1, 6))
}
