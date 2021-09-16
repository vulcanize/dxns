//
// Copyright 2020 Wireline, Inc.
//

package sync

import (
	"errors"
	"fmt"
	"time"

	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"github.com/vulcanize/dxns/x/nameservice"
)

// DefaultLightNodeHome is the root directory for the dxnsd-lite node.
const DefaultLightNodeHome = "$HOME/.wire/dxnsd-lite"

const (
	NameStorePath    = "/store/nameservice/key"
	AuctionStorePath = "/store/auction/key"
)

// getCurrentHeight gets the current WNS block height.
func (rpcNodeHandler *RPCNodeHandler) getCurrentHeight() (int64, error) {
	rpcNodeHandler.Calls++
	rpcNodeHandler.LastCalledAt = time.Now().UTC()

	// Note: Always get from primary node.
	status, err := rpcNodeHandler.Client.Status()
	if err != nil {
		rpcNodeHandler.Errors++
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

func (rpcNodeHandler *RPCNodeHandler) getBlockChangeset(ctx *Context, height int64) (*nameservice.BlockChangeset, error) {
	value, err := rpcNodeHandler.getStoreValue(ctx, NameStorePath, nameservice.GetBlockChangesetIndexKey(height), height)
	if err != nil {
		return nil, err
	}

	var changeset nameservice.BlockChangeset
	ctx.codec.MustUnmarshalBinaryBare(value, &changeset)

	return &changeset, nil
}

func (rpcNodeHandler *RPCNodeHandler) getStoreValue(ctx *Context, path string, key []byte, height int64) ([]byte, error) {
	opts := rpcclient.ABCIQueryOptions{
		Height: height,
		Prove:  true,
	}

	rpcNodeHandler.Calls++
	rpcNodeHandler.LastCalledAt = time.Now().UTC()

	res, err := rpcNodeHandler.Client.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		rpcNodeHandler.Errors++
		return nil, err
	}

	if res.Response.IsErr() {
		rpcNodeHandler.Errors++
		return nil, fmt.Errorf("error fetching state: %s", res.Response.GetLog())
	}

	if res.Response.Height == 0 && res.Response.Value != nil {
		rpcNodeHandler.Errors++
		return nil, errors.New("invalid response height/value")
	}

	if res.Response.Height > 0 && res.Response.Height != height {
		rpcNodeHandler.Errors++
		return nil, fmt.Errorf("invalid response height: %d", res.Response.Height)
	}

	if res.Response.Height > 0 {
		// Note: Fails with `panic: runtime error: invalid memory address or nil pointer dereference` if called with empty response.
		err = VerifyProof(ctx, path, res.Response)
		if err != nil {
			return nil, err
		}
	}

	return res.Response.Value, nil
}

func (ctx *Context) getStoreSubspace(subspace string, key []byte, height int64) ([]storeTypes.KVPair, error) {
	opts := rpcclient.ABCIQueryOptions{Height: height}
	path := fmt.Sprintf("/store/%s/subspace", subspace)

	ctx.PrimaryNode.Calls++
	ctx.PrimaryNode.LastCalledAt = time.Now().UTC()

	res, err := ctx.PrimaryNode.Client.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		ctx.PrimaryNode.Errors++
		return nil, err
	}

	var KVs []storeTypes.KVPair
	ctx.codec.MustUnmarshalBinaryLengthPrefixed(res.Response.Value, &KVs)

	return KVs, nil
}
