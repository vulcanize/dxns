//
// Copyright 2019 Wireline, Inc.
//

package gql

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/wirelineio/dxns/x/auction"
	"github.com/wirelineio/dxns/x/bond"
	"github.com/wirelineio/dxns/x/nameservice"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OwnerAttributeName denotes the owner attribute name for a bond.
const OwnerAttributeName = "owner"

// BondIDAttributeName denotes the record bond ID.
const BondIDAttributeName = "bondId"

// ExpiryTimeAttributeName denotes the record expiry time.
const ExpiryTimeAttributeName = "expiryTime"

func GetGQLRecord(ctx context.Context, resolver QueryResolver, record *nameservice.Record) (*Record, error) {
	// Nil record.
	if record == nil || record.Deleted {
		return nil, nil
	}

	attributes, err := getAttributes(record)
	if err != nil {
		return nil, err
	}

	references, err := getReferences(ctx, resolver, record)
	if err != nil {
		return nil, err
	}

	return &Record{
		ID:         string(record.ID),
		Names:      record.Names,
		BondID:     record.GetBondID(),
		CreateTime: record.GetCreateTime(),
		ExpiryTime: record.GetExpiryTime(),
		Owners:     record.GetOwners(),
		Attributes: attributes,
		References: references,
	}, nil
}

func GetGQLNameRecord(ctx context.Context, resolver QueryResolver, record *nameservice.NameRecord) (*NameRecord, error) {
	if record == nil {
		return nil, nil
	}

	records := make([]*NameRecordEntry, len(record.History))
	for index, entry := range record.History {
		records[index] = getNameRecordEntry(entry)
	}

	return &NameRecord{
		Latest:  getNameRecordEntry(record.NameRecordEntry),
		History: records,
	}, nil
}

func getNameRecordEntry(record nameservice.NameRecordEntry) *NameRecordEntry {
	return &NameRecordEntry{
		ID:     string(record.ID),
		Height: strconv.FormatInt(record.Height, 10),
	}
}

func GetGQLNameAuthorityRecord(ctx context.Context, resolver QueryResolver, record *nameservice.NameAuthority) (*AuthorityRecord, error) {
	if record == nil {
		return nil, nil
	}

	return &AuthorityRecord{
		OwnerAddress:   record.OwnerAddress,
		OwnerPublicKey: record.OwnerPublicKey,
		Height:         strconv.FormatInt(record.Height, 10),
		Status:         string(record.Status),
		BondID:         record.GetBondID(),
		ExpiryTime:     record.GetExpiryTime(),
	}, nil
}

func getAuctionBid(bid *auction.Bid) *AuctionBid {
	return &AuctionBid{
		BidderAddress: bid.BidderAddress,
		Status:        bid.Status,
		CommitHash:    bid.CommitHash,
		CommitTime:    bid.GetCommitTime(),
		RevealTime:    bid.GetRevealTime(),
		CommitFee:     getGQLCoin(bid.CommitFee),
		RevealFee:     getGQLCoin(bid.RevealFee),
		BidAmount:     getGQLCoin(bid.BidAmount),
	}
}

func GetGQLAuction(ctx context.Context, resolver QueryResolver, auction *auction.Auction, bids []*auction.Bid) (*Auction, error) {
	if auction == nil {
		return nil, nil
	}

	gqlAuction := Auction{
		ID:             string(auction.ID),
		Status:         auction.Status,
		OwnerAddress:   auction.OwnerAddress,
		CreateTime:     auction.GetCreateTime(),
		CommitsEndTime: auction.GetCommitsEndTime(),
		RevealsEndTime: auction.GetRevealsEndTime(),
		CommitFee:      getGQLCoin(auction.CommitFee),
		RevealFee:      getGQLCoin(auction.RevealFee),
		MinimumBid:     getGQLCoin(auction.MinimumBid),
		WinnerAddress:  auction.WinnerAddress,
		WinnerBid:      getGQLCoin(auction.WinnerBid),
		WinnerPrice:    getGQLCoin(auction.WinnerPrice),
	}

	auctionBids := make([]*AuctionBid, len(bids))
	for index, entry := range bids {
		auctionBids[index] = getAuctionBid(entry)
	}

	gqlAuction.Bids = auctionBids

	return &gqlAuction, nil
}

func getReferences(ctx context.Context, resolver QueryResolver, r *nameservice.Record) ([]*Record, error) {
	var ids []string

	for _, value := range r.Attributes {
		switch value.(type) {
		case interface{}:
			if obj, ok := value.(map[string]interface{}); ok {
				if _, ok := obj["/"]; ok && len(obj) == 1 {
					if _, ok := obj["/"].(string); ok {
						ids = append(ids, obj["/"].(string))
					}
				}
			}
		}
	}

	return resolver.GetRecordsByIds(ctx, ids)
}

func getAttributes(r *nameservice.Record) ([]*KeyValue, error) {
	return mapToKeyValuePairs(r.Attributes)
}

func mapToKeyValuePairs(attrs map[string]interface{}) ([]*KeyValue, error) {
	kvPairs := []*KeyValue{}

	trueVal := true
	falseVal := false

	for key, value := range attrs {

		kvPair := &KeyValue{
			Key: key,
		}

		switch val := value.(type) {
		case nil:
			kvPair.Value.Null = &trueVal
		case int:
			kvPair.Value.Int = &val
		case float64:
			kvPair.Value.Float = &val
		case string:
			kvPair.Value.String = &val
		case bool:
			kvPair.Value.Boolean = &val
		case interface{}:
			if obj, ok := value.(map[string]interface{}); ok {
				if _, ok := obj["/"]; ok && len(obj) == 1 {
					if _, ok := obj["/"].(string); ok {
						kvPair.Value.Reference = &Reference{
							ID: obj["/"].(string),
						}
					}
				} else {
					bytes, err := json.Marshal(obj)
					if err != nil {
						return nil, err
					}

					jsonStr := string(bytes)
					kvPair.Value.JSON = &jsonStr
				}
			}
		}

		if kvPair.Value.Null == nil {
			kvPair.Value.Null = &falseVal
		}

		valueType := reflect.ValueOf(value)
		if valueType.Kind() == reflect.Slice {
			bytes, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			jsonStr := string(bytes)
			kvPair.Value.JSON = &jsonStr
		}

		kvPairs = append(kvPairs, kvPair)
	}

	return kvPairs, nil
}

func matchOnRecordField(record *nameservice.Record, attr *KeyValueInput) (fieldFound bool, matched bool) {
	fieldFound = false
	matched = true

	switch attr.Key {
	case BondIDAttributeName:
		{
			fieldFound = true
			if attr.Value.String == nil || record.GetBondID() != *attr.Value.String {
				matched = false
				return
			}
		}
	case ExpiryTimeAttributeName:
		{
			fieldFound = true
			if attr.Value.String == nil || record.GetExpiryTime() != *attr.Value.String {
				matched = false
				return
			}
		}
	}

	return
}

func MatchOnAttributes(record *nameservice.Record, attributes []*KeyValueInput, all bool) bool {
	// Filter deleted records.
	if record.Deleted {
		return false
	}

	// If ONLY named records are requested, check for that condition first.
	if !all && len(record.Names) == 0 {
		return false
	}

	recAttrs := record.Attributes

	for _, attr := range attributes {
		// First try matching on record struct fields.
		fieldFound, matched := matchOnRecordField(record, attr)
		if fieldFound {
			if !matched {
				return false
			}

			continue
		}

		recAttrVal, recAttrFound := recAttrs[attr.Key]
		if !recAttrFound {
			return false
		}

		if attr.Value.Int != nil {
			recAttrValInt, ok := recAttrVal.(int)
			if !ok || *attr.Value.Int != recAttrValInt {
				return false
			}
		}

		if attr.Value.Float != nil {
			recAttrValFloat, ok := recAttrVal.(float64)
			if !ok || *attr.Value.Float != recAttrValFloat {
				return false
			}
		}

		if attr.Value.String != nil {
			recAttrValString, ok := recAttrVal.(string)
			if !ok {
				return false
			}

			if *attr.Value.String != recAttrValString {
				return false
			}
		}

		if attr.Value.Boolean != nil {
			recAttrValBool, ok := recAttrVal.(bool)
			if !ok || *attr.Value.Boolean != recAttrValBool {
				return false
			}
		}

		if attr.Value.Reference != nil {
			obj, ok := recAttrVal.(map[string]interface{})
			if !ok {
				// Attr value is not an object.
				return false
			}

			if _, ok := obj["/"].(string); !ok {
				// Attr value is not a reference.
				return false
			}

			recAttrValRefID := obj["/"].(string)
			if recAttrValRefID != attr.Value.Reference.ID {
				return false
			}
		}

		// TODO(ashwin): Handle arrays.
	}

	return true
}

func getGQLCoin(coin sdk.Coin) *Coin {
	gqlCoin := Coin{
		Type:     coin.Denom,
		Quantity: strconv.FormatInt(coin.Amount.Int64(), 10),
	}

	return &gqlCoin
}

func getGQLCoins(coins sdk.Coins) []*Coin {
	gqlCoins := make([]*Coin, len(coins))
	for index, coin := range coins {
		gqlCoins[index] = getGQLCoin(coin)
	}

	return gqlCoins
}

func getGQLBond(ctx context.Context, resolver *queryResolver, bondObj *bond.Bond) (*Bond, error) {
	// Nil record.
	if bondObj == nil {
		return nil, nil
	}

	return &Bond{
		ID:      string(bondObj.ID),
		Owner:   bondObj.Owner,
		Balance: getGQLCoins(bondObj.Balance),
	}, nil
}

func matchBondOnAttributes(bondObj *bond.Bond, attributes []*KeyValueInput) bool {
	for _, attr := range attributes {
		switch attr.Key {
		case OwnerAttributeName:
			{
				if attr.Value.String == nil || bondObj.Owner != *attr.Value.String {
					return false
				}
			}
		default:
			{
				// Only attributes explicitly listed in the switch are queryable.
				return false
			}
		}
	}

	return true
}
