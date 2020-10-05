#!/bin/bash

LOG=/tmp/wns-tunnel.log

function start_tunnel ()
{
  if [[ -z "$1" ]]; then
    echo "Usage: ./script/tunnel.sh start <user@remote-host> [--tail]" && exit
  fi

  stop_tunnel
  set -x

  rm -f ${LOG}

  # Start the tunnel (see https://www.everythingcli.org/ssh-tunnelling-for-fun-and-profit-autossh for details).
  nohup autossh -M 0 -o "ServerAliveInterval 30" -o "ServerAliveCountMax 3" -vvv -nNT -R 26656:localhost:26656 "${1}" > ${LOG} 2>&1 &

  if [[ $2 = "--tail" ]]; then
    log
  fi
}

function stop_tunnel ()
{
  set -x
  killall autossh
}

function log ()
{
  echo
  echo "Log file: ${LOG}"
  echo

  tail -f ${LOG}
}

function command ()
{
  case $1 in
    start ) start_tunnel $2 $3; exit;;
    stop ) stop_tunnel; exit;;
    log ) log; exit;;
  esac
}

command=$1
if [[ ! -z "$command" ]]; then
  command $1 $2 $3
  exit
fi

select oper in "start" "stop" "log"; do
  command $oper
  exit
done
