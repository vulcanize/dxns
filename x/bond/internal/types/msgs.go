//
// Copyright 2019 Wireline, Inc.
//

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RouterKey is the module name router key
const RouterKey = ModuleName // this was defined in your key.go file

var _ sdk.Msg = &MsgCreateBond{}

// MsgCreateBond defines a create bond message.
type MsgCreateBond struct {
	Coins  sdk.Coins      `json:"coins"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgCreateBond is the constructor function for MsgCreateBond.
func NewMsgCreateBond(coins sdk.Coins, signer sdk.AccAddress) MsgCreateBond {
	return MsgCreateBond{
		Coins:  coins,
		Signer: signer,
	}
}

// Route Implements Msg.
func (msg MsgCreateBond) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgCreateBond) Type() string { return "create" }

// ValidateBasic Implements Msg.
func (msg MsgCreateBond) ValidateBasic() error {

	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid amount.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgCreateBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgCreateBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgRefillBond{}

// MsgRefillBond defines a refill bond message.
type MsgRefillBond struct {
	ID     ID             `json:"id"`
	Coins  sdk.Coins      `json:"coins"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgRefillBond is the constructor function for MsgRefillBond.
func NewMsgRefillBond(id string, denom string, amount int64, signer sdk.AccAddress) MsgRefillBond {
	return MsgRefillBond{
		ID:     ID(id),
		Coins:  sdk.NewCoins(sdk.NewInt64Coin(denom, amount)),
		Signer: signer,
	}
}

// Route Implements Msg.
func (msg MsgRefillBond) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgRefillBond) Type() string { return "refill" }

// ValidateBasic Implements Msg.
func (msg MsgRefillBond) ValidateBasic() error {

	if string(msg.ID) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid bond ID.")
	}

	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid amount.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgRefillBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgRefillBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgWithdrawBond{}

// MsgWithdrawBond defines a withdraw (funds from) bond message.
type MsgWithdrawBond struct {
	ID     ID             `json:"id"`
	Coins  sdk.Coins      `json:"coins"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgWithdrawBond is the constructor function for MsgWithdrawBond.
func NewMsgWithdrawBond(id string, denom string, amount int64, signer sdk.AccAddress) MsgWithdrawBond {
	return MsgWithdrawBond{
		ID:     ID(id),
		Coins:  sdk.NewCoins(sdk.NewInt64Coin(denom, amount)),
		Signer: signer,
	}
}

// Route Implements Msg.
func (msg MsgWithdrawBond) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgWithdrawBond) Type() string { return "withdraw" }

// ValidateBasic Implements Msg.
func (msg MsgWithdrawBond) ValidateBasic() error {

	if string(msg.ID) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid bond ID.")
	}

	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	if len(msg.Coins) == 0 || !msg.Coins.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid amount.")
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgWithdrawBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgWithdrawBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

var _ sdk.Msg = &MsgCancelBond{}

// MsgCancelBond defines a cancel bond message.
type MsgCancelBond struct {
	ID     ID             `json:"id"`
	Signer sdk.AccAddress `json:"signer"`
}

// NewMsgCancelBond is the constructor function for MsgCancelBond.
func NewMsgCancelBond(id string, signer sdk.AccAddress) MsgCancelBond {
	return MsgCancelBond{
		ID:     ID(id),
		Signer: signer,
	}
}

// Route Implements Msg.
func (msg MsgCancelBond) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgCancelBond) Type() string { return "cancel" }

// ValidateBasic Implements Msg.
func (msg MsgCancelBond) ValidateBasic() error {

	if string(msg.ID) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid bond ID.")
	}

	if msg.Signer.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Signer.String())
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgCancelBond) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgCancelBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
