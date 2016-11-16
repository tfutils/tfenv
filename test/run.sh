#!/usr/bin/env bash

if [ -n "$TFENV_DEBUG" ]; then
  export PS4='+ [${BASH_SOURCE##*/}:${LINENO}] '
  set -x
fi
export TFENV_ROOT=$(cd $(dirname $0)/.. && pwd)
export PATH="${TFENV_ROOT}/bin:${PATH}"

errors=()
if [ $# -ne 0 ];then
  targets="$@"
else
  targets=`ls $(dirname $0) | grep 'test_'`
fi

for t in ${targets}; do
  bash $(dirname $0)/${t} || errors+=( ${t} )
done

if [ ${#errors[@]} -ne 0 ];then
  echo -e "\033[0;31m===== Following tests failed ====\033[0;39m" 1>&2
  echo "${errors[@]}"
  exit 1
fi
