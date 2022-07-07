#!/usr/bin/env bash
set -uo pipefail;

if [[ $(uname) == 'Darwin' ]] && [ $(which brew) ]; then
  brew install grep;
fi;
