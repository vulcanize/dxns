//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ID for bonds.
type ID string

// Bond represents funds deposited by an account for record rent payments.
type Bond struct {
	ID      ID        `json:"id,omitempty"`
	Owner   string    `json:"owner,omitempty"`
	Balance sdk.Coins `json:"balance"`
}

// BondID simplifies generation of bond IDs.
type BondID struct {
	Address  sdk.Address
	AccNum   uint64
	Sequence uint64
}

// Generate creates the bond ID.
func (bondID BondID) Generate() string {
	hasher := sha256.New()
	str := fmt.Sprintf("%s:%d:%d", bondID.Address.String(), bondID.AccNum, bondID.Sequence)
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
