#!/bin/bash

LOG=/tmp/wns.log
API_ENDPOINT=http://localhost:9473/graphql
GQL_PLAYGROUND_API_BASE=""

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --tail)
    TAIL_LOGS=1
    shift
    ;;
    --gql-playground-api-base)
    GQL_PLAYGROUND_API_BASE="$2"
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

  rm -f ${LOG}

  # Start the server.
  nohup wnsd start --gql-server --gql-playground --gql-playground-api-base "${GQL_PLAYGROUND_API_BASE}" --log-file "${LOG}" > ${LOG} 2>&1 &

  if [[ ! -z "${TAIL_LOGS}" ]]; then
    log
  fi
}

function stop_server ()
{
  set -x
  killall wnsd
}

function log ()
{
  echo
  echo "Log file: ${LOG}"
  echo

  sleep 5

  tail -f ${LOG}
}

function test ()
{
  set -x
  curl -s -X POST -H "Content-Type: application/json" -d '{ "query": "{ getStatus { version } }" }' ${API_ENDPOINT} | jq
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
