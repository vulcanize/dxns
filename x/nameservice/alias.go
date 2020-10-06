//
// Copyright 2019 Wireline, Inc.
//

package nameservice

import (
	"github.com/wirelineio/dxns/x/nameservice/internal/keeper"
	"github.com/wirelineio/dxns/x/nameservice/internal/types"
)

const (
	ModuleName                     = types.ModuleName
	RecordRentModuleAccountName    = types.RecordRentModuleAccountName
	AuthorityRentModuleAccountName = types.AuthorityRentModuleAccountName
	RouterKey                      = types.RouterKey
	StoreKey                       = types.StoreKey
)

var (
	DefaultParamspace = types.DefaultParamspace
	NewKeeper         = keeper.NewKeeper
	NewRecordKeeper   = keeper.NewRecordKeeper
	NewQuerier        = keeper.NewQuerier
	ModuleCdc         = types.ModuleCdc
	RegisterCodec     = types.RegisterCodec

	RegisterInvariants = keeper.RegisterInvariants

	PrefixCIDToRecordIndex         = keeper.PrefixCIDToRecordIndex
	PrefixNameAuthorityRecordIndex = keeper.PrefixNameAuthorityRecordIndex
	PrefixWRNToNameRecordIndex     = keeper.PrefixWRNToNameRecordIndex

	GetBlockChangesetIndexKey = keeper.GetBlockChangesetIndexKey
	GetRecordIndexKey         = keeper.GetRecordIndexKey
	GetNameAuthorityIndexKey  = keeper.GetNameAuthorityIndexKey
	GetNameRecordIndexKey     = keeper.GetNameRecordIndexKey

	HasRecord        = keeper.HasRecord
	GetRecord        = keeper.GetRecord
	ResolveWRN       = keeper.ResolveWRN
	GetNameAuthority = keeper.GetNameAuthority
	GetNameRecord    = keeper.GetNameRecord
	MatchRecords     = keeper.MatchRecords
	KeySyncStatus    = keeper.KeySyncStatus

	SetNameRecord             = keeper.SetNameRecord
	AddRecordToNameMapping    = keeper.AddRecordToNameMapping
	RemoveRecordToNameMapping = keeper.RemoveRecordToNameMapping
)

type (
	Keeper       = keeper.Keeper
	RecordKeeper = keeper.RecordKeeper

	MsgSetRecord = types.MsgSetRecord

	ID        = types.ID
	Record    = types.Record
	RecordObj = types.RecordObj

	NameAuthority   = types.NameAuthority
	NameRecord      = types.NameRecord
	NameRecordEntry = types.NameRecordEntry

	BlockChangeset = types.BlockChangeset
)
