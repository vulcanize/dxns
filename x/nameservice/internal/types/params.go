//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Nameservice params default values.
const (
	// DefaultRecordRent is the default record rent for 1 time period (see expiry time).
	DefaultRecordRent string = "1000000uwire"

	// DefaultRecordExpiryTime is the default record expiry time (1 year).
	DefaultRecordExpiryTime time.Duration = time.Hour * 24 * 365

	DefaultAuthorityRent        string        = "10000000uwire"
	DefaultAuthorityExpiryTime  time.Duration = time.Hour * 24 * 365
	DefaultAuthorityGracePeriod time.Duration = time.Hour * 24 * 2

	DefaultAuthorityAuctionsEnabled               = true
	DefaultCommitsDuration          time.Duration = time.Hour * 24
	DefaultRevealsDuration          time.Duration = time.Hour * 24
	DefaultCommitFee                string        = "1000000uwire"
	DefaultRevealFee                string        = "1000000uwire"
	DefaultMinimumBid               string        = "5000000uwire"
)

// nolint - Keys for parameter access
var (
	KeyRecordRent         = []byte("RecordRent")
	KeyRecordRentDuration = []byte("RecordRentDuration")

	KeyAuthorityRent         = []byte("AuthorityRent")
	KeyAuthorityRentDuration = []byte("AuthorityRentDuration")
	KeyAuthorityGracePeriod  = []byte("AuthorityGracePeriod")

	KeyAuthorityAuctions = []byte("AuthorityAuctionEnabled")
	KeyCommitsDuration   = []byte("AuthorityAuctionCommitsDuration")
	KeyRevealsDuration   = []byte("AuthorityAuctionRevealsDuration")
	KeyCommitFee         = []byte("AuthorityAuctionCommitFee")
	KeyRevealFee         = []byte("AuthorityAuctionRevealFee")
	KeyMinimumBid        = []byte("AuthorityAuctionMinimumBid")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for nameservice
type Params struct {
	RecordRent         string        `json:"record_rent" yaml:"record_rent"`
	RecordRentDuration time.Duration `json:"record_rent_duration" yaml:"record_rent_duration"`

	AuthorityRent         string        `json:"authority_rent" yaml:"authority_rent"`
	AuthorityRentDuration time.Duration `json:"authority_rent_duration" yaml:"authority_rent_duration"`
	AuthorityGracePeriod  time.Duration `json:"authority_grace_period" yaml:"authority_grace_period"`

	// Are name auctions enabled?
	AuthorityAuctions bool          `json:"authority_auctions" yaml:"name_auctions"`
	CommitsDuration   time.Duration `json:"authority_auction_commits_duration" yaml:"authority_auction_commits_duration"`
	RevealsDuration   time.Duration `json:"authority_auction_reveals_duration" yaml:"authority_auction_reveals_duration"`
	CommitFee         string        `json:"authority_auction_commit_fee" yaml:"authority_auction_commit_fee"`
	RevealFee         string        `json:"authority_auction_reveal_fee" yaml:"authority_auction_reveal_fee"`
	MinimumBid        string        `json:"authority_auction_minimum_bid" yaml:"authority_auction_minimum_bid"`
}

// NewParams creates a new Params instance
func NewParams(recordRent string, recordRentDuration time.Duration,
	authorityRent string, authorityRentDuration time.Duration, authorityGracePeriod time.Duration,
	authorityAuctions bool, commitsDuration time.Duration, revealsDuration time.Duration,
	commitFee string, revealFee string, minimumBid string) Params {

	return Params{
		RecordRent:         recordRent,
		RecordRentDuration: recordRentDuration,

		AuthorityRent:         authorityRent,
		AuthorityRentDuration: authorityRentDuration,
		AuthorityGracePeriod:  authorityGracePeriod,

		AuthorityAuctions: authorityAuctions,
		CommitsDuration:   commitsDuration,
		RevealsDuration:   revealsDuration,
		CommitFee:         commitFee,
		RevealFee:         revealFee,
		MinimumBid:        minimumBid,
	}
}

// ParamSetPairs - implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyRecordRent, Value: &p.RecordRent},
		{Key: KeyRecordRentDuration, Value: &p.RecordRentDuration},

		{Key: KeyAuthorityRent, Value: &p.AuthorityRent},
		{Key: KeyAuthorityRentDuration, Value: &p.AuthorityRentDuration},
		{Key: KeyAuthorityGracePeriod, Value: &p.AuthorityGracePeriod},

		{Key: KeyAuthorityAuctions, Value: &p.AuthorityAuctions},
		{Key: KeyCommitsDuration, Value: &p.CommitsDuration},
		{Key: KeyRevealsDuration, Value: &p.RevealsDuration},
		{Key: KeyCommitFee, Value: &p.CommitFee},
		{Key: KeyRevealFee, Value: &p.RevealFee},
		{Key: KeyMinimumBid, Value: &p.MinimumBid},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(DefaultRecordRent, DefaultRecordExpiryTime,
		DefaultAuthorityRent, DefaultAuthorityExpiryTime, DefaultAuthorityGracePeriod,
		DefaultAuthorityAuctionsEnabled, DefaultCommitsDuration, DefaultRevealsDuration,
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
		p.AuthorityAuctions, p.CommitsDuration, p.RevealsDuration, p.CommitFee, p.RevealFee, p.MinimumBid)
}

// Validate a set of params.
func (p Params) Validate() error {
	if p.RecordRent == "" {
		return fmt.Errorf("nameservice parameter RecordRent can't be an empty string")
	}

	if p.RecordRentDuration <= 0 {
		return fmt.Errorf("nameservice parameter RecordRentDuration must be a positive integer")
	}

	if p.AuthorityRent == "" {
		return fmt.Errorf("nameservice parameter AuthorityRent can't be an empty string")
	}

	if p.AuthorityRentDuration <= 0 {
		return fmt.Errorf("nameservice parameter AuthorityRentDuration must be a positive integer")
	}

	if p.AuthorityGracePeriod <= 0 {
		return fmt.Errorf("nameservice parameter AuthorityGracePeriod must be a positive integer")
	}

	if p.CommitsDuration <= 0 {
		return fmt.Errorf("nameservice parameter CommitsDuration must be a positive integer")
	}

	if p.RevealsDuration <= 0 {
		return fmt.Errorf("nameservice parameter RevealsDuration must be a positive integer")
	}

	if p.CommitFee == "" {
		return fmt.Errorf("nameservice parameter CommitFee can't be an empty string")
	}

	if p.RevealFee == "" {
		return fmt.Errorf("nameservice parameter RevealFee can't be an empty string")
	}

	if p.MinimumBid == "" {
		return fmt.Errorf("nameservice parameter MinimumBid can't be an empty string")
	}

	return nil
}
