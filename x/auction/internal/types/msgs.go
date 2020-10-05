//
// Copyright 2020 Wireline, Inc.
//

package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is the module name router key
const RouterKey = ModuleName // this was defined in your key.go file

// MsgCreateAuction defines a create auction message.
type MsgCreateAuction struct {
	CommitsDuration time.Duration  `json:"commitsDuration,omitempty"`
	RevealsDuration time.Duration  `json:"revealsDuration,omitempty"`
	CommitFee       sdk.Coin       `json:"commitFee,omitempty"`
	RevealFee       sdk.Coin       `json:"revealFee,omitempty"`
	MinimumBid      sdk.Coin       `json:"minimumBid,omitempty"`
	Signer          sdk.AccAddress `json:"signer"`
}

// NewMsgCreateAuction is the constructor function for MsgCreateAuction.
func NewMsgCreateAuction(params Params, signer sdk.AccAddress) MsgCreateAuction {
	return MsgCreateAuction{
		CommitsDuration: params.CommitsDuration,
		RevealsDuration: params.RevealsDuration,
		CommitFee:       params.CommitFee,
		RevealFee:       params.RevealFee,
		MinimumBid:      params.MinimumBid,
		Signer:          signer,
	}
}

// Route Implements Msg.
func (msg MsgCreateAuction) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgCreateAuction) Type() string { return "create" }

// ValidateBasic Implements Msg.
func (msg MsgCreateAuction) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if msg.CommitsDuration <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "commit phase duration invalid.")
	}

	if msg.RevealsDuration <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "reveal phase duration invalid.")
	}

	if !msg.MinimumBid.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "minimum bid should be greater than zero.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgCreateAuction) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgCreateAuction) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgCommitBid defines a commit bid message.
type MsgCommitBid struct {
	AuctionID  ID             `json:"auctionId,omitempty"`
	CommitHash string         `json:"commit,omitempty"`
	Signer     sdk.AccAddress `json:"signer"`
}

// NewMsgCommitBid is the constructor function for MsgCommitBid.
func NewMsgCommitBid(auctionID string, commitHash string, signer sdk.AccAddress) MsgCommitBid {

	return MsgCommitBid{
		AuctionID:  ID(auctionID),
		CommitHash: commitHash,
		Signer:     signer,
	}
}

// Route Implements Msg.
func (msg MsgCommitBid) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgCommitBid) Type() string { return "commit" }

// ValidateBasic Implements Msg.
func (msg MsgCommitBid) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if msg.AuctionID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction ID.")
	}

	if msg.CommitHash == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid commit hash.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgCommitBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgCommitBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// MsgRevealBid defines a reveal bid message.
type MsgRevealBid struct {
	AuctionID ID             `json:"auctionId,omitempty"`
	Reveal    string         `json:"reveal,omitempty"`
	Signer    sdk.AccAddress `json:"signer"`
}

// NewMsgRevealBid is the constructor function for MsgRevealBid.
func NewMsgRevealBid(auctionID string, reveal string, signer sdk.AccAddress) MsgRevealBid {

	return MsgRevealBid{
		AuctionID: ID(auctionID),
		Reveal:    reveal,
		Signer:    signer,
	}
}

// Route Implements Msg.
func (msg MsgRevealBid) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgRevealBid) Type() string { return "reveal" }

// ValidateBasic Implements Msg.
func (msg MsgRevealBid) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if msg.AuctionID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid auction ID.")
	}

	if msg.Reveal == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid reveal data.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgRevealBid) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgRevealBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
