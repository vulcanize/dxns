# Bonds Demo

## Setup

Setup the machine as documented in https://github.com/wirelineio/dxns#setup-machine.

WNS:

```bash
# Clone `wns` repo.
$ git clone git@github.com:wirelineio/dxns.git
$ cd wns

# Build and install the binaries.
$ make install
```

Install the latest wire CLI:

```bash
$ yarn global add @wirelineio/cli @wirelineio/cli-peer @wirelineio/cli-bot @wirelineio/cli-pad

$ wire --version
0.6.1
```

### Blockchain

```bash
# Delete old folders.
$ rm -rf ~/.wire/wnsd ~/.wire/wnscli

# Init the chain.
$ wnsd init my-node --chain-id wireline
```

```bash
# Update genesis params.
# Note: On Linux, use just `-i` instead of `-i ''`.

# Change staking token to uwire.
$ sed -i '' 's/stake/uwire/g' ~/.wire/wnsd/config/genesis.json

# Change gov proposal pass timeout to 5 mins.
$ sed -i '' 's/172800000000000/300000000000/g' ~/.wire/wnsd/config/genesis.json

# Change max bond amount.
$ sed -i '' 's/10wire/10000wire/g' ~/.wire/wnsd/config/genesis.json
```

```bash
# Create root accounts/keys.
$ echo "temp12345\nsalad portion potato insect unknown exile lion soft layer evolve flavor hollow emerge celery ankle sponsor easy effort flush furnace life maximum rotate apple" | wnscli keys add root --recover

# Use the same mnemonic to generate the private key for use with the `wire` CLI.
$ wire keys generate --mnemonic="salad portion potato insect unknown exile lion soft layer evolve flavor hollow emerge celery ankle sponsor easy effort flush furnace life maximum rotate apple"
Mnemonic:  salad portion potato insect unknown exile lion soft layer evolve flavor hollow emerge celery ankle sponsor easy effort flush furnace life maximum rotate apple
Private key:  b1e4e95dd3e3294f15869b56697b5e3bdcaa24d9d0af1be9ee57d5a59457843a
Public key:  02ead12ad29c532364b2f7b565582499840fcf45a51f16542385072961f4df62d8
Address:  cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094

```bash
# Add genesis accounts to chain.
$ wnsd add-genesis-account $(wnscli keys show root -a) 100000000000000uwire
$ wnsd add-genesis-account $(wnscli keys show alice -a) 100000000000000uwire
$ wnsd add-genesis-account $(wnscli keys show bob -a) 100000000000000uwire
```

```bash
# CLI config.
$ wnscli config chain-id wireline
$ wnscli config output json
$ wnscli config indent true
$ wnscli config trust-node true
```

```bash
# Setup genesis transactions.
$ echo "temp12345" | wnsd gentx --name root --amount 10000000000000uwire
$ wnsd collect-gentxs
$ wnsd validate-genesis
```

```bash
# Start the chain.
$ wnsd start --gql-server --gql-playground
```

## Run

Set ENV variables for WNS endpoint and private key.

```bash
export WIRE_WNS_ENDPOINT='http://localhost:9473/graphql'
export WIRE_WNS_USER_KEY="b1e4e95dd3e3294f15869b56697b5e3bdcaa24d9d0af1be9ee57d5a59457843a"
```

Create bonds.

```bash
$ wire wns create-bond --type uwire --quantity 1000000000
$ wire wns create-bond --type uwire --quantity 1000000000
```

List bonds.

```bash
# Note: Bond ID, owner and balance.
$ wire wns list-bonds
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000000"
            }
        ]
    },
    {
        "id": "e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000000"
            }
        ]
    }
]
```

List balance across all bonds.

```bash
$ wnscli query bond balance
{
  "bond": [
    {
      "denom": "uwire",
      "amount": "2000000000"
    }
  ]
}
```

Get bond by ID.

```bash
$ wire wns get-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000000"
            }
        ]
    }
]
```

Query bonds by owner.

```bash
# Uses a secondary index: Owner -> Bond ID.
$ wire wns list-bonds --owner cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000000"
            }
        ]
    },
    {
        "id": "e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000000"
            }
        ]
    }
]
```

Refill bond.

```bash
$ wire wns refill-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3 --type uwire --quantity 1000
```

```bash
$ wire wns get-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000001000"
            }
        ]
    }
]
```

Withdraw funds from bond.

```bash
# Transfers the funds back into the bond owner account.
$ wire wns withdraw-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3 --type uwire --quantity 500

$ wire wns get-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "1000000500"
            }
        ]
    }
]
```

Publish records (w/ bond).

```bash
$ cd x/nameservice/examples
$ wire wns publish --filename protocol.yml --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
$ wire wns publish --filename bot.yml --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3

# Note: bondID and expiryTime attributes on the records.
$ wire wns list-records
[
    {
        "id": "QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i",
        "type": "wrn:bot",
        "name": "wireline.io/chess-bot",
        "version": "2.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:55.824917000",
        "attributes": {
            "accessKey": "7db6a0c2b8bc79e733612b5cfd45e9f69bec4d05d424076826a0d08a2a62641c",
            "displayName": "ChessBot",
            "name": "wireline.io/chess-bot",
            "protocol": {
                "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe"
            },
            "type": "wrn:bot",
            "version": "2.0.0"
        }
    },
    {
        "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe",
        "type": "wrn:protocol",
        "name": "wireline.io/chess",
        "version": "1.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:00.161930000",
        "attributes": {
            "displayName": "Chess",
            "name": "wireline.io/chess",
            "type": "wrn:protocol",
            "version": "1.0.0"
        }
    }
]

# Note: Rent has been deducted from the bond.
$ wire wns get-bond --id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "998000500"
            }
        ]
    }
]

# Note: Check balance of bond and record rent module accounts.
$ wnscli query bond balance
{
  "record_rent": [
    {
      "denom": "uwire",
      "amount": "2000000"
    }
  ],
  "bond": [
    {
      "denom": "uwire",
      "amount": "1998000500"
    }
  ]
}
```

List records by bond.

```bash
# Note: Uses a secondary index: Bond ID -> Record ID.
$ wire wns list-records --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i",
        "type": "wrn:bot",
        "name": "wireline.io/chess-bot",
        "version": "2.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:55.824917000",
        "attributes": {
            "version": "2.0.0",
            "accessKey": "7db6a0c2b8bc79e733612b5cfd45e9f69bec4d05d424076826a0d08a2a62641c",
            "displayName": "ChessBot",
            "name": "wireline.io/chess-bot",
            "protocol": {
                "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe"
            },
            "type": "wrn:bot"
        }
    },
    {
        "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe",
        "type": "wrn:protocol",
        "name": "wireline.io/chess",
        "version": "1.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:00.161930000",
        "attributes": {
            "displayName": "Chess",
            "name": "wireline.io/chess",
            "type": "wrn:protocol",
            "version": "1.0.0"
        }
    }
]
```

Dissociate bond from record.

```bash
$ wire wns dissociate-bond --id QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i

$ wire wns list-records --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe",
        "type": "wrn:protocol",
        "name": "wireline.io/chess",
        "version": "1.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:00.161930000",
        "attributes": {
            "name": "wireline.io/chess",
            "type": "wrn:protocol",
            "version": "1.0.0",
            "displayName": "Chess"
        }
    }
]
```

Associate bond with record.

```bash
$ wire wns associate-bond --id QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3

$ wire wns list-records --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[
    {
        "id": "QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i",
        "type": "wrn:bot",
        "name": "wireline.io/chess-bot",
        "version": "2.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:55.824917000",
        "attributes": {
            "displayName": "ChessBot",
            "name": "wireline.io/chess-bot",
            "protocol": {
                "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe"
            },
            "type": "wrn:bot",
            "version": "2.0.0",
            "accessKey": "7db6a0c2b8bc79e733612b5cfd45e9f69bec4d05d424076826a0d08a2a62641c"
        }
    },
    {
        "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe",
        "type": "wrn:protocol",
        "name": "wireline.io/chess",
        "version": "1.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "expiryTime": "2021-01-01T11:40:00.161930000",
        "attributes": {
            "displayName": "Chess",
            "name": "wireline.io/chess",
            "type": "wrn:protocol",
            "version": "1.0.0"
        }
    }
]
```

Reassociate bond.

```bash
$ wire wns reassociate-records --old-bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3 --new-bond-id e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d

$ wire wns list-records --bond-id 8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3
[]

# Check both records are now associated with bond e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d.
$ wire wns list-records --bond-id e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d
[
    {
        "id": "QmNgCCwB2AGQADe1X4P1kVjd1asdTandaRHbRp1fKrDH9i",
        "type": "wrn:bot",
        "name": "wireline.io/chess-bot",
        "version": "2.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d",
        "expiryTime": "2021-01-01T11:40:55.824917000",
        "attributes": {
            "name": "wireline.io/chess-bot",
            "protocol": {
                "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe"
            },
            "type": "wrn:bot",
            "version": "2.0.0",
            "accessKey": "7db6a0c2b8bc79e733612b5cfd45e9f69bec4d05d424076826a0d08a2a62641c",
            "displayName": "ChessBot"
        }
    },
    {
        "id": "Qmb8beMVtdjQpRiN9bFLUsisByun8EHPKPo2g1jSUUnHqe",
        "type": "wrn:protocol",
        "name": "wireline.io/chess",
        "version": "1.0.0",
        "owners": [
            "6ee3328f65c8566cd5451e49e97a767d10a8adf7"
        ],
        "bondId": "e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d",
        "expiryTime": "2021-01-01T11:40:00.161930000",
        "attributes": {
            "type": "wrn:protocol",
            "version": "1.0.0",
            "displayName": "Chess",
            "name": "wireline.io/chess"
        }
    }
]
```

Dissociate bond from all records.

```bash
$ wire wns dissociate-records --bond-id e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d

# Note: No records found.
$ wire wns list-records --bond-id e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d
[]
```

Cancel bond.

```bash
# Note: Cancel works if bond doesn't have associated records.
# Note: Cancel fails if there are associated records.
# Note: Cancelled bond is deleted.
$ wire wns cancel-bond --id e205a46f6ec6f662cbfad84f4f926973422bf6217d8d2c2eebff94d148fd486d

$ wire wns list-bonds --owner cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094
[
    {
        "id": "8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3",
        "owner": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
        "balance": [
            {
                "type": "uwire",
                "quantity": "998000500"
            }
        ]
    }
]
```

Consensus params.

```bash
$ wnscli query bond params
{
  "max_bond_amount": "10000wire"
}

$ wnscli query nameservice params
{
  "record_rent": "1wire",
  "record_expiry_time": "31536000000000000"
}
```

Create proposal to change param.

```bash
$ echo temp12345 | wnscli tx gov submit-proposal param-change params/update_max_bond_amount.json --from root --yes -b block
```

Query proposals.

```bash
# Note: Proposal status = DepositPeriod.
$ wnscli query gov proposals
[
  {
    "content": {
      "type": "cosmos-sdk/ParameterChangeProposal",
      "value": {
        "title": "Bond Param Change",
        "description": "Update max bond amount.",
        "changes": [
          {
            "subspace": "bond",
            "key": "MaxBondAmount",
            "value": "\"15000wire\""
          }
        ]
      }
    },
    "id": "1",
    "proposal_status": "DepositPeriod",
    "final_tally_result": {
      "yes": "0",
      "abstain": "0",
      "no": "0",
      "no_with_veto": "0"
    },
    "submit_time": "2019-12-18T11:41:20.691337Z",
    "deposit_end_time": "2019-12-18T11:46:20.691337Z",
    "total_deposit": [
      {
        "denom": "uwire",
        "amount": "1000000"
      }
    ],
    "voting_start_time": "0001-01-01T00:00:00Z",
    "voting_end_time": "0001-01-01T00:00:00Z"
  }
]
```

Deposit sufficient funds to move the proposal into the voting stage.

```bash
# Note: Proposal ID = 1.
$ echo temp12345 | wnscli tx gov deposit 1 10000000uwire --from root -yes -b block

# Note: Proposal status = VotingPeriod.
$ wnscli query gov proposals
[
  {
    "content": {
      "type": "cosmos-sdk/ParameterChangeProposal",
      "value": {
        "title": "Bond Param Change",
        "description": "Update max bond amount.",
        "changes": [
          {
            "subspace": "bond",
            "key": "MaxBondAmount",
            "value": "\"15000wire\""
          }
        ]
      }
    },
    "id": "1",
    "proposal_status": "VotingPeriod",
    "final_tally_result": {
      "yes": "0",
      "abstain": "0",
      "no": "0",
      "no_with_veto": "0"
    },
    "submit_time": "2019-12-18T11:41:20.691337Z",
    "deposit_end_time": "2019-12-18T11:46:20.691337Z",
    "total_deposit": [
      {
        "denom": "uwire",
        "amount": "11000000"
      }
    ],
    "voting_start_time": "2019-12-18T11:42:46.695024Z",
    "voting_end_time": "2019-12-18T11:47:46.695024Z"
  }
]
```

Vote (yes) on proposal.

```bash
$ echo temp12345 | wnscli tx gov vote 1 yes --from root --yes -b block
```

Check votes.

```bash
$ wnscli query gov votes 1
[
  {
    "proposal_id": "1",
    "voter": "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094",
    "option": "Yes"
  }
]

$ wnscli query gov tally 1
{
  "yes": "10000000000000",
  "abstain": "0",
  "no": "0",
  "no_with_veto": "0"
}
```

Wait for 5 mins. Proposal enters `Passed` status.

```bash
# Note: Proposal status = Passed.
$ wnscli query gov proposals
[
  {
    "content": {
      "type": "cosmos-sdk/ParameterChangeProposal",
      "value": {
        "title": "Bond Param Change",
        "description": "Update max bond amount.",
        "changes": [
          {
            "subspace": "bond",
            "key": "MaxBondAmount",
            "value": "\"15000wire\""
          }
        ]
      }
    },
    "id": "1",
    "proposal_status": "Passed",
    "final_tally_result": {
      "yes": "10000000000000",
      "abstain": "0",
      "no": "0",
      "no_with_veto": "0"
    },
    "submit_time": "2019-12-18T11:41:20.691337Z",
    "deposit_end_time": "2019-12-18T11:46:20.691337Z",
    "total_deposit": [
      {
        "denom": "uwire",
        "amount": "11000000"
      }
    ],
    "voting_start_time": "2019-12-18T11:42:46.695024Z",
    "voting_end_time": "2019-12-18T11:47:46.695024Z"
  }
]
```

Check updated value of param.

```bash
$ wnscli query bond params
{
  "max_bond_amount": "15000wire"
}
```
