//
// Copyright 2020 Wireline, Inc.
//

package gql

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/rpc/core"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// NodeDataPath is the path to the wnsd data folder.
var NodeDataPath = os.ExpandEnv("$HOME/.wire/wnsd/data")

func getStatusInfo(ctx *rpctypes.Context) (*NodeInfo, *SyncInfo, *ValidatorInfo, error) {
	res, err := core.Status(ctx)

	if err != nil {
		return nil, nil, nil, err
	}

	nodeInfo := res.NodeInfo
	syncInfo := res.SyncInfo
	valInfo := res.ValidatorInfo

	return &NodeInfo{
			ID:      string(nodeInfo.ID()),
			Moniker: nodeInfo.Moniker,
			Network: nodeInfo.Network,
		}, &SyncInfo{
			LatestBlockHash:   syncInfo.LatestBlockHash.String(),
			LatestBlockHeight: strconv.FormatInt(syncInfo.LatestBlockHeight, 10),
			LatestBlockTime:   syncInfo.LatestBlockTime.UTC().String(),
			CatchingUp:        syncInfo.CatchingUp,
		}, &ValidatorInfo{
			Address:     valInfo.Address.String(),
			VotingPower: strconv.FormatInt(valInfo.VotingPower, 10),
		}, nil
}

func getNetInfo(ctx *rpctypes.Context) (string, []*PeerInfo, error) {
	res, err := core.NetInfo(ctx)
	if err != nil {
		return "", nil, err
	}

	peers := res.Peers
	peersInfo := make([]*PeerInfo, len(peers))
	for index, peer := range peers {
		peersInfo[index] = &PeerInfo{
			Node: &NodeInfo{
				ID:      string(peer.NodeInfo.ID()),
				Moniker: peer.NodeInfo.Moniker,
				Network: peer.NodeInfo.Network,
			},
			IsOutbound: peer.IsOutbound,
			RemoteIP:   peer.RemoteIP,
		}
	}

	return strconv.FormatInt(int64(res.NPeers), 10), peersInfo, nil
}

// GetDiskUsage returns disk usage for the given path.
func GetDiskUsage(dirPath string) (string, error) {
	out, err := exec.Command("du", "-sh", dirPath).Output()
	if err != nil {
		return "", err
	}

	return strings.Fields(string(out))[0], nil
}

// GetValidatorSet creates the validator set GQL response.
func GetValidatorSet(res *coretypes.ResultValidators) []*ValidatorInfo {
	validatorSet := make([]*ValidatorInfo, len(res.Validators))
	for index, validator := range res.Validators {
		proposerPriority := strconv.FormatInt(validator.ProposerPriority, 10)
		validatorSet[index] = &ValidatorInfo{
			Address:          validator.Address.String(),
			VotingPower:      strconv.FormatInt(validator.VotingPower, 10),
			ProposerPriority: &proposerPriority,
		}
	}

	return validatorSet
}

func getValidatorSet(ctx *rpctypes.Context) ([]*ValidatorInfo, error) {
	res, err := core.Validators(ctx, nil, 1, 100)

	if err != nil {
		return nil, err
	}

	return GetValidatorSet(res), nil
}
