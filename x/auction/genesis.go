//
// Copyright 2020 Wireline, Inc.
//

package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/auction/internal/types"
)

type GenesisState struct {
	Params   types.Params    `json:"params" yaml:"params"`
	Auctions []types.Auction `json:"auctions" yaml:"auctions"`
}

func NewGenesisState(params types.Params, auctions []types.Auction) GenesisState {
	return GenesisState{Params: params, Auctions: auctions}
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

	for _, auction := range data.Auctions {
		keeper.SaveAuction(ctx, auction)
	}

	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)
	auctions := keeper.ListAuctions(ctx)

	return GenesisState{Params: params, Auctions: auctions}
}
