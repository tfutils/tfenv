#!/usr/bin/env bash

[ -n "$TFENV_DEBUG" ] && set -x
source $(dirname $0)/helpers.sh

echo "### Install latest version"
cleanup

v=$(tfenv list-remote | head -n 1)
tfenv install latest
tfenv use ${v}
if ! check_version ${v}; then
  echo "Installing latest version ${v}" 1>&2
  exit 1
fi

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
cleanup

v=0.7.13
tfenv install ${v}
tfenv use ${v}
if ! check_version ${v}; then
  echo "Installing specific version ${v} failed" 1>&2
  exit 1
fi

echo "### Install .terraform-version"
cleanup

v=0.8.8
echo ${v} > ./.terraform-version
tfenv install
if ! check_version ${v}; then
  echo "Installing .terraform-version ${v}" 1>&2
  exit 1
fi

echo "### Install invalid version"
cleanup

v=9.9.9
expected_error_message="'${v}' doesn't exist in remote, please confirm version name."
if [ -z "$(tfenv install ${v} 2>&1 | grep "${expected_error_message}")" ]; then
  echo "Installing invalid version ${v}" 1>&2
  exit 1
fi
