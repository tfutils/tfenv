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

echo "### Install latest version"
cleanup || error_and_die "Cleanup failed?!"

v=$(tfenv list-remote | head -n 1)
(
  tfenv install latest || exit 1
  tfenv use ${v} || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing latest version ${v}"

echo "### Install latest version with Regex"
cleanup || error_and_die "Cleanup failed?!"

v=$(tfenv list-remote | grep 0.8 | head -n 1)
(
  tfenv install latest:^0.8 || exit 1
  tfenv use latest:^0.8 || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing latest version ${v} with Regex"

echo "### Install latest version with Regex"
cleanup

v=$(tfenv list-remote | grep 0.8 | head -n 1)
tfenv install latest:^0.8
tfenv use latest:^0.8
if ! check_version ${v}; then
  echo "Installing latest version ${v} with Regex" 1>&2
  exit 1
fi

echo "### Install specific version"
cleanup || error_and_die "Cleanup failed?!"

v=0.7.13
(
  tfenv install ${v} || exit 1
  tfenv use ${v} || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing specific version ${v}"

echo "### Install specific .terraform-version"
cleanup || error_and_die "Cleanup failed?!"

v=0.8.8
echo ${v} > ./.terraform-version
(
  tfenv install || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing .terraform-version ${v}"

echo "### Install latest:<regex> .terraform-version"
cleanup || error_and_die "Cleanup failed?!"

v=$(tfenv list-remote | grep -e '^0.8' | head -n 1)
echo "latest:^0.8" > ./.terraform-version
(
  tfenv install || exit 1
  check_version ${v} || exit 1
) || error_and_proceed "Installing .terraform-version ${v}" 

echo "### Install invalid specific version"
cleanup || error_and_die "Cleanup failed?!"

v=9.9.9
expected_error_message="No versions matching '${v}' found in remote"
[ -z "$(tfenv install ${v} 2>&1 | grep "${expected_error_message}")" ] \
  && error_and_proceed "Installing invalid version ${v}"

echo "### Install invalid latest:<regex> version"
cleanup || error_and_die "Cleanup failed?!"

v="latest:word"
expected_error_message="No versions matching '${v}' found in remote"
[ -z "$(tfenv install ${v} 2>&1 | grep "${expected_error_message}")" ] \
  && error_and_proceed "Installing invalid version ${v}"

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
