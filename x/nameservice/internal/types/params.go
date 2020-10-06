//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"bytes"
	"fmt"
	"time"

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
	// DefaultRecordRent is the default record rent for 1 time period (see expiry time).
	DefaultRecordRent string = "1000000uwire"

	// DefaultRecordExpiryTime is the default record expiry time (1 year).
	DefaultRecordExpiryTime time.Duration = time.Hour * 24 * 365

	DefaultAuthorityRent        string        = "10000000uwire"
	DefaultAuthorityExpiryTime  time.Duration = time.Hour * 24 * 365
	DefaultAuthorityGracePeriod time.Duration = time.Hour * 24 * 2

	DefaultAuthorityAuctionEnabled               = false
	DefaultCommitsDuration         time.Duration = time.Hour * 24
	DefaultRevealsDuration         time.Duration = time.Hour * 24
	DefaultCommitFee               string        = "1000000uwire"
	DefaultRevealFee               string        = "1000000uwire"
	DefaultMinimumBid              string        = "5000000uwire"
)

// Keys for parameter access
var (
	KeyRecordRent         = []byte("RecordRent")
	KeyRecordRentDuration = []byte("RecordRentDuration")

	KeyAuthorityRent         = []byte("AuthorityRent")
	KeyAuthorityRentDuration = []byte("AuthorityRentDuration")
	KeyAuthorityGracePeriod  = []byte("AuthorityGracePeriod")

	KeyAuthorityAuctionEnabled = []byte("AuthorityAuctionEnabled")
	KeyCommitsDuration         = []byte("AuthorityAuctionCommitsDuration")
	KeyRevealsDuration         = []byte("AuthorityAuctionRevealsDuration")
	KeyCommitFee               = []byte("AuthorityAuctionCommitFee")
	KeyRevealFee               = []byte("AuthorityAuctionRevealFee")
	KeyMinimumBid              = []byte("AuthorityAuctionMinimumBid")
)

var _ subspace.ParamSet = &Params{}

// Params defines the high level settings for the nameservice module.
type Params struct {
	RecordRent         string        `json:"record_rent" yaml:"record_rent"`
	RecordRentDuration time.Duration `json:"record_rent_duration" yaml:"record_rent_duration"`

	AuthorityRent         string        `json:"authority_rent" yaml:"authority_rent"`
	AuthorityRentDuration time.Duration `json:"authority_rent_duration" yaml:"authority_rent_duration"`
	AuthorityGracePeriod  time.Duration `json:"authority_grace_period" yaml:"authority_grace_period"`

	// Are name auctions enabled?
	AuthorityAuctionEnabled bool          `json:"authority_auction_enabled" yaml:"authority_auction_enabled"`
	CommitsDuration         time.Duration `json:"authority_auction_commits_duration" yaml:"authority_auction_commits_duration"`
	RevealsDuration         time.Duration `json:"authority_auction_reveals_duration" yaml:"authority_auction_reveals_duration"`
	CommitFee               string        `json:"authority_auction_commit_fee" yaml:"authority_auction_commit_fee"`
	RevealFee               string        `json:"authority_auction_reveal_fee" yaml:"authority_auction_reveal_fee"`
	MinimumBid              string        `json:"authority_auction_minimum_bid" yaml:"authority_auction_minimum_bid"`
}

// NewParams creates a new Params instance
func NewParams(recordRent string, recordRentDuration time.Duration,
	authorityRent string, authorityRentDuration time.Duration, authorityGracePeriod time.Duration,
	authorityAuctionEnabled bool, commitsDuration time.Duration, revealsDuration time.Duration,
	commitFee string, revealFee string, minimumBid string) Params {

	return Params{
		RecordRent:         recordRent,
		RecordRentDuration: recordRentDuration,

		AuthorityRent:         authorityRent,
		AuthorityRentDuration: authorityRentDuration,
		AuthorityGracePeriod:  authorityGracePeriod,

		AuthorityAuctionEnabled: authorityAuctionEnabled,
		CommitsDuration:         commitsDuration,
		RevealsDuration:         revealsDuration,
		CommitFee:               commitFee,
		RevealFee:               revealFee,
		MinimumBid:              minimumBid,
	}
}

// ParamKeyTable - ParamTable for nameservice module.
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs - implements params.ParamSet
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		params.NewParamSetPair(KeyRecordRent, &p.RecordRent, validateRecordRent),
		params.NewParamSetPair(KeyRecordRentDuration, &p.RecordRentDuration, validateRecordRentDuration),

		params.NewParamSetPair(KeyAuthorityRent, &p.AuthorityRent, validateAuthorityRent),
		params.NewParamSetPair(KeyAuthorityRentDuration, &p.AuthorityRentDuration, validateAuthorityRentDuration),
		params.NewParamSetPair(KeyAuthorityGracePeriod, &p.AuthorityGracePeriod, validateAuthorityGracePeriod),

		params.NewParamSetPair(KeyAuthorityAuctionEnabled, &p.AuthorityAuctionEnabled, validateAuthorityAuctionEnabled),
		params.NewParamSetPair(KeyCommitsDuration, &p.CommitsDuration, validateCommitsDuration),
		params.NewParamSetPair(KeyRevealsDuration, &p.RevealsDuration, validateRevealsDuration),
		params.NewParamSetPair(KeyCommitFee, &p.CommitFee, validateCommitFee),
		params.NewParamSetPair(KeyRevealFee, &p.RevealFee, validateRevealFee),
		params.NewParamSetPair(KeyMinimumBid, &p.MinimumBid, validateMinimumBid),
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
	return NewParams(DefaultRecordRent, DefaultRecordExpiryTime,
		DefaultAuthorityRent, DefaultAuthorityExpiryTime, DefaultAuthorityGracePeriod,
		DefaultAuthorityAuctionEnabled, DefaultCommitsDuration, DefaultRevealsDuration,
		DefaultCommitFee, DefaultRevealFee, DefaultMinimumBid,
	)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Record Rent                     : %v
  Record Rent Duration            : %v

  Authority Rent                  : %v
  Authority Rent Duration         : %v
  Authority Grace Period          : %v

  Authority Auction Enabled          : %v
  Authority Auction Commits Duration : %v
  Authority Auction Reveals Duration : %v
  Authority Auction Commit Fee       : %v
  Authority Auction Reveal Fee       : %v
  Authority Auction Minimum Bid      : %v`,
		p.RecordRent, p.RecordRentDuration,
		p.AuthorityRent, p.AuthorityRentDuration, p.AuthorityGracePeriod,
		p.AuthorityAuctionEnabled, p.CommitsDuration, p.RevealsDuration, p.CommitFee, p.RevealFee, p.MinimumBid)
}

func validateAmount(name string, i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", name, i)
	}

	if v == "" {
		return fmt.Errorf("%s can't be an empty string", name)
	}

	amount, err := sdk.ParseCoins(v)
	if err != nil {
		return err
	}

	if amount.IsAnyNegative() {
		return fmt.Errorf("%s can't be negative", name)
	}

	return nil
}

func validateDuration(name string, i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", name, i)
	}

	if v <= 0 {
		return fmt.Errorf("%s must be a positive integer", name)
	}

	return nil
}

func validateRecordRent(i interface{}) error {
	return validateAmount("RecordRent", i)
}

func validateRecordRentDuration(i interface{}) error {
	return validateDuration("RecordRentDuration", i)
}

func validateAuthorityRent(i interface{}) error {
	return validateAmount("AuthorityRent", i)
}

func validateAuthorityRentDuration(i interface{}) error {
	return validateDuration("AuthorityRentDuration", i)
}

func validateAuthorityGracePeriod(i interface{}) error {
	return validateDuration("AuthorityGracePeriod", i)
}

func validateAuthorityAuctionEnabled(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("%s invalid parameter type: %T", "AuthorityAuctionEnabled", i)
	}

	return nil
}

func validateCommitsDuration(i interface{}) error {
	return validateDuration("CommitsDuration", i)
}

func validateRevealsDuration(i interface{}) error {
	return validateDuration("RevealsDuration", i)
}

func validateCommitFee(i interface{}) error {
	return validateAmount("CommitFee", i)
}

func validateRevealFee(i interface{}) error {
	return validateAmount("RevealFee", i)
}

func validateMinimumBid(i interface{}) error {
	return validateAmount("MinimumBid", i)
}

// Validate a set of params.
func (p Params) Validate() error {
	if err := validateRecordRent(p.RecordRent); err != nil {
		return err
	}

	if err := validateRecordRentDuration(p.RecordRentDuration); err != nil {
		return err
	}

	if err := validateAuthorityRent(p.AuthorityRent); err != nil {
		return err
	}

	if err := validateAuthorityRentDuration(p.AuthorityRentDuration); err != nil {
		return err
	}

	if err := validateAuthorityGracePeriod(p.AuthorityGracePeriod); err != nil {
		return err
	}

	if err := validateAuthorityAuctionEnabled(p.AuthorityAuctionEnabled); err != nil {
		return err
	}

	if err := validateCommitsDuration(p.CommitsDuration); err != nil {
		return err
	}

	if err := validateRevealsDuration(p.RevealsDuration); err != nil {
		return err
	}

	if err := validateCommitFee(p.CommitFee); err != nil {
		return err
	}

	if err := validateRevealFee(p.RevealFee); err != nil {
		return err
	}

	if err := validateMinimumBid(p.MinimumBid); err != nil {
		return err
	}

	return nil
}
