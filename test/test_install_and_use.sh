#!/usr/bin/env bash

install_and_use() {
  rm -rf ${TFENV_ROOT}/versions
  v=${1}
  tfenv install ${v}
  tfenv use ${v}
  [ "$(terraform --version | grep -E "^Terraform v[0-9\.]+$")" == "Terraform v${v}" ]
}

### Install specific version
v=0.6.16
if ! install_and_use ${v}; then
  echo "Installing specific version ${v} failed" 1>&2
  exit 1
fi

### Install latest version
rm -rf ${TFENV_ROOT}/versions
if ! tfenv install; then
  echo "Installing latest version ${v}" 1>&2
  exit 1
fi
