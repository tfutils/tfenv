#!/usr/bin/env bash

set -uo pipefail;

function tfenv-exec() {
  for _arg in ${@:1}; do
    if [[ "${_arg}" == -chdir=* ]]; then
      log 'debug' "Found -chdir arg. Setting TFENV_DIR to: ${_arg#-chdir=}";
      export TFENV_DIR="${PWD}/${_arg#-chdir=}";
    fi;
  done;

  log 'debug' 'Getting version from tfenv-version-name';
  TFENV_VERSION="$(tfenv-version-name)" \
    && log 'debug' "TFENV_VERSION is ${TFENV_VERSION}" \
    || {
      # Errors will be logged from tfenv-version name,
      # we don't need to trouble STDERR with repeat information here
      log 'debug' 'Failed to get version from tfenv-version-name';
      return 1;
    };
  export TFENV_VERSION;

  if [ ! -d "${TFENV_CONFIG_DIR}/versions/${TFENV_VERSION}" ]; then
  if [ "${TFENV_AUTO_INSTALL:-true}" == "true" ]; then
    if [ -z "${TFENV_TERRAFORM_VERSION:-""}" ]; then
      TFENV_VERSION_SOURCE="$(tfenv-version-file)";
    else
      TFENV_VERSION_SOURCE='TFENV_TERRAFORM_VERSION';
    fi;
      log 'info' "version '${TFENV_VERSION}' is not installed (set by ${TFENV_VERSION_SOURCE}). Installing now as TFENV_AUTO_INSTALL==true";
      tfenv-install;
    else
      log 'error' "version '${TFENV_VERSION}' was requested, but not installed and TFENV_AUTO_INSTALL is not 'true'";
    fi;
  fi;

  TF_BIN_PATH="${TFENV_CONFIG_DIR}/versions/${TFENV_VERSION}/terraform";
  export PATH="${TF_BIN_PATH}:${PATH}";
  log 'debug' "TF_BIN_PATH added to PATH: ${TF_BIN_PATH}";
  log 'debug' "Executing: ${TF_BIN_PATH} $@";

  exec "${TF_BIN_PATH}" "$@" \
  || log 'error' "Failed to execute: ${TF_BIN_PATH} $*";

  return 0;
};
export -f tfenv-exec;
