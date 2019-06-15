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
source "$(dirname "${0}")/helpers.sh" \
  || error_and_die "Failed to load test helpers: $(dirname "${0}")/helpers.sh"

TFENV_BIN_DIR="/tmp/tfenv-test"
rm -rf "${TFENV_BIN_DIR}" && mkdir "${TFENV_BIN_DIR}"
ln -s "${PWD}"/bin/* "${TFENV_BIN_DIR}"
export PATH="${TFENV_BIN_DIR}:${PATH}"

echo "### Display Usage message"
cleanup || error_and_die "Cleanup failed?!"

expected_error_message="usage: tfenv local <version>"
[ -z "$(tfenv local 2>&1 | grep "${expected_error_message}")" ] \
  && error_and_proceed "local invalid version"

echo "### Delete .terraform-version"
cleanup || error_and_die "Cleanup failed?!"

echo "0.12.2" > ./.terraform-version
tfenv local --unset
if [[ -a "./.terraform-version" ]]; then
  error_and_proceed ".terraform-version is not deleted"
fi

echo "### Create latest .terraform-version"
cleanup || error_and_die "Cleanup failed?!"

v="latest"
(
  tfenv install "${v}" || exit 1
  tfenv local "${v}" || exit 1
) || error_and_proceed ".terraform-version ${v} is not created"

v="latest:^0.8"
(
  tfenv install 0.8.8 || exit 1
  tfenv local "${v}" || exit 1
) || error_and_proceed ".terraform-version ${v} is not created"

echo "### Create 0.12.2 .terraform-version"
cleanup || error_and_die "Cleanup failed?!"

v="0.12.2"
(
  tfenv install "${v}" || exit 1
  tfenv local "${v}" || exit 1
) || error_and_proceed ".terraform-version ${v} is not created"

echo "### Can not create invalid specific version"
cleanup || error_and_die "Cleanup failed?!"

v="9.9.9"
tfenv install 0.12.2
expected_error_message="No installed versions of terraform matched '${v}'"
[ -z "$(tfenv local "${v}" 2>&1 | grep "${expected_error_message}")" ] \
  && error_and_proceed "created invalid version ${v}"

if [ "${#errors[@]}" -gt 0 ]; then
  echo -e "\033[0;31m===== The following list tests failed =====\033[0;39m" >&2
  for error in "${errors[@]}"; do
    echo -e "\t${error}"
  done
  exit 1
else
  echo -e "\033[0;32mAll list tests passed.\033[0;39m"
fi;
exit 0
