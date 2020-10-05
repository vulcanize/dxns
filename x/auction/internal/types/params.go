//
// Copyright 2020 Wireline, Inc.
//

package types

import (
	"bytes"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Default parameter namespace.
const (
	DefaultParamspace = ModuleName
)

var _ subspace.ParamSet = (*Params)(nil)

// Params defines the parameters for the auction module.
type Params struct {
	// Duration of commits phase in seconds.
	CommitsDuration time.Duration `json:"commits_duration"`

	// Duration of reveals phase in seconds.
	RevealsDuration time.Duration `json:"reveals_duration"`

	// Commit and reveal fees.
	CommitFee sdk.Coin `json:"commit_fee"`
	RevealFee sdk.Coin `json:"reveal_fee"`

	MinimumBid sdk.Coin `json:"minimum_bid"`
}

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{}
}

// ParamKeyTable - ParamTable for bond module.
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs - implements params.ParamSet
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{}
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	return sb.String()
}

// Validate a set of params.
func (p Params) Validate() error {
	return nil
}
