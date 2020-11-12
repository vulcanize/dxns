//
// Copyright 2020 Wireline, Inc.
//

package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	tmlite "github.com/tendermint/tendermint/lite"
	tmliteErr "github.com/tendermint/tendermint/lite/errors"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Note: Verifier code based on ~/go/pkg/mod/github.com/cosmos/cosmos-sdk@v0.37.0/client/context/query.go.

// TODO(ashwin): Determine appropriate cache size.
const cacheSize = 10

// CreateVerifier creates a light client verifier.
func CreateVerifier(config *Config) tmlite.Verifier {
	chainID := config.ChainID
	home := config.Home
	nodeAddress := config.NodeAddress

	node, err := rpcclient.New(nodeAddress, "/websocket")
	if err != nil {
		fmt.Printf("Verifier creation failed: %s\n", err.Error())
		fmt.Printf("Please check network connection and verify the address of the DXNS node to connect to.\n")
		os.Exit(1)
	}

	verifier, err := tmliteProxy.NewVerifier(
		chainID, filepath.Join(home, "data", ".lite_verifier"),
		node, log.NewNopLogger(), cacheSize,
	)

	if err != nil {
		fmt.Printf("Verifier creation failed: %s\n", err.Error())
		fmt.Printf("Please check network connection and verify the address of the DXNS node to connect to.\n")
		os.Exit(1)
	}

	return verifier
}

// Verify verifies the consensus proof at given height.
func Verify(ctx *Context, height int64) (tmtypes.SignedHeader, error) {
	check, err := tmliteProxy.GetCertifiedCommit(height, ctx.PrimaryNode.Client, ctx.verifier)
	switch {
	case tmliteErr.IsErrCommitNotFound(err):
		return tmtypes.SignedHeader{}, ErrVerifyCommit(height)
	case err != nil:
		return tmtypes.SignedHeader{}, err
	}

	return check, nil
}

// VerifyProof verifies the ABCI response.
func VerifyProof(ctx *Context, queryPath string, resp abci.ResponseQuery) error {
	if ctx.verifier == nil {
		return fmt.Errorf("missing valid certifier to verify data from distrusted node")
	}

	// The AppHash for height H is in header H+1.
	commit, err := Verify(ctx, resp.Height+1)
	if err != nil {
		return err
	}

	prt := rootmulti.DefaultProofRuntime()

	storeName, err := parseQueryStorePath(queryPath)
	if err != nil {
		return err
	}

	kp := merkle.KeyPath{}
	kp = kp.AppendKey([]byte(storeName), merkle.KeyEncodingURL)
	kp = kp.AppendKey(resp.Key, merkle.KeyEncodingURL)

	if resp.Value == nil {
		err = prt.VerifyAbsence(resp.Proof, commit.Header.AppHash, kp.String())
		if err != nil {
			return errors.Wrap(err, "failed to prove merkle proof")
		}

		return nil
	}

	err = prt.VerifyValue(resp.Proof, commit.Header.AppHash, kp.String(), resp.Value)
	if err != nil {
		return errors.Wrap(err, "failed to prove merkle proof")
	}

	return nil
}

// ErrVerifyCommit returns a common error reflecting that the blockchain commit at a given
// height can't be verified. The reason is that the base checkpoint of the certifier is
// newer than the given height
func ErrVerifyCommit(height int64) error {
	return fmt.Errorf(`The height of base truststore in the light client is higher than height %d.
Can't verify blockchain proof at this height`, height)
}

// parseQueryStorePath expects a format like /store/<storeName>/key.
func parseQueryStorePath(path string) (storeName string, err error) {
	if !strings.HasPrefix(path, "/") {
		return "", errors.New("expected path to start with /")
	}

	paths := strings.SplitN(path[1:], "/", 3)
	switch {
	case len(paths) != 3:
		return "", errors.New("expected format like /store/<storeName>/key")
	case paths[0] != "store":
		return "", errors.New("expected format like /store/<storeName>/key")
	case paths[2] != "key":
		return "", errors.New("expected format like /store/<storeName>/key")
	}

	return paths[1], nil
}
