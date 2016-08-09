#!/usr/bin/env bash

check_version() {
  v=${1}
  [ "$(terraform --version | grep -E "^Terraform v[0-9\.]+$")" == "Terraform v${v}" ]
}

echo "### Install latest version"
rm -rf ${TFENV_ROOT}/versions
rm -rf ${TFENV_ROOT}/.terraform-version

if ! tfenv install; then
  echo "Installing latest version ${v}" 1>&2
  exit 1
fi

echo "### Install specific version"
rm -rf ${TFENV_ROOT}/versions
rm -rf ${TFENV_ROOT}/.terraform-version

v=0.6.16
tfenv install ${v}
tfenv use ${v}
if ! check_version ${v}; then
  echo "Installing specific version ${v} failed" 1>&2
  exit 1
fi

echo "### Install .terraform-version"
rm -rf ${TFENV_ROOT}/versions
rm -rf ${TFENV_ROOT}/.terraform-version

v=0.6.15
touch ${TFENV_ROOT}/.terraform-version && echo ${v} > ${TFENV_ROOT}/.terraform-version
tfenv install
if ! check_version ${v}; then
  echo "Installing .terraform-version ${v}" 1>&2
  exit 1
fi
