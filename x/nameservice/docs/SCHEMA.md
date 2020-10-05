# Record Schemas

WNS supports arbitrary record types. To aid developers, common/current records types and their schema is documented here.

## App

* `package` - IPFS CID of app package.
* `displayName` - application display name.

Example:

```json
  {
    "id": "Qmbjeafr7vj7G6X7Zta8nGNVDyrU5UaK93L3FJCuxiJoQf",
    "type": "wrn:app",
    "name": "d.boreh.am/dbeditor",
    "version": "0.0.16",
    "owners": [
      "cdc80cc5d5177a12690f198379ebceacaa121d0b"
    ],
    "bondId": "4572d23bb0fff82168fc4ec1f61e35ac796d7ea7266b713490f448eaf7063879",
    "createTime": "2020-04-30T22:28:53.879371213",
    "expiryTime": "2021-04-30T22:28:53.879371213",
    "attributes": {
      "id": "wrn:app:d.boreh.am/dbeditor",
      "name": "d.boreh.am/dbeditor",
      "package": "QmNg7dcEKAfv8DB355ekYUJDbRVKp27TF7oJ9zSxXoAMy7",
      "type": "wrn:app",
      "version": "0.0.16",
      "build": "yarn dist",
      "displayName": "DBDB Editor App"
    }
  }
```

## Bot

* `package` - IPFS CIDs of binary packages for multiple platforms.
* `displayName` - bot display name.

Example:

```json
  {
    "id": "QmRUBGFgxftNKoY15ght7bPwzePAdAqGXnhs9Zpy9A3NHU",
    "type": "wrn:bot",
    "name": "wireline.io/store",
    "version": "1.0.0",
    "owners": [
      "dd72cdc790dbead7f01a0ed04a2f62239e4f6963"
    ],
    "bondId": "b14dcf0db8bfc66b58b5a3d4534ba8450f47422f1724d9807085f25de4a0f3ef",
    "createTime": "2020-04-22T21:13:12.869149935",
    "expiryTime": "2021-04-22T21:13:12.869149935",
    "attributes": {
      "type": "wrn:bot",
      "version": "1.0.0",
      "displayName": "Store",
      "name": "wireline.io/store",
      "package": {
        "linux": {
          "x64": "QmUE6PALgZbKVT9uQwbqGtckjmo2kniFV2HCqg7B6nDbyB"
        },
        "macos": {
          "x64": "QmQqxJRD5yXxbazE5UEgDXKdmK7aFknummgm5xM58pFkm3"
        }
      }
    }
  }
```

## Resource

* `name` - Docker image name & tag.
* `docker.hash` - Docker image hash.
* `docker.url` - Docker Hub image URL.

Example:

```json
  {
    "id": "Qmcqgt95g1P8UGAcHpqifcGpyXB9EKak99mW99Jv1VTJyL",
    "type": "wrn:resource",
    "name": "dxos/xbox:devnet-unstable",
    "version": "0.1.29",
    "owners": [
      "3ffed5d3c00a0ddf30fa25f13c7bf18f57728b82"
    ],
    "bondId": "b2cc272774defbc134bfa4106d1bbe585ce91c138a6699c71cacb2d09911bf85",
    "createTime": "2020-04-30T20:56:40.297531621",
    "expiryTime": "2021-04-30T20:56:40.297531621",
    "attributes": {
      "version": "0.1.29",
      "docker": {
        "hash": "df9cb3caa90fa8240dbeb4c3497f74d19d545897ae256bcafc4d3976c5c7940e",
        "url": "https://hub.docker.com/layers/dxos/xbox/devnet-unstable-0.1.29/images/sha256-df9cb3caa90fa8240dbeb4c3497f74d19d545897ae256bcafc4d3976c5c7940e"
      },
      "name": "dxos/xbox:devnet-unstable",
      "type": "wrn:resource"
    }
  }
```

## XBox

`xbox` records are used to discover xbox nodes on the network, and describe services offered/exposed by that node.

### WNS

* `wns.rpc` - RPC endpoint of WNS full-node, used by lite nodes to sync state.

Example:

```json
  {
    "id": "QmbVymAeFw37mW7FEHnZmSsDs25PtBgXiYqxH5ciyzawoe",
    "type": "wrn:xbox",
    "name": "ashwinp/wns-good",
    "version": "0.0.1",
    "owners": [
      "233b436a205539f0f8082507e300fc5f3ca9eb0a"
    ],
    "bondId": "8a359128068c85f9982a36308772057d098f16dc21288e312205bdf60a6961e9",
    "createTime": "2020-04-22T09:20:11.652707970",
    "expiryTime": "2021-04-22T09:20:11.652707970",
    "attributes": {
      "name": "ashwinp/wns-good",
      "type": "wrn:xbox",
      "version": "0.0.1",
      "wns": {
        "rpc": "tcp://node1.dxos.network:26657"
      }
    }
  }
```

### Service

The `service` attribute must be included for a well-formed record, but `description` is optional.

If the service needs to define service-specific details (eg, endpoints, protocol version info), by convention they
should be contained in a tag that matches the `service` (eg, `service: "ipfs", "ipfs": { ... }`).

#### IPFS Example
```json
{
  "record": {
    "type": "wrn:service",
    "name": "example.com/services/ipfs",
    "version": "0.0.1",
    "service": "ipfs",
    "description": "Helpful description of this IPFS service.",
    "ipfs": {
        "protocol": "ipfs/0.1.0",
        "addresses": [
            "/ip4/192.168.123.56/tcp/4001/p2p/QmR5EQkRx4sLXV3vzgewe3UyXxZJXr4hwL2uwcTScrRtFE"
        ]
    }
  }
}
```
#### Signal Example
```json
{
  "record": {
    "type": "wrn:service",
    "name": "example.com/services/signal",
    "version": "0.0.1",
    "service": "signal",
    "description": "Helpful description of this signal service.",
    "signal": {
        "url": "wss://my.host.name/dxos/signal",
        "bootstrap": "my.host.name:4000"
    }
  }
}
```

#### STUN Example
```json
{
  "record": {
    "type": "wrn:service",
    "name": "example.com/services/stun",
    "version": "0.0.1",
    "service": "stun",
    "description": "Helpful description of this STUN service.",
    "stun": {
        "url": "stun:my.host.name:3478"
    }
  }
}
```

#### TURN Example
```json
{
  "record": {
    "type": "wrn:service",
    "name": "example.com/services/turn",
    "version": "0.0.1",
    "service": "turn",
    "description": "Helpful description of this TURN service.",
    "turn": {
        "url": "turn:my.host.name:3478",
        "username": "freeturn",
        "password": "freeturn"
    }
  }
}
```
