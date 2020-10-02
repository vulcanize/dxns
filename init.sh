#!/bin/bash
rm -r ~/.dxnscli
rm -r ~/.dxnsd

dxnsd init mynode --chain-id dxns-1

dxnscli config keyring-backend test

dxnscli keys add me
dxnscli keys add you

dxnsd add-genesis-account $(dxnscli keys show me -a) 3500000000000000000aphoton,100000000stake
dxnsd add-genesis-account $(dxnscli keys show you -a) 1000000000000000000aphoton

dxnscli config chain-id dxns-1
dxnscli config output json
dxnscli config indent true
dxnscli config trust-node true

dxnsd gentx --name me --keyring-backend test
dxnsd collect-gentxs