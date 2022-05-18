#!/usr/bin/env bash

set -uo pipefail;

if [ -z "${TFENV_ROOT:-""}" ]; then
  # http://stackoverflow.com/questions/1055671/how-can-i-get-the-behavior-of-gnus-readlink-f-on-a-mac
  readlink_f() {
    local target_file="${1}";
    local file_name;

    while [ "${target_file}" != "" ]; do
      cd "$(dirname ${target_file})" || early_death "Failed to 'cd \$(dirname ${target_file})' while trying to determine TFENV_ROOT";
      file_name="$(basename "${target_file}")" || early_death "Failed to 'basename \"${target_file}\"' while trying to determine TFENV_ROOT";
      target_file="$(readlink "${file_name}")";
    done;

    echo "$(pwd -P)/${file_name}";
  };

  TFENV_ROOT="$(cd "$(dirname "$(readlink_f "${0}")")/.." && pwd)";
  [ -n ${TFENV_ROOT} ] || early_death "Failed to 'cd \"\$(dirname \"\$(readlink_f \"${0}\")\")/..\" && pwd' while trying to determine TFENV_ROOT";
else
  TFENV_ROOT="${TFENV_ROOT%/}";
fi;
export TFENV_ROOT;

if [ -z "${TFENV_CONFIG_DIR:-""}" ]; then
  TFENV_CONFIG_DIR="$TFENV_ROOT";
else
  TFENV_CONFIG_DIR="${TFENV_CONFIG_DIR%/}";
fi
export TFENV_CONFIG_DIR;

if [ "${TFENV_DEBUG:-0}" -gt 0 ]; then
  # Only reset DEBUG if TFENV_DEBUG is set, and DEBUG is unset or already a number
  if [[ "${DEBUG:-0}" =~ ^[0-9]+$ ]] && [ "${DEBUG:-0}" -gt "${TFENV_DEBUG:-0}" ]; then
    export DEBUG="${TFENV_DEBUG:-0}";
  fi;
  if [[ "${TFENV_DEBUG}" -gt 2 ]]; then
    export PS4='+ [${BASH_SOURCE##*/}:${LINENO}] ';
    set -x;
  fi;
fi;

source "${TFENV_ROOT}/lib/bashlog.sh";

resolve_version () {
  declare version_requested version regex min_required version_file;

  declare arg="${1:-""}";

  if [ -z "${arg}" -a -z "${TFENV_TERRAFORM_VERSION:-""}" ]; then
    version_file="$(tfenv-version-file)";
    log 'debug' "Version File: ${version_file}";

    if [ "${version_file}" != "${TFENV_CONFIG_DIR}/version" ]; then
      log 'debug' "Version File (${version_file}) is not the default \${TFENV_CONFIG_DIR}/version (${TFENV_CONFIG_DIR}/version)";
      version_requested="$(cat "${version_file}")" \
        || log 'error' "Failed to open ${version_file}";
    fi

    if [ -z "${version_requested:-""}" ]; then
      log 'debug' 'Tryng to set version from "required_version" under "terraform" section'
      versions="$( echo $(cat {*.tf,*.tf.json} 2>/dev/null | grep -h required_version) | grep  -o '\([0-9]\+\.\?\)\{2,3\}\(-[a-z]\+[0-9]\+\)\?')";
      if [[ "${versions}" =~ ([~=!<>]{0,2}[[:blank:]]*[0-9]+[0-9.]+)[^0-9]*(-[a-z]+[0-9]+)? ]]; then
        found_min_required="${BASH_REMATCH[1]}${BASH_REMATCH[2]}"
        if [[ "${found_min_required}" =~ ^!=.+ ]]; then
          log 'debug' "required_version is a negation - we cannot guess the desired one, skipping.";
        else
          found_min_required="$(echo "$found_min_required")";

          # Probably not an advisable way to choose a terraform version,
          # but this is the way this functionality works in terraform:
          # add .0 to versions without a minor and/or patch version (e.g. 12.0)
          while ! [[ "${found_min_required}" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; do
            found_min_required="${found_min_required}.0";
          done;
          version_requested="${found_min_required}";
        fi;
      fi;
    fi;

    if [ -z "${version_requested}" -a -f "${version_file}" ]; then
      log 'debug' "Version File is the default \${TFENV_CONFIG_DIR}/version (${TFENV_CONFIG_DIR}/version)";
      version_requested="$(cat "${version_file}")" \
        || log 'error' "Failed to open ${version_file}";

      if [ -z "${version_requested}" ]; then
        log 'debug' 'Version file had no content. Falling back to "latest"';
        version_requested='latest';
      fi;

    # Absolute fallback
    elif [ -z "${version_requested}" ]; then
      log 'debug' "Version File is the default \${TFENV_CONFIG_DIR}/version (${TFENV_CONFIG_DIR}/version) but it doesn't exist";
      log 'info' 'No version requested on the command line or in the version file search path. Installing "latest"';
      version_requested='latest';
    fi;
  elif [ -n "${TFENV_TERRAFORM_VERSION:-""}" ]; then
    version_requested="${TFENV_TERRAFORM_VERSION}";
    log 'debug' "TFENV_TERRAFORM_VERSION is set: ${TFENV_TERRAFORM_VERSION}";
  else
    version_requested="${arg}";
  fi;

  log 'debug' "Version Requested: ${version_requested}";

  if [[ "${version_requested}" =~ ^min-required$ ]]; then
    log 'info' 'Detecting minimum required version...';
    min_required="$(tfenv-min-required)" \
      || log 'error' 'tfenv-min-required failed';

    log 'info' "Minimum required version detected: ${min_required}";
    version_requested="${min_required}";
  fi;

  if [[ "${version_requested}" =~ ^latest\:.*$ ]]; then
    version="${version_requested%%\:*}";
    regex="${version_requested##*\:}";
    log 'debug' "Version uses latest keyword with regex: ${regex}";
  elif [[ "${version_requested}" =~ ^latest$ ]]; then
    version="${version_requested}";
    regex="^[0-9]\+\.[0-9]\+\.[0-9]\+$";
    log 'debug' "Version uses latest keyword alone. Forcing regex to match stable versions only: ${regex}";
  else
    version="${version_requested}";
    regex="^${version_requested}$";
    log 'debug' "Version is explicit: ${version}. Regex enforces the version: ${regex}";
  fi;
}

# Curl wrapper to switch TLS option for each OS
function curlw () {
  local TLS_OPT="--tlsv1.2";

  # Check if curl is 10.12.6 or above
  if [[ -n "$(command -v sw_vers 2>/dev/null)" && ("$(sw_vers)" =~ 10\.12\.([6-9]|[0-9]{2}) || "$(sw_vers)" =~ 10\.1[3-9]) ]]; then
    TLS_OPT="";
  fi;

  curl ${TLS_OPT} "$@";
}
export -f curlw;

check_active_version() {
  local v="${1}";
  [ -n "$(${TFENV_ROOT}/bin/terraform version | grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?$")" ];
}
export -f check_active_version;

check_installed_version() {
  local v="${1}";
  local bin="${TFENV_CONFIG_DIR}/versions/${v}/terraform";
  [ -n "$(${bin} version | grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?$")" ];
};
export -f check_installed_version;

check_default_version() {
  local v="${1}";
  local def="$(cat "${TFENV_CONFIG_DIR}/version")";
  [ "${def}" == "${v}" ];
};
export -f check_default_version;

cleanup() {
  log 'info' 'Performing cleanup';
  local pwd="$(pwd)";
  log 'debug' "Deleting ${pwd}/versions";
  rm -rf ./versions;
  log 'debug' "Deleting ${pwd}/.terraform-version";
  rm -rf ./.terraform-version;
  log 'debug' "Deleting ${pwd}/min_required.tf";
  rm -rf ./min_required.tf;
  log 'debug' "Deleting ${pwd}/required_version.tf";
  rm -rf ./required_version.tf;
};
export -f cleanup;

function error_and_proceed() {
  errors+=("${1}");
  log 'warn' "Test Failed: ${1}";
};
export -f error_and_proceed;

export TFENV_HELPERS=1;
