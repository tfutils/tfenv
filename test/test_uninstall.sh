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

v=0.9.1
(
  tfenv install ${v} || exit 1
  tfenv uninstall ${v} || exit 1
  check_version ${v} && exit 1 || exit 0
) || error_and_proceed "Uninstall of version ${v} failed"

echo "### Uninstall latest version"
cleanup || error_and_die "Cleanup failed?!"

v=$(tfenv list-remote | head -n 1)
(
  tfenv install latest || exit 1
  tfenv uninstall latest || exit 1
  check_version ${v} && exit 1 || exit 0
) || error_and_proceed "Uninstalling latest version ${v}"

echo "### Uninstall latest version with Regex"
cleanup || error_and_die "Cleanup failed?!"

v=$(tfenv list-remote | grep 0.8 | head -n 1)
(
  tfenv install latest:^0.8 || exit 1
  tfenv uninstall latest:^0.8 || exit 1
  check_version ${v} && exit 1 || exit 0
) || error_and_proceed "Uninstalling latest version ${v} with Regex"

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
