//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// nolint - Keys for parameter access
var (
	KeyMaxBondAmount = []byte("MaxBondAmount")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for the bond module.
type Params struct {
	MaxBondAmount string `json:"max_bond_amount" yaml:"max_bond_amount"`
}

// NewParams creates a new Params instance
func NewParams(maxBondAmount string) Params {
	return Params{
		MaxBondAmount: maxBondAmount,
	}
}

// ParamSetPairs - implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyMaxBondAmount, Value: &p.MaxBondAmount},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams("")
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Max Bond Amount: %s`, p.MaxBondAmount)
}

// Validate a set of params.
func (p Params) Validate() error {
	_, err := sdk.ParseCoins(p.MaxBondAmount)
	if err != nil {
		return err
	}

	return nil
}
