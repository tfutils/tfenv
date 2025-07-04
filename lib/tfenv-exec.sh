#!/usr/bin/env bash

set -uo pipefail;

function realpath-relative-to() {
  # A basic implementation of GNU `realpath --relative-to=$1 $2`
  # that can also be used on macOS.

  # http://stackoverflow.com/questions/1055671/how-can-i-get-the-behavior-of-gnus-readlink-f-on-a-mac
  readlink_f() {
    local target_file="${1}";
    local file_name;

    while [ "${target_file}" != "" ]; do
      cd "$(dirname "$target_file")" || early_death "Failed to 'cd \$(${target_file%/*})'";
      file_name="${target_file##*/}" || early_death "Failed to '\"${target_file##*/}\"'";
      target_file="$(readlink "${file_name}")";
    done;

    echo "$(pwd -P)/${file_name}";
  };

  local relative_to="$(readlink_f "${1}")";
  local path="$(readlink_f "${2}")";

  echo "${path#"${relative_to}/"}";
  return 0;
}
export -f realpath-relative-to;

function tfenv-exec() {
  for _arg in ${@:1}; do
    if [[ "${_arg}" == -chdir=* ]]; then
      chdir="${_arg#-chdir=}";
      log 'debug' "Found -chdir arg: ${chdir}";
      export TFENV_DIR="${PWD}/$(realpath-relative-to "${PWD}" "${chdir}")";
      log 'debug' "Setting TFENV_DIR to: ${TFENV_DIR}";
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
