#!/usr/bin/env bash

available() {
  v=${1}
  [ "$(terraform --version | grep -E "^Terraform v[0-9\.]+$")" == "Terraform v${v}" ]
}

### Install specific version
rm -rf ${TFENV_ROOT}/versions

v=0.6.16
tfenv install ${v}
tfenv use ${v}
available ${v}
if ! available ${v}; then
  echo "Installing specific version ${v} failed" 1>&2
  exit 1
fi
