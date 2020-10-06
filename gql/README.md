# WNS GQL Server

## Web UI Queries

Basic node status:

```graphql
{
  getStatus {
    version
    node {
      id
      network
      moniker
    }
    sync {
      latest_block_height
      catching_up
    }
    num_peers
    peers {
      is_outbound
      remote_ip
    }
    disk_usage
  }
}
```

Full node status:

```graphql
{
  getStatus {
    version
    node {
      id
      network
      moniker
    }
    sync {
      latest_block_hash
      latest_block_time
      latest_block_height
      catching_up
    }
    validator {
      address
      voting_power
      proposer_priority
    }
    validators {
      address
      voting_power
      proposer_priority
    }
    num_peers
    peers {
      node {
        id
        network
        moniker
      }
      is_outbound
      remote_ip
    }
    disk_usage
  }
}
```

Get account details:

```graphql
{
  getAccounts(addresses: ["cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094"]) {
    address
    pubKey
    number
    sequence
    balance {
      type
      quantity
    }
  }
}
```

Get bonds by IDs.

```graphql
{
  getBondsByIds(
    ids: ["8e340dd7cf6fc91c27eeefce9cca1406c262e93fd6f3a4f3b1e99b01161fcef3"]
  ) {
    id
    owner
    balance {
      type
      quantity
    }
  }
}
```

Query bonds:

```graphql
{
  queryBonds(
    attributes: [
      {
        key: "owner"
        value: { string: "cosmos1wh8vvd0ymc5nt37h29z8kk2g2ays45ct2qu094" }
      }
    ]
  ) {
    id
    owner
    balance {
      type
      quantity
    }
  }
}
```

Get records by IDs.

```graphql
{
  getRecordsByIds(ids: ["QmYDtNCKtTu6u6jaHaFAC5PWZXcj7fAmry6NoWwMaixFHz"]) {
    id
    type
    name
    version
  }
}
```

Query records.

```graphql
{
  queryRecords(attributes: [{ key: "type", value: { string: "wrn:bot" } }]) {
    id
    type
    name
    version
    bondId
    createTime
    expiryTime
    owners
    attributes {
      key
      value {
        string
      }
    }
  }
}
```

Resolve records.

```graphql
{
  resolveRecords(refs: ["wrn:bot:wireline.io/echo-bot"]) {
    id
    type
    name
    version
  }
}
```
