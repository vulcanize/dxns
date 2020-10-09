#!/bin/bash

#
# Initial set-up.
#

WNS_LITE_SERVER_CONFIG_DIR="${HOME}/.wire/dxnsd-lite"
CHAIN_ID="wireline-1"
WNS_NODE_ADDRESS="tcp://localhost:26657"
RESET=

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --reset)
    RESET=1
    shift
    ;;
    --node)
    WNS_NODE_ADDRESS="$2"
    shift
    shift
    ;;
    --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;
    *)
    POSITIONAL+=("$1")
    shift
    ;;
  esac
done
set -- "${POSITIONAL[@]}"

function reset ()
{
  killall -SIGTERM dxnsd-lite
  rm -rf "${WNS_LITE_SERVER_CONFIG_DIR}"
}

function init_node ()
{
  dxnsd-lite init --chain-id "${CHAIN_ID}" --from-node --node "${WNS_NODE_ADDRESS}"
}

if [[ ! -z "${RESET}" ]]; then
  reset
fi

# Test if installed already.
if [[ -d "${WNS_LITE_SERVER_CONFIG_DIR}" ]]; then
  echo "Do you wish to RESET?"
  select yn in "Yes" "No"; do
    case $yn in
      Yes ) reset; break;;
      No ) exit;;
    esac
  done
fi

init_node
