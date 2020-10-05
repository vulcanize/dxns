//
// Copyright 2019 Wireline, Inc.
//

package bond

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/bond/internal/types"
)

type GenesisState struct {
	Params types.Params `json:"params" yaml:"params"`
	Bonds  []types.Bond `json:"bonds" yaml:"bonds"`
}

func NewGenesisState(params types.Params, bonds []types.Bond) GenesisState {
	return GenesisState{Params: params, Bonds: bonds}
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

	for _, bond := range data.Bonds {
		keeper.SaveBond(ctx, bond)
	}

	return []abci.ValidatorUpdate{}
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)
	bonds := keeper.ListBonds(ctx)

	return GenesisState{Params: params, Bonds: bonds}
}
