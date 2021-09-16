//
// Copyright 2020 Wireline, Inc.
//

package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/vulcanize/dxns/x/auction/internal/types"
)

// NewHandler returns a handler for "auction" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgCreateAuction:
			return handleMsgCreateAuction(ctx, keeper, msg)
		case types.MsgCommitBid:
			return handleMsgCommitBid(ctx, keeper, msg)
		case types.MsgRevealBid:
			return handleMsgRevealBid(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

// Handle MsgCreateAuction.
func handleMsgCreateAuction(ctx sdk.Context, keeper Keeper, msg types.MsgCreateAuction) (*sdk.Result, error) {
	auction, err := keeper.CreateAuction(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Data:   []byte(auction.ID),
		Events: ctx.EventManager().Events(),
	}, nil
}

// Handle MsgCommitBid.
func handleMsgCommitBid(ctx sdk.Context, keeper Keeper, msg types.MsgCommitBid) (*sdk.Result, error) {
	auction, err := keeper.CommitBid(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Data:   []byte(auction.ID),
		Events: ctx.EventManager().Events(),
	}, nil
}

// Handle MsgRevealBid.
func handleMsgRevealBid(ctx sdk.Context, keeper Keeper, msg types.MsgRevealBid) (*sdk.Result, error) {
	auction, err := keeper.RevealBid(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &sdk.Result{
		Data:   []byte(auction.ID),
		Events: ctx.EventManager().Events(),
	}, nil
}
