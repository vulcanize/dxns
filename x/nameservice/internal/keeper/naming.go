//
// Copyright 2020 Wireline, Inc.
//

package keeper

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	wnsUtils "github.com/vulcanize/dxns/utils"
	"github.com/vulcanize/dxns/x/auction"
	"github.com/vulcanize/dxns/x/bond"
	"github.com/vulcanize/dxns/x/nameservice/internal/helpers"
	"github.com/vulcanize/dxns/x/nameservice/internal/types"
)

// NoExpiry => really long duration (used to indicate no-expiry).
const NoExpiry = time.Hour * 24 * 365 * 100

func GetCIDToNamesIndexKey(id types.ID) []byte {
	return append(PrefixCIDToNamesIndex, []byte(id)...)
}

// Generates name -> NameAuthority index key.
func GetNameAuthorityIndexKey(name string) []byte {
	return append(PrefixNameAuthorityRecordIndex, []byte(name)...)
}

func GetAuctionToAuthorityIndexKey(auctionID auction.ID) []byte {
	return append(PrefixAuctionToAuthorityNameIndex, []byte(auctionID)...)
}

// Generates WRN -> NameRecord index key.
func GetNameRecordIndexKey(wrn string) []byte {
	return append(PrefixWRNToNameRecordIndex, []byte(wrn)...)
}

// HasNameAuthority - checks if a name/authority exists.
func (k Keeper) HasNameAuthority(ctx sdk.Context, name string) bool {
	return HasNameAuthority(ctx.KVStore(k.storeKey), name)
}

// HasNameAuthority - checks if a name authority entry exists.
func HasNameAuthority(store sdk.KVStore, name string) bool {
	return store.Has(GetNameAuthorityIndexKey(name))
}

func SetNameAuthority(ctx sdk.Context, store sdk.KVStore, codec *amino.Codec, name string, authority types.NameAuthority) {
	store.Set(GetNameAuthorityIndexKey(name), codec.MustMarshalBinaryBare(authority))
	updateBlockChangesetForNameAuthority(ctx, store, codec, name)
}

// SetNameAuthority creates the NameAutority record.
func (k Keeper) SetNameAuthority(ctx sdk.Context, name string, authority types.NameAuthority) {
	SetNameAuthority(ctx, ctx.KVStore(k.storeKey), k.cdc, name, authority)
}

// GetNameAuthority - gets a name authority from the store.
func GetNameAuthority(store sdk.KVStore, codec *amino.Codec, name string) *types.NameAuthority {
	authorityKey := GetNameAuthorityIndexKey(name)
	if !store.Has(authorityKey) {
		return nil
	}

	bz := store.Get(authorityKey)
	var obj types.NameAuthority
	codec.MustUnmarshalBinaryBare(bz, &obj)

	return &obj
}

// GetNameAuthority - gets a name authority from the store.
func (k Keeper) GetNameAuthority(ctx sdk.Context, name string) *types.NameAuthority {
	return GetNameAuthority(ctx.KVStore(k.storeKey), k.cdc, name)
}

// AddRecordToNameMapping adds a name to the record ID -> []names index.
func AddRecordToNameMapping(store sdk.KVStore, codec *amino.Codec, id types.ID, wrn string) {
	reverseNameIndexKey := GetCIDToNamesIndexKey(id)

	var names []string
	if store.Has(reverseNameIndexKey) {
		codec.MustUnmarshalBinaryBare(store.Get(reverseNameIndexKey), &names)
	}

	nameSet := wnsUtils.SliceToSet(names)
	nameSet.Add(wrn)
	store.Set(reverseNameIndexKey, codec.MustMarshalBinaryBare(wnsUtils.SetToSlice(nameSet)))
}

// RemoveRecordToNameMapping removes a name from the record ID -> []names index.
func RemoveRecordToNameMapping(store sdk.KVStore, codec *amino.Codec, id types.ID, wrn string) {
	reverseNameIndexKey := GetCIDToNamesIndexKey(id)

	var names []string
	codec.MustUnmarshalBinaryBare(store.Get(reverseNameIndexKey), &names)
	nameSet := wnsUtils.SliceToSet(names)
	nameSet.Remove(wrn)

	if nameSet.Cardinality() == 0 {
		// Delete as storing empty slice throws error from baseapp.
		store.Delete(reverseNameIndexKey)
	} else {
		store.Set(reverseNameIndexKey, codec.MustMarshalBinaryBare(wnsUtils.SetToSlice(nameSet)))
	}
}

// SetNameRecord - sets a name record.
func SetNameRecord(store sdk.KVStore, codec *amino.Codec, wrn string, id types.ID, height int64) {
	nameRecordIndexKey := GetNameRecordIndexKey(wrn)

	var nameRecord types.NameRecord
	if store.Has(nameRecordIndexKey) {
		bz := store.Get(nameRecordIndexKey)
		codec.MustUnmarshalBinaryBare(bz, &nameRecord)
		nameRecord.History = append(nameRecord.History, nameRecord.NameRecordEntry)

		// Update old CID -> []Name index.
		if nameRecord.NameRecordEntry.ID != "" {
			RemoveRecordToNameMapping(store, codec, nameRecord.NameRecordEntry.ID, wrn)
		}
	}

	nameRecord.NameRecordEntry = types.NameRecordEntry{
		ID:     id,
		Height: height,
	}

	store.Set(nameRecordIndexKey, codec.MustMarshalBinaryBare(nameRecord))

	// Update new CID -> []Name index.
	if id != "" {
		AddRecordToNameMapping(store, codec, id, wrn)
	}
}

// SetNameRecord - sets a name record.
func (k Keeper) SetNameRecord(ctx sdk.Context, wrn string, id types.ID) {
	SetNameRecord(ctx.KVStore(k.storeKey), k.cdc, wrn, id, ctx.BlockHeight())

	// Update changeset for name.
	k.updateBlockChangesetForName(ctx, wrn)
}

// HasNameRecord - checks if a name record exists.
func (k Keeper) HasNameRecord(ctx sdk.Context, wrn string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(GetNameRecordIndexKey(wrn))
}

// GetNameRecord - gets a name record from the store.
func GetNameRecord(store sdk.KVStore, codec *amino.Codec, wrn string) *types.NameRecord {
	nameRecordKey := GetNameRecordIndexKey(wrn)
	if !store.Has(nameRecordKey) {
		return nil
	}

	bz := store.Get(nameRecordKey)
	var obj types.NameRecord
	codec.MustUnmarshalBinaryBare(bz, &obj)

	return &obj
}

// GetNameRecord - gets a name record from the store.
func (k Keeper) GetNameRecord(ctx sdk.Context, wrn string) *types.NameRecord {
	_, _, authority, err := k.getAuthority(ctx, wrn)
	if err != nil || authority.Status != types.AuthorityActive {
		// If authority is not active (or any other error), lookup fails.
		return nil
	}

	nameRecord := GetNameRecord(ctx.KVStore(k.storeKey), k.cdc, wrn)

	// Name record may not exist.
	if nameRecord == nil {
		return nil
	}

	// Name lookup should fail if the name record is stale.
	// i.e. authority was registered later than the name.
	if authority.Height > nameRecord.Height {
		return nil
	}

	return nameRecord
}

// ListNameAuthorityRecords - get all name authority records.
func (k Keeper) ListNameAuthorityRecords(ctx sdk.Context) map[string]types.NameAuthority {
	nameAuthorityRecords := make(map[string]types.NameAuthority)

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixNameAuthorityRecordIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var record types.NameAuthority
			k.cdc.MustUnmarshalBinaryBare(bz, &record)
			nameAuthorityRecords[string(itr.Key()[len(PrefixNameAuthorityRecordIndex):])] = record
		}
	}

	return nameAuthorityRecords
}

// ListNameRecords - get all name records.
func (k Keeper) ListNameRecords(ctx sdk.Context) map[string]types.NameRecord {
	nameRecords := make(map[string]types.NameRecord)

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixWRNToNameRecordIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		bz := store.Get(itr.Key())
		if bz != nil {
			var record types.NameRecord
			k.cdc.MustUnmarshalBinaryBare(bz, &record)
			nameRecords[string(itr.Key()[len(PrefixWRNToNameRecordIndex):])] = record
		}
	}

	return nameRecords
}

// ResolveWRN resolves a WRN to a record.
func (k Keeper) ResolveWRN(ctx sdk.Context, wrn string) *types.Record {
	_, _, authority, err := k.getAuthority(ctx, wrn)
	if err != nil || authority.Status != types.AuthorityActive {
		// If authority is not active (or any other error), resolution fails.
		return nil
	}

	// Name should not resolve if it's stale.
	// i.e. authority was registered later than the name.
	record, nameRecord := ResolveWRN(ctx.KVStore(k.storeKey), k.cdc, wrn)
	if authority.Height > nameRecord.Height {
		return nil
	}

	return record
}

// ResolveWRN resolves a WRN to a record.
func ResolveWRN(store sdk.KVStore, codec *amino.Codec, wrn string) (*types.Record, *types.NameRecord) {
	nameKey := GetNameRecordIndexKey(wrn)

	if store.Has(nameKey) {
		bz := store.Get(nameKey)
		var obj types.NameRecord
		codec.MustUnmarshalBinaryBare(bz, &obj)

		recordExists := HasRecord(store, obj.ID)
		if !recordExists || obj.ID == "" {
			return nil, &obj
		}

		record := GetRecord(store, codec, obj.ID)
		return &record, &obj
	}

	return nil, nil
}

// UsesAuction returns true if the auction is used for an name authority.
func (k RecordKeeper) UsesAuction(ctx sdk.Context, auctionID auction.ID) bool {
	return k.GetAuctionToAuthorityMapping(ctx, auctionID) != ""
}

func (k RecordKeeper) OnAuction(ctx sdk.Context, auctionID auction.ID) {
	updateBlockChangesetForAuction(ctx, ctx.KVStore(k.storeKey), k.cdc, auctionID)
}

func (k RecordKeeper) OnAuctionBid(ctx sdk.Context, auctionID auction.ID, bidderAddress string) {
	updateBlockChangesetForAuctionBid(ctx, ctx.KVStore(k.storeKey), k.cdc, auctionID, bidderAddress)
}

// OnAuctionWinnerSelected is called when an auction winner is selected.
func (k RecordKeeper) OnAuctionWinnerSelected(ctx sdk.Context, auctionID auction.ID) {
	// Update authority status based on auction status/winner.
	name := k.GetAuctionToAuthorityMapping(ctx, auctionID)
	if name == "" {
		// We don't know about this auction, ignore.
		ctx.Logger().Info(fmt.Sprintf("Ignoring auction notification, name mapping not found: %s", auctionID))
		return
	}

	store := ctx.KVStore(k.storeKey)
	if !HasNameAuthority(store, name) {
		// We don't know about this authority, ignore.
		ctx.Logger().Info(fmt.Sprintf("Ignoring auction notification, authority not found: %s", auctionID))
		return
	}

	authority := GetNameAuthority(store, k.cdc, name)
	auctionObj := k.auctionKeeper.GetAuction(ctx, auctionID)

	if auctionObj.Status == auction.AuctionStatusCompleted {
		store := ctx.KVStore(k.storeKey)

		if auctionObj.WinnerAddress != "" {
			// Mark authority owner and change status to active.
			authority.OwnerAddress = auctionObj.WinnerAddress
			authority.Status = types.AuthorityActive

			// Reset bond ID if required, as owner has changed.
			if authority.BondID != "" {
				RemoveBondToAuthorityIndexEntry(store, authority.BondID, name)
				authority.BondID = ""
			}

			// Update height for updated/changed authority (owner).
			// Can be used to check if names are older than the authority itself (stale names).
			authority.Height = ctx.BlockHeight()

			ctx.Logger().Info(fmt.Sprintf("Winner selected, marking authority as active: %s", name))
		} else {
			// Mark as expired.
			authority.Status = types.AuthorityExpired

			ctx.Logger().Info(fmt.Sprintf("No winner, marking authority as expired: %s", name))
		}

		authority.AuctionID = ""
		SetNameAuthority(ctx, store, k.cdc, name, *authority)

		// Forget about this auction now, we no longer need it.
		removeAuctionToAuthorityMapping(store, auctionID)
	} else {
		ctx.Logger().Info(fmt.Sprintf("Ignoring auction notification, status: %s", auctionObj.Status))
	}
}

// ProcessReserveAuthority reserves a name authority.
func (k Keeper) ProcessReserveAuthority(ctx sdk.Context, msg types.MsgReserveAuthority) (string, error) {
	wrn := fmt.Sprintf("wrn://%s", msg.Name)

	parsedWRN, err := url.Parse(wrn)
	if err != nil {
		return "", sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid name.")
	}

	name := parsedWRN.Host
	if fmt.Sprintf("wrn://%s", name) != wrn {
		return "", sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid name.")
	}

	if strings.Contains(name, ".") {
		return k.ProcessReserveSubAuthority(ctx, name, msg)
	}

	// Reserve name with signer as owner.
	sdkErr := k.createAuthority(ctx, name, msg.Signer, true)
	if sdkErr != nil {
		return "", sdkErr
	}

	return name, nil
}

// ProcessSetAuthorityBond sets a bond on an authority.
func (k Keeper) ProcessSetAuthorityBond(ctx sdk.Context, msg types.MsgSetAuthorityBond) (string, error) {
	name := msg.Name
	signer := msg.Signer

	authority := k.GetNameAuthority(ctx, name)
	if authority == nil {
		return "", sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Name authority not found.")
	}

	if authority.OwnerAddress != signer.String() {
		return "", sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Access denied.")
	}

	if !k.bondKeeper.HasBond(ctx, msg.BondID) {
		return "", sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Bond not found.")
	}

	bond := k.bondKeeper.GetBond(ctx, msg.BondID)
	if bond.Owner != signer.String() {
		return "", sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Bond owner mismatch.")
	}

	// No-op if bond hasn't changed.
	if authority.BondID == msg.BondID {
		return name, nil
	}

	// Remove old bond ID mapping, if any.
	if authority.BondID != "" {
		k.RemoveBondToAuthorityIndexEntry(ctx, authority.BondID, name)
	}

	// Update bond ID for authority.
	authority.BondID = bond.ID
	k.SetNameAuthority(ctx, name, *authority)

	// Add new bond ID mapping.
	k.AddBondToAuthorityIndexEntry(ctx, authority.BondID, name)

	return name, nil
}

func getBondIDToAuthoritiesIndexKey(bondID bond.ID, name string) []byte {
	return append(append(PrefixBondIDToAuthoritiesIndex, []byte(bondID)...), []byte(name)...)
}

// AddBondToAuthorityIndexEntry adds the Bond ID -> [Authority] index entry.
func (k Keeper) AddBondToAuthorityIndexEntry(ctx sdk.Context, bondID bond.ID, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(getBondIDToAuthoritiesIndexKey(bondID, name), []byte{})
}

// RemoveBondToAuthorityIndexEntry removes the Bond ID -> [Authority] index entry.
func (k Keeper) RemoveBondToAuthorityIndexEntry(ctx sdk.Context, bondID bond.ID, name string) {
	RemoveBondToAuthorityIndexEntry(ctx.KVStore(k.storeKey), bondID, name)
}

func RemoveBondToAuthorityIndexEntry(store sdk.KVStore, bondID bond.ID, name string) {
	store.Delete(getBondIDToAuthoritiesIndexKey(bondID, name))
}

func (k Keeper) AddAuctionToAuthorityMapping(ctx sdk.Context, auctionID auction.ID, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GetAuctionToAuthorityIndexKey(auctionID), k.cdc.MustMarshalBinaryBare(name))
}

func removeAuctionToAuthorityMapping(store sdk.KVStore, auctionID auction.ID) {
	store.Delete(GetAuctionToAuthorityIndexKey(auctionID))
}

func (k Keeper) RemoveAuctionToAuthorityMapping(ctx sdk.Context, auctionID auction.ID) {
	removeAuctionToAuthorityMapping(ctx.KVStore(k.storeKey), auctionID)
}

func (k RecordKeeper) GetAuctionToAuthorityMapping(ctx sdk.Context, auctionID auction.ID) string {
	store := ctx.KVStore(k.storeKey)

	auctionToAuthorityIndexKey := GetAuctionToAuthorityIndexKey(auctionID)
	if store.Has(auctionToAuthorityIndexKey) {
		bz := store.Get(auctionToAuthorityIndexKey)
		var name string
		k.cdc.MustUnmarshalBinaryBare(bz, &name)

		return name
	}

	return ""
}

func (k Keeper) createAuthority(ctx sdk.Context, name string, owner sdk.AccAddress, isRoot bool) error {
	moduleParams := k.GetParams(ctx)

	// Authorities can be re-registered if they have expired.
	if k.HasNameAuthority(ctx, name) {
		authority := k.GetNameAuthority(ctx, name)
		if authority.Status != types.AuthorityExpired {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Name already reserved.")
		}
	}

	ownerAccount := k.accountKeeper.GetAccount(ctx, owner)
	if ownerAccount == nil {
		return sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "Account not found.")
	}

	authority := types.NameAuthority{
		Height: ctx.BlockHeight(),

		OwnerAddress: owner.String(),
		BondID:       bond.ID(""),

		// PubKey is only set on first tx from the account, so it might be empty.
		// In that case, it's set later during a "set WRN -> CID" Tx.
		OwnerPublicKey: getAuthorityPubKey(ownerAccount.GetPubKey()),

		Status:    types.AuthorityActive,
		AuctionID: auction.ID(""),

		// Grace period to set bond (assume no auction for now).
		ExpiryTime: ctx.BlockTime().Add(moduleParams.AuthorityGracePeriod),
	}

	// Create auction if root authority and name auctions are enabled.
	if isRoot && moduleParams.AuthorityAuctionEnabled {
		// If auctions are enabled, clear out owner fields. They will be set after a winner is picked.
		authority.OwnerAddress = ""
		authority.OwnerPublicKey = ""

		// Reset bond ID if required.
		if authority.BondID != "" {
			k.RemoveBondToAuthorityIndexEntry(ctx, authority.BondID, name)
			authority.BondID = ""
		}

		commitFee, err := sdk.ParseCoin(moduleParams.CommitFee)
		if err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid name auction commit fee.")
		}

		revealFee, err := sdk.ParseCoin(moduleParams.RevealFee)
		if err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid name auction reveal fee.")
		}

		minimumBid, err := sdk.ParseCoin(moduleParams.MinimumBid)
		if err != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "Invalid name auction minimum bid.")
		}

		params := auction.Params{
			CommitsDuration: moduleParams.CommitsDuration,
			RevealsDuration: moduleParams.RevealsDuration,
			CommitFee:       commitFee,
			RevealFee:       revealFee,
			MinimumBid:      minimumBid,
		}

		// Create an auction.
		msg := auction.NewMsgCreateAuction(params, owner)

		// TODO(ashwin): Perhaps consume extra gas for auction creation.
		auction, sdkErr := k.auctionKeeper.CreateAuction(ctx, msg)
		if sdkErr != nil {
			return sdkErr
		}

		// Create auction ID -> authority name index.
		k.AddAuctionToAuthorityMapping(ctx, auction.ID, name)

		authority.Status = types.AuthorityUnderAuction
		authority.AuctionID = auction.ID
		authority.ExpiryTime = auction.RevealsEndTime.Add(moduleParams.AuthorityGracePeriod)
	}

	k.SetNameAuthority(ctx, name, authority)
	k.InsertAuthorityExpiryQueue(ctx, name, authority.ExpiryTime)

	return nil
}

// ProcessReserveSubAuthority reserves a sub-authority.
func (k Keeper) ProcessReserveSubAuthority(ctx sdk.Context, name string, msg types.MsgReserveAuthority) (string, error) {
	// Get parent authority name.
	names := strings.Split(name, ".")
	parent := strings.Join(names[1:], ".")

	// Check if parent authority exists.
	parentAuthority := k.GetNameAuthority(ctx, parent)
	if parentAuthority == nil {
		return name, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Parent authority not found.")
	}

	// Sub-authority creator needs to be the owner of the parent authority.
	if parentAuthority.OwnerAddress != msg.Signer.String() {
		return name, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Access denied.")
	}

	// Sub-authority owner defaults to parent authority owner.
	subAuthorityOwner := msg.Signer
	if !msg.Owner.Empty() {
		// Override sub-authority owner if provided in message.
		subAuthorityOwner = msg.Owner
	}

	sdkErr := k.createAuthority(ctx, name, subAuthorityOwner, false)
	if sdkErr != nil {
		return "", sdkErr
	}

	return name, nil
}

func getAuthorityPubKey(pubKey crypto.PubKey) string {
	if pubKey != nil {
		return helpers.BytesToBase64(pubKey.Bytes())
	}

	return ""
}

func (k Keeper) getAuthority(ctx sdk.Context, wrn string) (string, *url.URL, *types.NameAuthority, error) {
	parsedWRN, err := url.Parse(wrn)
	if err != nil {
		return "", nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid WRN.")
	}

	name := parsedWRN.Host
	authority := k.GetNameAuthority(ctx, name)
	if authority == nil {
		return name, nil, nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Name authority not found.")
	}

	return name, parsedWRN, authority, nil
}

func (k Keeper) checkWRNAccess(ctx sdk.Context, signer sdk.AccAddress, wrn string) error {
	name, parsedWRN, authority, err := k.getAuthority(ctx, wrn)
	if err != nil {
		return err
	}

	formattedWRN := fmt.Sprintf("wrn://%s%s", name, parsedWRN.RequestURI())
	if formattedWRN != wrn {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Invalid WRN.")
	}

	if authority.OwnerAddress != signer.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Access denied.")
	}

	if authority.Status != types.AuthorityActive {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Authority is not active.")
	}

	if authority.BondID == "" {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "Authority bond not found.")
	}

	if authority.OwnerPublicKey == "" {
		// Try to set owner public key if account has it available now.
		ownerAccount := k.accountKeeper.GetAccount(ctx, signer)
		pubKey := ownerAccount.GetPubKey()
		if pubKey != nil {
			// Update public key in authority record.
			authority.OwnerPublicKey = getAuthorityPubKey(pubKey)
			k.SetNameAuthority(ctx, name, *authority)
		}
	}

	return nil
}

// ProcessSetName creates a WRN -> Record ID mapping.
func (k Keeper) ProcessSetName(ctx sdk.Context, msg types.MsgSetName) error {
	err := k.checkWRNAccess(ctx, msg.Signer, msg.WRN)
	if err != nil {
		return err
	}

	nameRecord := k.GetNameRecord(ctx, msg.WRN)
	if nameRecord != nil && nameRecord.ID == msg.ID {
		// Already pointing to same ID, no-op.
		return nil
	}

	k.SetNameRecord(ctx, msg.WRN, msg.ID)

	return nil
}

// ProcessDeleteName removes a WRN -> Record ID mapping.
func (k Keeper) ProcessDeleteName(ctx sdk.Context, msg types.MsgDeleteName) error {
	err := k.checkWRNAccess(ctx, msg.Signer, msg.WRN)
	if err != nil {
		return err
	}

	if !k.HasNameRecord(ctx, msg.WRN) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Name not found.")
	}

	// Set CID to empty string.
	k.SetNameRecord(ctx, msg.WRN, "")

	return nil
}

func getAuthorityExpiryQueueTimeKey(timestamp time.Time) []byte {
	timeBytes := sdk.FormatTimeBytes(timestamp)
	return append(PrefixExpiryTimeToAuthoritiesIndex, timeBytes...)
}

func (k Keeper) InsertAuthorityExpiryQueue(ctx sdk.Context, name string, expiryTime time.Time) {
	timeSlice := k.GetAuthorityExpiryQueueTimeSlice(ctx, expiryTime)
	timeSlice = append(timeSlice, name)
	k.SetAuthorityExpiryQueueTimeSlice(ctx, expiryTime, timeSlice)
}

func (k Keeper) GetAuthorityExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (names []string) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(getAuthorityExpiryQueueTimeKey(timestamp))
	if bz == nil {
		return []string{}
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &names)
	return names
}

func (k Keeper) SetAuthorityExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time, names []string) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(names)
	store.Set(getAuthorityExpiryQueueTimeKey(timestamp), bz)
}

// ProcessAuthorityExpiryQueue tries to renew expiring authorities (by collecting rent) else marks them as expired.
func (k Keeper) ProcessAuthorityExpiryQueue(ctx sdk.Context) {
	names := k.GetAllExpiredAuthorities(ctx, ctx.BlockHeader().Time)
	for _, name := range names {
		authority := k.GetNameAuthority(ctx, name)

		// If authority doesn't have an associated bond or if bond no longer exists, mark it expired.
		if authority.BondID == "" || !k.bondKeeper.HasBond(ctx, authority.BondID) {
			authority.Status = types.AuthorityExpired
			k.SetNameAuthority(ctx, name, *authority)
			k.DeleteAuthorityExpiryQueue(ctx, name, *authority)

			ctx.Logger().Info(fmt.Sprintf("Marking authority expired as no bond present: %s", name))

			return
		}

		// Try to renew the authority by taking rent.
		k.TryTakeAuthorityRent(ctx, name, *authority)
	}
}

// DeleteAuthorityExpiryQueueTimeSlice deletes a specific authority expiry queue timeslice.
func (k Keeper) DeleteAuthorityExpiryQueueTimeSlice(ctx sdk.Context, timestamp time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getAuthorityExpiryQueueTimeKey(timestamp))
}

// DeleteAuthorityExpiryQueue deletes an authority name from the authority expiry queue.
func (k Keeper) DeleteAuthorityExpiryQueue(ctx sdk.Context, name string, authority types.NameAuthority) {
	timeSlice := k.GetAuthorityExpiryQueueTimeSlice(ctx, authority.ExpiryTime)
	newTimeSlice := []string{}

	for _, existingName := range timeSlice {
		if !bytes.Equal([]byte(existingName), []byte(name)) {
			newTimeSlice = append(newTimeSlice, existingName)
		}
	}

	if len(newTimeSlice) == 0 {
		k.DeleteAuthorityExpiryQueueTimeSlice(ctx, authority.ExpiryTime)
	} else {
		k.SetAuthorityExpiryQueueTimeSlice(ctx, authority.ExpiryTime, newTimeSlice)
	}
}

// GetAllExpiredAuthorities returns a concatenated list of all the timeslices before currTime.
func (k Keeper) GetAllExpiredAuthorities(ctx sdk.Context, currTime time.Time) (expiredAuthorityNames []string) {
	// Gets an iterator for all timeslices from time 0 until the current block header time.
	itr := k.AuthorityExpiryQueueIterator(ctx, ctx.BlockHeader().Time)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		timeslice := []string{}
		k.cdc.MustUnmarshalBinaryLengthPrefixed(itr.Value(), &timeslice)
		expiredAuthorityNames = append(expiredAuthorityNames, timeslice...)
	}

	return expiredAuthorityNames
}

// AuthorityExpiryQueueIterator returns all the authority expiry queue timeslices from time 0 until endTime.
func (k Keeper) AuthorityExpiryQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	rangeEndBytes := sdk.InclusiveEndBytes(getAuthorityExpiryQueueTimeKey(endTime))
	return store.Iterator(PrefixExpiryTimeToAuthoritiesIndex, rangeEndBytes)
}

func (k Keeper) GetAuthorityExpiryQueue(ctx sdk.Context) (expired map[string][]string) {
	records := make(map[string][]string)

	store := ctx.KVStore(k.storeKey)
	itr := sdk.KVStorePrefixIterator(store, PrefixExpiryTimeToAuthoritiesIndex)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		var record []string
		k.cdc.MustUnmarshalBinaryLengthPrefixed(itr.Value(), &record)
		records[string(itr.Key()[len(PrefixExpiryTimeToAuthoritiesIndex):])] = record
	}

	return records
}

// TryTakeAuthorityRent tries to take rent from the authority bond.
func (k Keeper) TryTakeAuthorityRent(ctx sdk.Context, name string, authority types.NameAuthority) {
	ctx.Logger().Info(fmt.Sprintf("Trying to take rent for authority: %s", name))

	params := k.GetParams(ctx)

	rent, err := sdk.ParseCoins(params.AuthorityRent)
	if err != nil {
		panic("Invalid authority rent.")
	}

	sdkErr := k.bondKeeper.TransferCoinsToModuleAccount(ctx, authority.BondID, types.AuthorityRentModuleAccountName, rent)
	if sdkErr != nil {
		// Insufficient funds, mark authority as expired.
		authority.Status = types.AuthorityExpired
		k.SetNameAuthority(ctx, name, authority)
		k.DeleteAuthorityExpiryQueue(ctx, name, authority)

		ctx.Logger().Info(fmt.Sprintf("Insufficient funds in owner account to pay authority rent, marking as expired: %s", name))

		return
	}

	// Delete old expiry queue entry, create new one.
	k.DeleteAuthorityExpiryQueue(ctx, name, authority)
	authority.ExpiryTime = ctx.BlockTime().Add(params.AuthorityRentDuration)
	k.InsertAuthorityExpiryQueue(ctx, name, authority.ExpiryTime)

	// Save authority.
	authority.Status = types.AuthorityActive
	k.SetNameAuthority(ctx, name, authority)
	k.AddBondToAuthorityIndexEntry(ctx, authority.BondID, name)

	ctx.Logger().Info(fmt.Sprintf("Authority rent paid successfully: %s", name))
}
