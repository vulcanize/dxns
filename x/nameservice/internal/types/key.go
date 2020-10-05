//
// Copyright 2019 Wireline, Inc.
//

package types

const (
	// ModuleName is the name of the module
	ModuleName = "nameservice"

	// RecordRentModuleAccountName is the name of the module account that keeps track of record rents paid.
	RecordRentModuleAccountName = "record_rent"

	// AuthorityRentModuleAccountName is the name of the module account that keeps track of authority rents paid.
	AuthorityRentModuleAccountName = "authority_rent"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName
)
