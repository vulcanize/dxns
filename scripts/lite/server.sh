#!/bin/bash

LOG="/tmp/dxnsd-lite.log"
GQL_SERVER_PORT="9475"
GQL_PLAYGROUND_API_BASE=""
WNS_NODE_ADDRESS="tcp://localhost:26657"
WNS_GQL_ENDPOINT=""
RESET=
SCRIPT_DIR="$(dirname "$0")"
SYNC_TIMEOUT=10
CHAIN_ID="vulcanize-1"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --node)
    WNS_NODE_ADDRESS="$2"
    shift
    shift
    ;;
    --gql-port)
    GQL_SERVER_PORT="$2"
    shift
    shift
    ;;
    --gql-playground-api-base)
    GQL_PLAYGROUND_API_BASE="$2"
    shift
    shift
    ;;
    --log)
    LOG="$2"
    shift
    shift
    ;;
    --tail)
    TAIL_LOGS=1
    shift
    ;;
    --reset)
    RESET=1
    shift
    ;;
    --endpoint)
    WNS_GQL_ENDPOINT="$2"
    shift
    shift
    ;;
    --sync-timeout)
    SYNC_TIMEOUT="$2"
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

function start_server ()
{
  stop_server
  set -x

  rm -f "${LOG}"

  if [[ ! -z "${RESET}" ]]; then
    /bin/bash "${SCRIPT_DIR}/setup.sh" --node "${WNS_NODE_ADDRESS}" --reset
  fi

  # Start the server.
  nohup dxnsd-lite start \
    --chain-id "${CHAIN_ID}" \
    --gql-port "${GQL_SERVER_PORT}" \
    --gql-playground-api-base "${GQL_PLAYGROUND_API_BASE}" \
    --node "${WNS_NODE_ADDRESS}" \
    --endpoint "${WNS_GQL_ENDPOINT}" \
    --sync-timeout ${SYNC_TIMEOUT} \
    --log-file "${LOG}" \
    --log-level debug > "${LOG}" 2>&1 &

  if [[ ! -z "${TAIL_LOGS}" ]]; then
    log
  fi
}

function stop_server ()
{
  set -x
  killall dxnsd-lite
}

function log ()
{
  echo
  echo "Log file: ${LOG}"
  echo

  tail -f "${LOG}"
}

function test ()
{
  set -x
  curl -s -X POST -H "Content-Type: application/json" -d '{ "query": "{ getStatus { version } }" }' "http://localhost:${GQL_SERVER_PORT}/graphql" | jq
}

function command ()
{
  case $1 in
    start ) start_server; exit;;
    stop ) stop_server; exit;;
    log ) log; exit;;
    test ) test; exit;;
  esac
}

command=$1
if [[ ! -z "$command" ]]; then
  command $1
  exit
fi

select oper in "start" "stop" "log" "test"; do
  command $oper
  exit
done
