//
// Copyright 2019 Wireline, Inc.
//

package gql

import (
	"net/http"

	"github.com/spf13/viper"

	"github.com/99designs/gqlgen/handler"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/wirelineio/dxns/x/auction"
	"github.com/wirelineio/dxns/x/bond"
	"github.com/wirelineio/dxns/x/nameservice"

	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

// Server configures and starts the GQL server.
func Server(baseApp *bam.BaseApp,
	cdc *codec.Codec,
	keeper nameservice.Keeper,
	bondKeeper bond.Keeper,
	accountKeeper auth.AccountKeeper,
	auctionKeeper auction.Keeper) {

	if !viper.GetBool("gql-server") {
		return
	}

	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		Debug:          false,
	}).Handler)

	logFile := viper.GetString("log-file")

	if viper.GetBool("gql-playground") {
		apiBase := viper.GetString("gql-playground-api-base")

		router.Handle("/webui", handler.Playground("Wireline Naming Service", apiBase+"/api"))

		// TODO(ashwin): Kept for backward compat.
		router.Handle("/console", handler.Playground("Wireline Naming Service", apiBase+"/graphql"))
	}

	router.Handle("/api", handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		baseApp:       baseApp,
		codec:         cdc,
		keeper:        keeper,
		bondKeeper:    bondKeeper,
		accountKeeper: accountKeeper,
		auctionKeeper: auctionKeeper,
		logFile:       logFile,
	}})))

	// TODO(ashwin): Kept for backward compat.
	router.Handle("/graphql", handler.GraphQL(NewExecutableSchema(Config{Resolvers: &Resolver{
		baseApp:       baseApp,
		codec:         cdc,
		keeper:        keeper,
		bondKeeper:    bondKeeper,
		accountKeeper: accountKeeper,
		auctionKeeper: auctionKeeper,
		logFile:       logFile,
	}})))

	err := http.ListenAndServe(":"+viper.GetString("gql-port"), router)
	if err != nil {
		panic(err)
	}
}
