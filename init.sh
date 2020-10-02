#!/bin/bash
rm -r ~/.ethappcli
rm -r ~/.ethappd

ethappd init mynode --chain-id ethapp-1

ethappcli config keyring-backend test

ethappcli keys add me
ethappcli keys add you

ethappd add-genesis-account $(ethappcli keys show me -a) 3500000000000000000aphoton,100000000stake
ethappd add-genesis-account $(ethappcli keys show you -a) 1000000000000000000aphoton

ethappcli config chain-id ethapp-1
ethappcli config output json
ethappcli config indent true
ethappcli config trust-node true

ethappd gentx --name me --keyring-backend test
ethappd collect-gentxs