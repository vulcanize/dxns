//
// Copyright 2020 Wireline, Inc.
//

package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	sync "github.com/vulcanize/dxns/cmd/dxnsd-lite/sync"
)

func main() {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "dxnsd-lite",
		Short: "DXNS Lite",
	}

	rootCmd.PersistentFlags().String("chain-id", "wireline-1", "Chain identifier")
	rootCmd.PersistentFlags().String("log-level", "debug", "Log level")
	rootCmd.PersistentFlags().StringP("node", "n", "tcp://localhost:26657", "Upstream WNS node RPC address")
	rootCmd.PersistentFlags().String("log-file", "", "File to tail for GQL 'getLogs' API")

	rootCmd.AddCommand(versionCmd, initCmd, startCmd)

	executor := cli.PrepareBaseCmd(rootCmd, "DXNSL", os.ExpandEnv(sync.DefaultLightNodeHome))
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
