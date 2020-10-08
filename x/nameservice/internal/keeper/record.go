//
// Copyright 2020 Wireline, Inc.
//

package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethCrypto "github.com/cosmos/ethermint/crypto"
	"github.com/wirelineio/dxns/x/bond"
	"github.com/wirelineio/dxns/x/nameservice/internal/helpers"
	"github.com/wirelineio/dxns/x/nameservice/internal/types"
)

// ProcessSetRecord creates a record.
func (k Keeper) ProcessSetRecord(ctx sdk.Context, msg types.MsgSetRecord) (*types.Record, error) {
	payload := msg.Payload.ToPayload()
	record := types.Record{Attributes: payload.Record, BondID: msg.BondID}

	// Check signatures.
	resourceSignBytes, _ := record.GetSignBytes()
	cid, err := record.GetCID()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid record JSON")
	}

	record.ID = cid

	if exists := k.HasRecord(ctx, record.ID); exists {
		// Immutable record already exists. No-op.
		return &record, nil
	}

	record.Owners = []string{}
	for _, sig := range payload.Signatures {
		pubKey := ethCrypto.PubKeySecp256k1(helpers.BytesFromBase64(sig.PubKey))
		sigOK := pubKey.VerifyBytes(resourceSignBytes, helpers.BytesFromBase64(sig.Signature))
		if !sigOK {
			fmt.Println("Signature mismatch: ", sig.PubKey)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Invalid signature.")
		}

		record.Owners = append(record.Owners, helpers.GetAddressFromPubKey(pubKey))
	}

	// Sort owners list.
	sort.Strings(record.Owners)

	sdkErr := k.processRecord(ctx, &record, false)
	if sdkErr != nil {
		return nil, sdkErr
	}

	return &record, nil
}

// ProcessRenewRecord renews a record.
func (k Keeper) ProcessRenewRecord(ctx sdk.Context, msg types.MsgRenewRecord) (*types.Record, error) {
	if !k.HasRecord(ctx, msg.ID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Record not found.")
	}

	// Check if renewal is required (i.e. expired record marked as deleted).
	record := k.GetRecord(ctx, msg.ID)
	if !record.Deleted || record.ExpiryTime.After(ctx.BlockTime()) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Renewal not required.")
	}

	err := k.processRecord(ctx, &record, true)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (k Keeper) processRecord(ctx sdk.Context, record *types.Record, isRenewal bool) error {
	params := k.GetParams(ctx)

	rent, err := sdk.ParseCoins(params.RecordRent)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid record rent.")
	}

	sdkErr := k.bondKeeper.TransferCoinsToModuleAccount(ctx, record.BondID, types.RecordRentModuleAccountName, rent)
	if sdkErr != nil {
		return sdkErr
	}

	record.CreateTime = ctx.BlockHeader().Time
	record.ExpiryTime = ctx.BlockHeader().Time.Add(params.RecordRentDuration)
	record.Deleted = false

	k.PutRecord(ctx, *record)
	k.InsertRecordExpiryQueue(ctx, *record)

	// Renewal doesn't change the name and bond indexes.
	if !isRenewal {
		k.AddBondToRecordIndexEntry(ctx, record.BondID, record.ID)
	}

	return nil
}

// ProcessAssociateBond associates a record with a bond.
func (k Keeper) ProcessAssociateBond(ctx sdk.Context, msg types.MsgAssociateBond) (*types.Record, error) {

	if !k.HasRecord(ctx, msg.ID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Record not found.")
	}

	if !k.bondKeeper.HasBond(ctx, msg.BondID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	// Check if already associated with a bond.
	record := k.GetRecord(ctx, msg.ID)
	if record.BondID != "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond already exists.")
	}

	// Only the bond owner can associate a record with the bond.
	bond := k.bondKeeper.GetBond(ctx, msg.BondID)
	if msg.Signer.String() != bond.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	record.BondID = msg.BondID
	k.PutRecord(ctx, record)
	k.AddBondToRecordIndexEntry(ctx, msg.BondID, msg.ID)

	// Required so that renewal is triggered (with new bond ID) for expired records.
	if record.Deleted {
		k.InsertRecordExpiryQueue(ctx, record)
	}

	return &record, nil
}

// ProcessDissociateBond dissociates a record from its bond.
func (k Keeper) ProcessDissociateBond(ctx sdk.Context, msg types.MsgDissociateBond) (*types.Record, error) {

	if !k.HasRecord(ctx, msg.ID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Record not found.")
	}

	// Check if associated with a bond.
	record := k.GetRecord(ctx, msg.ID)
	bondID := record.BondID
	if bondID == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond not found.")
	}

	// Only the bond owner can dissociate a record from the bond.
	bond := k.bondKeeper.GetBond(ctx, bondID)
	if msg.Signer.String() != bond.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	// Clear bond ID.
	record.BondID = ""
	k.PutRecord(ctx, record)
	k.RemoveBondToRecordIndexEntry(ctx, bondID, record.ID)

	return &record, nil
}

// ProcessDissociateRecords dissociates all records associated with a given bond.
func (k Keeper) ProcessDissociateRecords(ctx sdk.Context, msg types.MsgDissociateRecords) (*bond.Bond, error) {

	if !k.bondKeeper.HasBond(ctx, msg.BondID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	// Only the bond owner can dissociate all records from the bond.
	bond := k.bondKeeper.GetBond(ctx, msg.BondID)
	if msg.Signer.String() != bond.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	// Dissociate all records from the bond.
	records := k.recordKeeper.QueryRecordsByBond(ctx, msg.BondID)
	for _, record := range records {
		// Clear bond ID.
		record.BondID = ""
		k.PutRecord(ctx, record)
		k.RemoveBondToRecordIndexEntry(ctx, msg.BondID, record.ID)
	}

	return &bond, nil
}

// ProcessReassociateRecords switches records from and old to new bond.
func (k Keeper) ProcessReassociateRecords(ctx sdk.Context, msg types.MsgReassociateRecords) (*bond.Bond, error) {

	if !k.bondKeeper.HasBond(ctx, msg.OldBondID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Old bond not found.")
	}

	if !k.bondKeeper.HasBond(ctx, msg.NewBondID) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "New bond not found.")
	}

	// Only the bond owner can reassociate all records.
	oldBond := k.bondKeeper.GetBond(ctx, msg.OldBondID)
	if msg.Signer.String() != oldBond.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Old bond owner mismatch.")
	}

	newBond := k.bondKeeper.GetBond(ctx, msg.NewBondID)
	if msg.Signer.String() != newBond.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "New bond owner mismatch.")
	}

	// Reassociate all records.
	records := k.recordKeeper.QueryRecordsByBond(ctx, msg.OldBondID)
	for _, record := range records {
		// Switch bond ID.
		record.BondID = msg.NewBondID
		k.PutRecord(ctx, record)

		k.RemoveBondToRecordIndexEntry(ctx, msg.OldBondID, record.ID)
		k.AddBondToRecordIndexEntry(ctx, msg.NewBondID, record.ID)

		// Required so that renewal is triggered (with new bond ID) for expired records.
		if record.Deleted {
			k.InsertRecordExpiryQueue(ctx, record)
		}
	}

	return &newBond, nil
}
