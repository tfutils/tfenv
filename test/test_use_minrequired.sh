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

[ -n "$TFENV_DEBUG" ] && set -x
source $(dirname $0)/helpers.sh \
  || error_and_die "Failed to load test helpers: $(dirname $0)/helpers.sh"

echo "### Install not min-required version"
cleanup || error_and_die "Cleanup failed?!"

v=0.2.0
(
  tfenv install ${v} || true
  tfenv use ${v} || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing specific version ${v}"

echo 'terraform {

  required_version = ">=0.1.0"
}' >> min_required.tf

tfenv install min-required
tfenv use min-required

check_version '0.1.0' || error_and_proceed "Min required version doesn't match"

cleanup || error_and_die "Cleanup failed?!"


if [ ${#errors[@]} -gt 0 ]; then
  echo -e "\033[0;31m===== The following install_and_use tests failed =====\033[0;39m" >&2
  for error in "${errors[@]}"; do
    echo -e "\t${error}"
  done
  exit 1
else
  echo -e "\033[0;32mAll install_and_use tests passed.\033[0;39m"
fi;
exit 0
