//
// Copyright 2020 Wireline, Inc.
//

package gql

import (
	"context"
	"errors"

	baseGql "github.com/vulcanize/dxns/gql"
)

type mutationResolver struct{ *Resolver }

// Mutation is the entry point to tx execution.
func (r *Resolver) Mutation() baseGql.MutationResolver {
	return &mutationResolver{r}
}

func (r *mutationResolver) InsertRecord(ctx context.Context, attributes []*baseGql.KeyValueInput) (*baseGql.Record, error) {
	// Only supported by mock server.
	return nil, errors.New("Not supported")
}

func (r *mutationResolver) Submit(ctx context.Context, tx string) (*string, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}

func (r *queryResolver) GetAccounts(ctx context.Context, addresses []string) ([]*baseGql.Account, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}

func (r *queryResolver) GetAccount(ctx context.Context, address string) (*baseGql.Account, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}

func (r *queryResolver) GetBondsByIds(ctx context.Context, ids []string) ([]*baseGql.Bond, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}

func (r *queryResolver) GetBond(ctx context.Context, id string) (*baseGql.Bond, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}

func (r *queryResolver) QueryBonds(ctx context.Context, attributes []*baseGql.KeyValueInput) ([]*baseGql.Bond, error) {
	// Only supported by a full-node.
	return nil, errors.New("Not supported")
}
