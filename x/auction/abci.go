//
// Copyright 2020 Wireline, Inc.
//

package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/auction/internal/keeper"
)

// EndBlocker is called every block, returns updated validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	k.EndBlockerProcessAuctions(ctx)
	return []abci.ValidatorUpdate{}
}
