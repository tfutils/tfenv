#!/usr/bin/env bash

if [ -n "${TFENV_DEBUG}" ]; then
  export PS4='+ [${BASH_SOURCE##*/}:${LINENO}] '
  set -x
fi
TFENV_ROOT=$(cd $(dirname ${0})/.. && pwd)
export PATH="${TFENV_ROOT}/bin:${PATH}"

errors=()
if [ ${#} -ne 0 ];then
  targets="${@}"
else
  targets=$(\ls $(dirname ${0}) | grep 'test_')
fi

for t in ${targets}; do
  bash $(dirname ${0})/${t} || errors+=( ${t} )
done

if [ ${#errors[@]} -ne 0 ];then
  echo -e "\033[0;31m===== The following test suites failed =====\033[0;39m" >&2
  for error in "${errors[@]}"; do
    echo -e "\t${error}" >&2
  done
  exit 1
else
  echo -e "\033[0;32mAll test suites passed.\033[0;39m"
fi
exit 0
