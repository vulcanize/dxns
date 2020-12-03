//
// Copyright 2019 Wireline, Inc.
//

package gql

import (
	"encoding/base64"
	"errors"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/core"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

func decodeStdTx(codec *amino.Codec, tx string) (*auth.StdTx, error) {
	bytes, err := base64.StdEncoding.DecodeString(tx)
	if err != nil {
		return nil, errors.New("{ \"log\": \"Tx bytes not base64 encoded.\" }")
	}

	var stdTx auth.StdTx
	err = codec.UnmarshalJSON(bytes, &stdTx)
	if err != nil {
		return nil, errors.New("{ \"log\": \"Invalid Tx bytes, check request JSON.\" }")
	}

	return &stdTx, nil
}

func broadcastTx(r *mutationResolver, stdTx *auth.StdTx) (*ctypes.ResultBroadcastTxCommit, error) {
	txBytes, err := r.Resolver.codec.MarshalBinaryLengthPrefixed(stdTx)
	if err != nil {
		return nil, err
	}

	ctx := &rpctypes.Context{}
	res, err := core.BroadcastTxCommit(ctx, txBytes)
	if err != nil {
		return nil, err
	}

	if res.CheckTx.IsErr() {
		errBytes, _ := res.CheckTx.MarshalJSON()
		return nil, errors.New(string(errBytes))
	}

	if res.DeliverTx.IsErr() {
		errBytes, _ := res.DeliverTx.MarshalJSON()
		return nil, errors.New(string(errBytes))
	}

	return res, nil
}
