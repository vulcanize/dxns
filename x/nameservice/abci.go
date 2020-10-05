//
// Copyright 2019 Wireline, Inc.
//

package nameservice

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/wirelineio/dxns/x/nameservice/internal/keeper"
)

// EndBlocker is called every block, returns updated validator set.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	k.ProcessRecordExpiryQueue(ctx)
	k.ProcessAuthorityExpiryQueue(ctx)

	return []abci.ValidatorUpdate{}
}
