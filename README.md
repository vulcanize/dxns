# DXOS Naming Service

The DXOS Naming Service (DXNS) is a custom blockchain built using Cosmos SDK.

| Module   | Status | Public URL |
| -------- | ------ | ---------- |
| DXOS DOCS DXNS | [![Netlify Status](https://api.netlify.com/api/v1/badges/6bbab0ad-84ad-4d77-a575-420940dc55af/deploy-status)](https://app.netlify.com/sites/dxos-docs-wns/deploys) | https://dxos-docs-wns.netlify.app/wns/ |

## Getting Started

### Installation

* [Install golang](https://golang.org/doc/install) 1.14.0+ for the required platform.
* Ensure that the version installed is older than 1.16 as `crypto/hmac` panics while using `dxnscli` in older versions.
* Test that the correct version of `golang` has been successfully installed on the machine.

```bash
$ go version
go version go1.14.9 darwin/amd64
```

Set the followin ENV variables (if `go mod` has never been used on the machine).

```bash
mkdir -p $HOME/go/bin
echo "export GOPATH=$HOME/go" >> ~/.profile
echo "export GOBIN=\$GOPATH/bin" >> ~/.profile
echo "export PATH=\$PATH:\$GOBIN" >> ~/.profile
echo "export GO111MODULE=on" >> ~/.profile
source ~/.profile
```

Clone the repo then build and install the binaries.

```bash
$ git clone git@github.com:wirelineio/dxns.git
$ cd dxns
$ make install
```

Test that the following commands work:

```bash
$ dxnsd help
$ dxnscli help
```

### Initializing the Local Node

```bash
$ ./scripts/setup.sh
```

### Working with the Local Node

Start the node:

```bash
$ ./scripts/server.sh start
```

Test if the node is up:

```bash
$ ./scripts/server.sh test
```

View the logs:

```bash
$ ./scripts/server.sh log
```

Stop the node:

```bash
$ ./scripts/server.sh stop
```

## DXNS CLI

`wire` CLI provides [commands](https://github.com/wirelineio/incubator/blob/master/dxos/wns-cli/README.md) for publishing and querying DXNS records.

## Tests

See https://github.com/wirelineio/registry-client#tests


## GQL Server API

The GQL server is controlled using the following `dxnsd` flags:

* `--gql-server` - Enable GQL server (Available at http://localhost:9473/graphql).
* `--gql-playground` - Enable GQL playground app (Available at http://localhost:9473/console).
* `--gql-port` - Port to run the GQL server on (default 9473).

See `dxnsd/gql/schema.graphql` for the GQL schema.


## References

* https://golang.org/doc/install
* https://github.com/cosmos/cosmos-sdk
* https://cosmos.network/docs/tutorial/
* https://github.com/cosmos/sdk-application-tutorial
