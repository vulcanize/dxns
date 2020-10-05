# WNS

## Clear Remote WNS

To clear a remote WNS, the following information is required:

* The RPC endpoint of the remote WNS (e.g. see https://github.com/wirelineio/dxns#testnets).
* The mnemonic for an account that has funds on the WNS.

The following example will work for https://wns-testnet.dev.wireline.ninja/console.

Create an account on a different machine (e.g. laptop/desktop), using the mnemonic for the remote `root` account.

```
$ wnscli keys add root-testnet-dev --recover
# Enter a passphrase for the new account, repeat it when prompted.
# Use the following mnemonic for recovery:
# salad portion potato insect unknown exile lion soft layer evolve flavor hollow emerge celery ankle sponsor easy effort flush furnace life maximum rotate apple
```

Clear the remote WNS using the following command:

```
$ wnscli tx nameservice clear --from root-testnet-dev --node tcp://wns-testnet.dev.wireline.ninja:26657
# Enter passphrase when prompted.
```

Use the GQL playground (https://wns-testnet.dev.wireline.ninja/console) to query and confirm that all records are gone.

## Invariant Checking

To turn on invariant checking, pass the `inv-check-period` flag to the server.

```bash
# Check invariants every N (here, N=1) blocks.
$ wnsd start --gql-server --gql-playground --inv-check-period=1
```

Also see https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/invariants.md.

## Export State

* Halt the chain.
* Run `wnsd export > state.json`.
* `state.json` contains the state of all modules (e.g. records and bonds).

## Import State

* Reset the chain state using `wnsd unsafe-reset-all`.
* Replace `~/.wire/wnsd/config/genesis.json` with a previously exported state (e.g. `state.json`).
* Start the chain.

## Denominations/Units

* `wire`  // 1 (base denom unit).
* `mwire` // 10^-3 (milli).
* `uwire` // 10^-6 (micro).

Also see `x/bond/internal/types/init.go`.

## Module Accounts

* `bond`: Module account for bonds. Balance reflects current total bonded amount.
* `record_rent`: Module account for record rent collection. Distribution of proceeds from rent collection (to be implemented) will reduce this amount.

```bash
$ wnscli query bond balance
{
  "bond": [
    {
      "denom": "uwire",
      "amount": "5993000000"
    }
  ],
  "record_rent": [
    {
      "denom": "uwire",
      "amount": "7000000"
    }
  ]
}
```
