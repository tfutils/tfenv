#!/usr/bin/env bash
set -uo pipefail;

if [[ $(uname) == 'Darwin' ]] && [ $(which brew) ]; then
  brew install grep;
  if [[ -n ${GITHUB_PATH+x} ]]; then
    echo "/opt/homebrew/opt/grep/libexec/gnubin" >> $GITHUB_PATH
  fi
fi;
