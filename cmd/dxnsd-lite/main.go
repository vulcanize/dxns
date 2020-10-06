//
// Copyright 2020 Wireline, Inc.
//

package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

// DefaultLightNodeHome is the root directory for the wnsd-lite node.
var DefaultLightNodeHome = os.ExpandEnv("$HOME/.wire/wnsd-lite")

func main() {
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "wnsd-lite",
		Short: "WNS Lite",
	}

	rootCmd.PersistentFlags().String("chain-id", "wireline", "Chain identifier")
	rootCmd.PersistentFlags().String("log-level", "debug", "Log level")
	rootCmd.PersistentFlags().StringP("node", "n", "tcp://localhost:26657", "Upstream WNS node RPC address")
	rootCmd.PersistentFlags().String("log-file", "", "File to tail for GQL 'getLogs' API")

	rootCmd.AddCommand(versionCmd, initCmd, startCmd)

	executor := cli.PrepareBaseCmd(rootCmd, "NSL", DefaultLightNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
