//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Default parameter namespace.
const (
	DefaultParamspace = ModuleName
)

// Default parameter values.
const (
	DefaultMaxBondAmount string = "100000000000stake"
)

// Keys for parameter access
var (
	KeyMaxBondAmount = []byte("MaxBondAmount")
)

var _ subspace.ParamSet = &Params{}

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

// ParamKeyTable - ParamTable for bond module.
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs - implements params.ParamSet
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		params.NewParamSetPair(KeyMaxBondAmount, &p.MaxBondAmount, validateMaxBondAmount),
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxBondAmount: DefaultMaxBondAmount,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("Max Bond Amount: %s\n", p.MaxBondAmount))
	return sb.String()
}

func validateMaxBondAmount(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	amount, err := sdk.ParseCoins(v)
	if err != nil {
		return err
	}

	if amount.IsAnyNegative() {
		return errors.New("max bond amount must be positive")
	}

	return nil
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateMaxBondAmount(p.MaxBondAmount); err != nil {
		return err
	}

	return nil
}
