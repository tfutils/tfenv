#!/usr/bin/env bash
set -x

TFENV_ROOT=$(cd $(dirname $0)/.. && pwd)
export PATH="${TFENV_ROOT}/bin:${PATH}"
export TFENV_DEBUG=1

rm -rf ${TFENV_ROOT}/versions

error_versions=()
for v in 0.{1..7}.0; do
  echo "======================================================================"
  tfenv install ${v}
  tfenv use ${v}
  if [ "$(terraform --version | grep -E "^Terraform v[0-9\.]+$")" != "Terraform v${v}" ]; then
    error_versions+=( ${v} )
  fi
done

if [ ${#error_versions[@]} -ne 0 ];then
  echo "Following versions couldn't be installed properly" 1>&2
  echo "${error_versions[@]}"
  exit 1
fi
