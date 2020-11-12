//
// Copyright 2020 Wireline, Inc.
//

package gql

import (
	"context"
	"os"
	"strconv"

	"github.com/wirelineio/dxns/cmd/dxnsd-lite/sync"
	baseGql "github.com/wirelineio/dxns/gql"
	"github.com/wirelineio/dxns/x/auction"
	"github.com/wirelineio/dxns/x/nameservice"
)

// LiteNodeDataPath is the path to the lite node data folder.
const LiteNodeDataPath = sync.DefaultLightNodeHome + "/data"

// Resolver is the GQL query resolver.
type Resolver struct {
	PrimaryNode *sync.RPCNodeHandler
	Keeper      *sync.Keeper
	LogFile     string
}

type queryResolver struct{ *Resolver }

// Query is the entry point to query execution.
func (r *Resolver) Query() baseGql.QueryResolver {
	return &queryResolver{r}
}

func (r *queryResolver) GetRecordsByIds(ctx context.Context, ids []string) ([]*baseGql.Record, error) {
	records := make([]*baseGql.Record, len(ids))
	for index, id := range ids {
		record, err := r.GetRecord(ctx, id)
		if err != nil {
			return nil, err
		}

		records[index] = record
	}

	return records, nil
}

// QueryRecords filters records by K=V conditions.
func (r *queryResolver) QueryRecords(ctx context.Context, attributes []*baseGql.KeyValueInput, all *bool) ([]*baseGql.Record, error) {
	var records = r.Keeper.MatchRecords(func(record *nameservice.Record) bool {
		return baseGql.MatchOnAttributes(record, attributes, (all != nil && *all))
	})

	return baseGql.QueryRecords(ctx, r, records, attributes)
}

// ResolveRecords resolves records by ref/WRN, with semver range support.
func (r *queryResolver) ResolveNames(ctx context.Context, names []string) (*baseGql.RecordResult, error) {
	gqlResponse := []*baseGql.Record{}

	for _, name := range names {
		record := r.Keeper.ResolveWRN(name)
		gqlRecord, err := baseGql.GetGQLRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := baseGql.RecordResult{
		Meta: &baseGql.ResultMeta{
			Height: strconv.FormatInt(r.Keeper.GetStatusRecord().LastSyncedHeight, 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

func (r *queryResolver) LookupAuthorities(ctx context.Context, names []string) (*baseGql.AuthorityResult, error) {
	gqlResponse := []*baseGql.AuthorityRecord{}

	for _, name := range names {
		record := r.Keeper.GetNameAuthority(name)
		gqlRecord, err := baseGql.GetGQLNameAuthorityRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		if record != nil && record.AuctionID != "" {
			auction := r.Keeper.GetAuction(record.AuctionID)
			bids := r.Keeper.GetBids(auction.ID)

			gqlAuction, err := baseGql.GetGQLAuction(ctx, r, auction, bids)
			if err != nil {
				return nil, err
			}

			gqlRecord.Auction = gqlAuction
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := baseGql.AuthorityResult{
		Meta: &baseGql.ResultMeta{
			Height: strconv.FormatInt(r.Keeper.GetStatusRecord().LastSyncedHeight, 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

func (r *queryResolver) LookupNames(ctx context.Context, names []string) (*baseGql.NameResult, error) {
	gqlResponse := []*baseGql.NameRecord{}

	for _, name := range names {
		record := r.Keeper.GetNameRecord(name)
		gqlRecord, err := baseGql.GetGQLNameRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := baseGql.NameResult{
		Meta: &baseGql.ResultMeta{
			Height: strconv.FormatInt(r.Keeper.GetStatusRecord().LastSyncedHeight, 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

func (r *queryResolver) GetLogs(ctx context.Context, count *int) ([]string, error) {
	return baseGql.GetLogs(ctx, r.LogFile, count)
}

func (r *queryResolver) GetStatus(ctx context.Context) (*baseGql.Status, error) {
	statusRecord := r.Keeper.GetStatusRecord()

	diskUsage, err := baseGql.GetDiskUsage(os.ExpandEnv(LiteNodeDataPath))
	if err != nil {
		return nil, err
	}

	validators, err := r.PrimaryNode.Client.Validators(nil, 1, 100)
	if err != nil {
		return nil, err
	}

	return &baseGql.Status{
		Version:    baseGql.NamserviceVersion,
		Node:       &baseGql.NodeInfo{Network: r.Keeper.GetChainID()},
		Validators: baseGql.GetValidatorSet(validators),
		Sync: &baseGql.SyncInfo{
			LatestBlockHeight: strconv.FormatInt(statusRecord.LastSyncedHeight, 10),
			CatchingUp:        statusRecord.CatchingUp,
		},
		DiskUsage: diskUsage,
	}, nil
}

func (r *queryResolver) GetRecord(ctx context.Context, id string) (*baseGql.Record, error) {
	dbID := nameservice.ID(id)
	if r.Keeper.HasRecord(dbID) {
		record := r.Keeper.GetRecord(dbID)
		if !record.Deleted {
			return baseGql.GetGQLRecord(ctx, r, &record)
		}
	}

	return nil, nil
}

func (r *queryResolver) GetAuctionsByIds(ctx context.Context, ids []string) ([]*baseGql.Auction, error) {
	gqlResponse := []*baseGql.Auction{}

	for _, id := range ids {
		auctionObj := r.Keeper.GetAuction(auction.ID(id))
		bids := r.Keeper.GetBids(auction.ID(id))
		gqlAuction, err := baseGql.GetGQLAuction(ctx, r, auctionObj, bids)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlAuction)
	}

	return gqlResponse, nil
}
