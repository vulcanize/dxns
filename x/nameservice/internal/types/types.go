//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	"crypto/sha256"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	canonicalJson "github.com/gibson042/canonicaljson-go"
	"github.com/vulcanize/dxns/x/auction"
	"github.com/vulcanize/dxns/x/bond"
	"github.com/vulcanize/dxns/x/nameservice/internal/helpers"
)

type AutorityStatus string

const (
	AuthorityActive       AutorityStatus = "active"
	AuthorityExpired      AutorityStatus = "expired"
	AuthorityUnderAuction AutorityStatus = "auction"
)

// ID for records.
type ID string

// Record represents a WNS record.
type Record struct {
	ID         ID                     `json:"id,omitempty"`
	Names      []string               `json:"names,omitempty"`
	BondID     bond.ID                `json:"bondId,omitempty"`
	CreateTime time.Time              `json:"createTime,omitempty"`
	ExpiryTime time.Time              `json:"expiryTime,omitempty"`
	Deleted    bool                   `json:"deleted,omitempty"`
	Owners     []string               `json:"owners,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// GetBondID returns the BondID of the Record.
func (r Record) GetBondID() string {
	return string(r.BondID)
}

// GetExpiryTime returns the expiry time of the Record.
func (r Record) GetExpiryTime() string {
	return string(sdk.FormatTimeBytes(r.ExpiryTime))
}

// GetCreateTime returns the create time of the Record.
func (r Record) GetCreateTime() string {
	return string(sdk.FormatTimeBytes(r.CreateTime))
}

// GetOwners returns the list of owners (for GQL).
func (r Record) GetOwners() []*string {
	owners := []*string{}
	for _, owner := range r.Owners {
		// Note: Without this copy, it's return the same address for all elements in the slice.
		ownerCopy := string(owner)
		owners = append(owners, &ownerCopy)
	}

	return owners
}

// ToRecordObj converts Record to RecordObj.
// Why? Because go-amino can't handle maps: https://github.com/tendermint/go-amino/issues/4.
func (r *Record) ToRecordObj() RecordObj {
	var resourceObj RecordObj

	resourceObj.ID = r.ID
	resourceObj.BondID = r.BondID
	resourceObj.CreateTime = r.CreateTime
	resourceObj.ExpiryTime = r.ExpiryTime
	resourceObj.Deleted = r.Deleted
	resourceObj.Owners = r.Owners
	resourceObj.Attributes = helpers.MarshalMapToJSONBytes(r.Attributes)

	return resourceObj
}

// ToNameRecordEntry gets a naming record entry for the record.
func (r *Record) ToNameRecordEntry() NameRecordEntry {
	var nameRecordEntry NameRecordEntry
	nameRecordEntry.ID = r.ID

	return nameRecordEntry
}

// CanonicalJSON returns the canonical JSON respresentation of the record.
func (r *Record) CanonicalJSON() []byte {
	bytes, err := canonicalJson.Marshal(r.Attributes)
	if err != nil {
		panic("Record marshal error.")
	}

	return bytes
}

// GetSignBytes generates a record hash to be signed.
func (r *Record) GetSignBytes() ([]byte, []byte) {
	// Double SHA256 hash.

	// Input to the first round of hashing.
	bytes := r.CanonicalJSON()

	// First round.
	first := sha256.New()
	first.Write(bytes)
	firstHash := first.Sum(nil)

	// Second round of hashing takes as input the output of the first round.
	second := sha256.New()
	second.Write(firstHash)
	secondHash := second.Sum(nil)

	return secondHash, bytes
}

// GetCID gets the record CID.
func (r *Record) GetCID() (ID, error) {
	id, err := helpers.GetCid(r.CanonicalJSON())
	if err != nil {
		return "", err
	}

	return ID(id), nil
}

// HasExpired returns true if the record has expired.
func (r *Record) HasExpired(ctx sdk.Context) bool {
	return ctx.BlockTime().After(r.ExpiryTime)
}

// Signature represents a record signature.
type Signature struct {
	PubKey    string `json:"pubKey"`
	Signature string `json:"sig"`
}

// PayloadObj represents a signed record payload.
type PayloadObj struct {
	Record     RecordObj   `json:"record"`
	Signatures []Signature `json:"signatures"`
}

// ToPayload converts Payload to PayloadObj object.
// Why? Because go-amino can't handle maps: https://github.com/tendermint/go-amino/issues/4.
func (payloadObj PayloadObj) ToPayload() Payload {
	var payload Payload

	payload.Record = helpers.UnMarshalMapFromJSONBytes(payloadObj.Record.Attributes)
	payload.Signatures = payloadObj.Signatures

	return payload
}

// RecordObj represents a WNS record.
type RecordObj struct {
	ID         ID        `json:"id,omitempty"`
	BondID     bond.ID   `json:"bondId,omitempty"`
	CreateTime time.Time `json:"createTime,omitempty"`
	ExpiryTime time.Time `json:"expiryTime,omitempty"`
	Deleted    bool      `json:"deleted,omitempty"`
	Owners     []string  `json:"owners,omitempty"`
	Attributes []byte    `json:"attributes,omitempty"`
}

// ToRecord converts RecordObj to Record.
// Why? Because go-amino can't handle maps: https://github.com/tendermint/go-amino/issues/4.
func (resourceObj *RecordObj) ToRecord() Record {
	var record Record

	record.ID = resourceObj.ID
	record.BondID = resourceObj.BondID
	record.CreateTime = resourceObj.CreateTime
	record.ExpiryTime = resourceObj.ExpiryTime
	record.Deleted = resourceObj.Deleted
	record.Owners = resourceObj.Owners
	record.Attributes = helpers.UnMarshalMapFromJSONBytes(resourceObj.Attributes)

	return record
}

// Payload represents a signed record payload that can be serialized from/to YAML.
type Payload struct {
	Record     map[string]interface{} `json:"record"`
	Signatures []Signature            `json:"signatures"`
}

// ToPayloadObj converts Payload to PayloadObj object.
// Why? Because go-amino can't handle maps: https://github.com/tendermint/go-amino/issues/4.
func (payload *Payload) ToPayloadObj() PayloadObj {
	var payloadObj PayloadObj

	payloadObj.Record.Attributes = helpers.MarshalMapToJSONBytes(payload.Record)
	payloadObj.Signatures = payload.Signatures

	return payloadObj
}

// NameAuthority records the name/authority ownership info.
type NameAuthority struct {
	// Owner public key.
	OwnerPublicKey string `json:"ownerPublicKey"`

	// Owner address.
	OwnerAddress string `json:"ownerAddress"`

	// Block height at which name/authority was created.
	Height int64 `json:"height"`

	Status AutorityStatus `json:"status"`

	AuctionID auction.ID `json:"auctionID"`

	BondID bond.ID `json:"bondID"`

	ExpiryTime time.Time `json:"expiryTime,omitempty"`
}

func (authority NameAuthority) GetBondID() string {
	return string(authority.BondID)
}

func (authority NameAuthority) GetExpiryTime() string {
	return string(sdk.FormatTimeBytes(authority.ExpiryTime))
}

// NameRecordEntry is a naming record entry for a WRN.
type NameRecordEntry struct {
	// Record ID.
	ID ID `json:"id"`

	// Block height at which name record was created.
	Height int64 `json:"height"`
}

// NameRecord stores name mapping info for a WRN.
type NameRecord struct {
	NameRecordEntry `json:"latest"`

	// TODO(ashwin): Move to external indexer when available.
	History []NameRecordEntry `json:"history"`
}

// BlockChangeset is a changeset corresponding to a block.
type BlockChangeset struct {
	Height          int64                    `json:"height"`
	Records         []ID                     `json:"records"`
	Auctions        []auction.ID             `json:"auctions"`
	AuctionBids     []auction.AuctionBidInfo `json:"auctionBids"`
	NameAuthorities []string                 `json:"authorities"`
	Names           []string                 `json:"names"`
}
