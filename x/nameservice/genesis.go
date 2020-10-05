//
// Copyright 2019 Wireline, Inc.
//

package nameservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/nameservice/internal/types"
)

type AuthorityEntry struct {
	Name  string              `json:"name" yaml:"name"`
	Entry types.NameAuthority `json:"record" yaml:"record"`
}

type NameEntry struct {
	Name  string           `json:"name" yaml:"name"`
	Entry types.NameRecord `json:"record" yaml:"record"`
}

type GenesisState struct {
	Params      types.Params      `json:"params" yaml:"params"`
	Records     []types.RecordObj `json:"records" yaml:"records"`
	Authorities []AuthorityEntry  `json:"authorities" yaml:"authorities"`
	Names       []NameEntry       `json:"names" yaml:"names"`
}

func NewGenesisState(params types.Params, records []types.RecordObj, authorities []AuthorityEntry, names []NameEntry) GenesisState {
	return GenesisState{
		Params:      params,
		Records:     records,
		Authorities: authorities,
		Names:       names,
	}
}

func ValidateGenesis(data GenesisState) error {
	err := data.Params.Validate()
	if err != nil {
		return err
	}

	return nil
}

func DefaultGenesisState() GenesisState {
	return GenesisState{Params: types.DefaultParams()}
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
	keeper.SetParams(ctx, data.Params)

	for _, record := range data.Records {
		obj := record.ToRecord()
		keeper.PutRecord(ctx, obj)

		// Add to record expiry queue if expiry time is in the future.
		if obj.ExpiryTime.After(ctx.BlockTime()) {
			keeper.InsertRecordExpiryQueue(ctx, obj)
		}

		// Note: Bond genesis runs first, so bonds will already be present.
		if record.BondID != "" {
			keeper.AddBondToRecordIndexEntry(ctx, record.BondID, record.ID)
		}
	}

	for _, authority := range data.Authorities {
		// Only import authorities that are marked active.
		if authority.Entry.Status == types.AuthorityActive {
			keeper.SetNameAuthority(ctx, authority.Name, authority.Entry)

			// Add authority name to expiry queue.
			keeper.InsertAuthorityExpiryQueue(ctx, authority.Name, authority.Entry.ExpiryTime)

			// Note: Bond genesis runs first, so bonds will already be present.
			if authority.Entry.BondID != "" {
				keeper.AddBondToAuthorityIndexEntry(ctx, authority.Entry.BondID, authority.Name)
			}
		}
	}

	for _, nameEntry := range data.Names {
		keeper.SetNameRecord(ctx, nameEntry.Name, nameEntry.Entry.ID)
	}

	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)

	records := keeper.ListRecords(ctx)
	recordEntries := []types.RecordObj{}
	for _, record := range records {
		recordEntries = append(recordEntries, record.ToRecordObj())
	}

	authorities := keeper.ListNameAuthorityRecords(ctx)
	authorityEntries := []AuthorityEntry{}
	for name, record := range authorities {
		authorityEntries = append(authorityEntries, AuthorityEntry{
			Name:  name,
			Entry: record,
		})
	}

	names := keeper.ListNameRecords(ctx)
	nameEntries := []NameEntry{}
	for name, record := range names {
		nameEntries = append(nameEntries, NameEntry{
			Name:  name,
			Entry: record,
		})
	}

	return GenesisState{
		Params:      params,
		Records:     recordEntries,
		Authorities: authorityEntries,
		Names:       nameEntries,
	}
}
