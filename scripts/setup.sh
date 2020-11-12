#!/bin/bash

#
# Initial set-up.
#

DEFAULT_MNEMONIC="salad portion potato insect unknown exile lion soft layer evolve flavor hollow emerge celery ankle sponsor easy effort flush furnace life maximum rotate apple"
DEFAULT_PASSPHRASE="12345678"

NODE_NAME=`hostname`
CHAIN_ID="wireline-1"
DENOM=uwire
KEYRING_BACKEND=test
WNS_CLI_CONFIG_DIR="${HOME}/.wire/dxnscli"
WNS_SERVER_CONFIG_DIR="${HOME}/.wire/dxnsd"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --reset)
    RESET=1
    shift
    ;;
    --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;
    --node-name)
    NODE_NAME="$2"
    shift
    shift
    ;;
    --mnemonic)
    MNEMONIC="$2"
    shift
    shift
    ;;
    --passphrase)
    PASSPHRASE="$2"
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

function init_secrets ()
{
  if [[ -z "${MNEMONIC}" ]]; then
    MNEMONIC="${DEFAULT_MNEMONIC}"
  fi

  if [[ -z "${PASSPHRASE}" ]]; then
    PASSPHRASE="${DEFAULT_PASSPHRASE}"
  fi
}

SED_ARGS=""

# On MacOS, sed needs `-i ''``. On Linux, just `-i`.
if [ "$(uname)" == "Darwin" ]; then
  SED_ARGS="''"
fi

function save_secrets ()
{
  mkdir -p ~/.wire
  echo "Root Account Mnemonic: ${MNEMONIC}" > ~/.wire/secrets
  echo "CLI Passphrase: ${PASSPHRASE}" >> ~/.wire/secrets
  echo "To generate wire CLI key:" >> ~/.wire/secrets
  echo "wire keys generate --mnemonic=\"<MNEMONIC>\"" >> ~/.wire/secrets
}

function reset ()
{
  killall -SIGKILL dxnsd
  rm -rf "${WNS_SERVER_CONFIG_DIR}"
  rm -rf "${WNS_CLI_CONFIG_DIR}"
}

function init_config ()
{
  # https://docs.cosmos.network/master/interfaces/keyring.html
  dxnscli config keyring-backend $KEYRING_BACKEND

  # Configure the CLI to eliminate the need for the chain-id flag.
  dxnscli config chain-id "${CHAIN_ID}"
  dxnscli config output json
  dxnscli config indent true
  dxnscli config trust-node true
}

function init_node ()
{
  # Init the chain.
  dxnsd init "${NODE_NAME}" --chain-id "${CHAIN_ID}"

  # Change the staking unit.
  sed -i $SED_ARGS "s/stake/${DENOM}/g" "${WNS_SERVER_CONFIG_DIR}/config/genesis.json"

  # TODO(ashwin): Patch genesis.json with max bond amount?
}

function init_root ()
{
  # Create a genesis validator account provisioned with 100 million WIRE.
  echo -e "${MNEMONIC}\n${PASSPHRASE}\n${PASSPHRASE}" | dxnscli keys add root --recover
  echo -e "${PASSPHRASE}" | dxnsd add-genesis-account $(dxnscli keys show root -a) 100000000000000uwire

  # Validator stake/bond => 10 million WIRE (out of total 100 million WIRE).
  echo -e "${PASSPHRASE}\n${PASSPHRASE}\n${PASSPHRASE}" | dxnsd gentx --name root --amount 10000000000000uwire --keyring-backend $KEYRING_BACKEND --home-client "${WNS_CLI_CONFIG_DIR}"
  dxnsd collect-gentxs
  dxnsd validate-genesis
}

#
# Options
#

if [[ ! -z "${RESET}" ]]; then
  reset
fi

# Test if installed already.
if [[ -d "${WNS_SERVER_CONFIG_DIR}" ]]; then
  echo "Do you wish to RESET?"
  select yn in "Yes" "No"; do
    case $yn in
      Yes ) reset; break;;
      No ) exit;;
    esac
  done
fi

init_secrets

init_config
init_node
init_root

save_secrets
