#!/usr/bin/env bash

[ -n "$TFENV_DEBUG" ] && set -x

check_version() {
  v=${1}
  [ -n "$(terraform --version | grep "Terraform v${v}")" ]
}

echo "### Install latest version"
rm -rf ./versions
rm -rf ./.terraform-version

v=$(tfenv list-remote | head -n 1)
tfenv install latest
tfenv use ${v}
if ! check_version ${v}; then
  echo "Installing latest version ${v}" 1>&2
  exit 1
fi

echo "### Install specific version"
rm -rf ./versions
rm -rf ./.terraform-version

v=0.6.16
tfenv install ${v}
tfenv use ${v}
if ! check_version ${v}; then
  echo "Installing specific version ${v} failed" 1>&2
  exit 1
fi

echo "### Install .terraform-version"
rm -rf ./versions
rm -rf ./.terraform-version

v=0.6.15
echo ${v} > ./.terraform-version
tfenv install
if ! check_version ${v}; then
  echo "Installing .terraform-version ${v}" 1>&2
  exit 1
fi
