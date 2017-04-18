#!/usr/bin/env bash

declare -a errors

function error_and_proceed() {
  errors+=("${1}")
  echo -e "tfenv: ${0}: Test Failed: ${1}" >&2
}

function error_and_die() {
  echo -e "tfenv: ${0}: ${1}" >&2
  exit 1
}

[ -n "${TFENV_DEBUG}" ] && set -x
source $(dirname $0)/helpers.sh \
  || error_and_die "Failed to load test helpers: $(dirname $0)/helpers.sh"

echo "### Uninstall local versions"
cleanup || error_and_die "Cleanup failed?!"

for v in 0.1.0; do
  tfenv install ${v} || error_and_proceed "Install of version ${v} failed"
done

tfenv uninstall 0.1.0 || error_and_proceed "Uninstall of version ${v} failed"

tfenv list | grep 0.1.0 && exit 1 || exit 0
if [ ${#errors[@]} -gt 0 ]; then
  echo -e "\033[0;31m===== The following list tests failed =====\033[0;39m" >&2
  for error in "${errors[@]}"; do
    echo -e "\t${error}"
  done
  exit 1
else
  echo -e "\033[0;32mAll list tests passed.\033[0;39m"
fi;
exit 0
