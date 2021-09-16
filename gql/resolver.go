//
// Copyright 2019 Wireline, Inc.
//

package gql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/vulcanize/dxns/x/auction"
	"github.com/vulcanize/dxns/x/bond"
	"github.com/vulcanize/dxns/x/nameservice"
)

// DefaultLogNumLines is the number of log lines to tail by default.
const DefaultLogNumLines = 50

// MaxLogNumLines is the max number of log lines that can be tailed.
const MaxLogNumLines = 1000

// Resolver is the GQL query resolver.
type Resolver struct {
	baseApp       *bam.BaseApp
	codec         *codec.Codec
	keeper        nameservice.Keeper
	bondKeeper    bond.Keeper
	accountKeeper auth.AccountKeeper
	auctionKeeper auction.Keeper
	logFile       string
}

// Mutation is the entry point to tx execution.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query is the entry point to query execution.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) InsertRecord(ctx context.Context, attributes []*KeyValueInput) (*Record, error) {
	// Only supported by mock server.
	return nil, errors.New("not implemented")
}

func (r *mutationResolver) Submit(ctx context.Context, tx string) (*string, error) {
	stdTx, err := decodeStdTx(r.codec, tx)
	if err != nil {
		return nil, err
	}

	r.baseApp.Logger().Info(string(r.codec.MustMarshalJSON(stdTx)))

	res, err := broadcastTx(r, stdTx)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.MarshalIndent(res, "", "  ")
	jsonResponse := string(jsonBytes)

	return &jsonResponse, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) GetRecordsByIds(ctx context.Context, ids []string) ([]*Record, error) {
	records := make([]*Record, len(ids))
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
func (r *queryResolver) QueryRecords(ctx context.Context, attributes []*KeyValueInput, all *bool) ([]*Record, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})

	var records = r.keeper.MatchRecords(sdkContext, func(record *nameservice.Record) bool {
		return MatchOnAttributes(record, attributes, (all != nil && *all))
	})

	return QueryRecords(ctx, r, records, attributes)
}

// QueryRecords filters records by K=V conditions.
func QueryRecords(ctx context.Context, resolver QueryResolver, records []*nameservice.Record, attributes []*KeyValueInput) ([]*Record, error) {
	gqlResponse := []*Record{}

	for _, record := range records {
		gqlRecord, err := GetGQLRecord(ctx, resolver, record)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	return gqlResponse, nil
}

// ResolveNames resolves records by name/WRN.
func (r *queryResolver) ResolveNames(ctx context.Context, names []string) (*RecordResult, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})
	gqlResponse := []*Record{}

	for _, name := range names {
		record := r.keeper.ResolveWRN(sdkContext, name)
		gqlRecord, err := GetGQLRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := RecordResult{
		Meta: &ResultMeta{
			Height: strconv.FormatInt(r.baseApp.LastBlockHeight(), 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

func (r *queryResolver) LookupAuthorities(ctx context.Context, names []string) (*AuthorityResult, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})
	gqlResponse := []*AuthorityRecord{}

	for _, name := range names {
		record := r.keeper.GetNameAuthority(sdkContext, name)

		gqlRecord, err := GetGQLNameAuthorityRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		if record != nil && record.AuctionID != "" {
			auction := r.auctionKeeper.GetAuction(sdkContext, record.AuctionID)
			bids := r.auctionKeeper.GetBids(sdkContext, auction.ID)

			gqlAuction, err := GetGQLAuction(ctx, r, auction, bids)
			if err != nil {
				return nil, err
			}

			gqlRecord.Auction = gqlAuction
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := AuthorityResult{
		Meta: &ResultMeta{
			Height: strconv.FormatInt(r.baseApp.LastBlockHeight(), 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

func (r *queryResolver) LookupNames(ctx context.Context, names []string) (*NameResult, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})
	gqlResponse := []*NameRecord{}

	for _, name := range names {
		record := r.keeper.GetNameRecord(sdkContext, name)
		gqlRecord, err := GetGQLNameRecord(ctx, r, record)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlRecord)
	}

	result := NameResult{
		Meta: &ResultMeta{
			Height: strconv.FormatInt(r.baseApp.LastBlockHeight(), 10),
		},
		Records: gqlResponse,
	}

	return &result, nil
}

// GetLogs tails the log file.
func GetLogs(ctx context.Context, logFile string, count *int) ([]string, error) {
	if logFile == "" {
		return []string{}, nil
	}

	numLines := DefaultLogNumLines
	if count != nil {
		// Lower bound check.
		if *count > 0 {
			numLines = *count
		}

		// Upper bound check.
		if *count > MaxLogNumLines {
			numLines = MaxLogNumLines
		}
	}

	bytes, err := exec.Command("tail", fmt.Sprintf("-%d", numLines), logFile).Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSuffix(string(bytes), "\n"), "\n"), nil
}

func (r *queryResolver) GetLogs(ctx context.Context, count *int) ([]string, error) {
	return GetLogs(ctx, r.logFile, count)
}

func (r *queryResolver) GetStatus(ctx context.Context) (*Status, error) {
	rpcContext := &rpctypes.Context{}

	nodeInfo, syncInfo, validatorInfo, err := getStatusInfo(rpcContext)
	if err != nil {
		return nil, err
	}

	numPeers, peers, err := getNetInfo(rpcContext)
	if err != nil {
		return nil, err
	}

	validatorSet, err := getValidatorSet(rpcContext)
	if err != nil {
		return nil, err
	}

	diskUsage, err := GetDiskUsage(NodeDataPath)
	if err != nil {
		return nil, err
	}

	return &Status{
		Version:    NamserviceVersion,
		Node:       nodeInfo,
		Sync:       syncInfo,
		Validator:  validatorInfo,
		Validators: validatorSet,
		NumPeers:   numPeers,
		Peers:      peers,
		DiskUsage:  diskUsage,
	}, nil
}

func (r *queryResolver) GetAccounts(ctx context.Context, addresses []string) ([]*Account, error) {
	accounts := make([]*Account, len(addresses))
	for index, address := range addresses {
		account, err := r.GetAccount(ctx, address)
		if err != nil {
			return nil, err
		}

		accounts[index] = account
	}

	return accounts, nil
}

func (r *queryResolver) GetAccount(ctx context.Context, address string) (*Account, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	account := r.accountKeeper.GetAccount(sdkContext, addr)
	if account == nil {
		return nil, nil
	}

	var pubKey *string
	if account.GetPubKey() != nil {
		pubKeyStr := base64.StdEncoding.EncodeToString(account.GetPubKey().Bytes())
		pubKey = &pubKeyStr
	}

	accNum := strconv.FormatUint(account.GetAccountNumber(), 10)
	seq := strconv.FormatUint(account.GetSequence(), 10)

	return &Account{
		Address:  address,
		Number:   accNum,
		Sequence: seq,
		PubKey:   pubKey,
		Balance:  getGQLCoins(account.GetCoins()),
	}, nil
}

func (r *queryResolver) GetRecord(ctx context.Context, id string) (*Record, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})

	dbID := nameservice.ID(id)
	if r.keeper.HasRecord(sdkContext, dbID) {
		record := r.keeper.GetRecord(sdkContext, dbID)
		if !record.Deleted {
			return GetGQLRecord(ctx, r, &record)
		}
	}

	return nil, nil
}

func (r *queryResolver) GetBondsByIds(ctx context.Context, ids []string) ([]*Bond, error) {
	bonds := make([]*Bond, len(ids))
	for index, id := range ids {
		bondObj, err := r.GetBond(ctx, id)
		if err != nil {
			return nil, err
		}

		bonds[index] = bondObj
	}

	return bonds, nil
}

func (r *queryResolver) GetBond(ctx context.Context, id string) (*Bond, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})

	dbID := bond.ID(id)
	if r.bondKeeper.HasBond(sdkContext, dbID) {
		bondObj := r.bondKeeper.GetBond(sdkContext, dbID)
		return getGQLBond(ctx, r, &bondObj)
	}

	return nil, nil
}

func (r *queryResolver) QueryBonds(ctx context.Context, attributes []*KeyValueInput) ([]*Bond, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})
	gqlResponse := []*Bond{}

	var bonds = r.bondKeeper.MatchBonds(sdkContext, func(bondObj *bond.Bond) bool {
		return matchBondOnAttributes(bondObj, attributes)
	})

	for _, bondObj := range bonds {
		gqlBond, err := getGQLBond(ctx, r, bondObj)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlBond)
	}

	return gqlResponse, nil
}

func (r *queryResolver) GetAuctionsByIds(ctx context.Context, ids []string) ([]*Auction, error) {
	sdkContext := r.baseApp.NewContext(true, abci.Header{})
	gqlResponse := []*Auction{}

	for _, id := range ids {
		auctionObj := r.auctionKeeper.GetAuction(sdkContext, auction.ID(id))
		bids := r.auctionKeeper.GetBids(sdkContext, auction.ID(id))
		gqlAuction, err := GetGQLAuction(ctx, r, auctionObj, bids)
		if err != nil {
			return nil, err
		}

		gqlResponse = append(gqlResponse, gqlAuction)
	}

	return gqlResponse, nil
}
