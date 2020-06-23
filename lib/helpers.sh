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

if [ "${TFENV_DEBUG:-0}" -gt 0 ]; then
  [ "${DEBUG:-0}" -gt "${TFENV_DEBUG:-0}" ] || export DEBUG="${TFENV_DEBUG:-0}";
  if [[ "${TFENV_DEBUG}" -gt 2 ]]; then
    export PS4='+ [${BASH_SOURCE##*/}:${LINENO}] ';
    set -x;
  fi;
fi;

source "${TFENV_ROOT}/lib/bashlog.sh";

resolve_version () {
  declare version_requested version regex min_required version_file;

  declare arg="${1:-""}";

  if [ -z "${arg}" ]; then
    version_file="$(tfenv-version-file)";
    log 'debug' "Version File: ${version_file}";

    if [ "${version_file}" != "${TFENV_ROOT}/version" ]; then
      log 'debug' "Version File (${version_file}) is not the default \${TFENV_ROOT}/version (${TFENV_ROOT}/version)";
      version_requested="$(cat "${version_file}")" \
        || log 'error' "Failed to open ${version_file}";

    elif [ -f "${version_file}" ]; then
      log 'debug' "Version File is the default \${TFENV_ROOT}/version (${TFENV_ROOT}/version)";
      version_requested="$(cat "${version_file}")" \
        || log 'error' "Failed to open ${version_file}";

      # Absolute fallback
      if [ -z "${version_requested}" ]; then
        log 'debug' 'Version file had no content. Falling back to "latest"';
        version_requested='latest';
      fi;

    else
      log 'debug' "Version File is the default \${TFENV_ROOT}/version (${TFENV_ROOT}/version) but it doesn't exist";
      log 'info' 'No version requested on the command line or in the version file search path. Installing "latest"';
      version_requested='latest';
    fi;
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
  local bin="${TFENV_ROOT}/versions/${v}/terraform";
  [ -n "$(${bin} version | grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?$")" ];
};
export -f check_installed_version;

check_default_version() {
  local v="${1}";
  local def="$(cat "${TFENV_ROOT}/version")";
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
};
export -f cleanup;

function error_and_proceed() {
  errors+=("${1}");
  log 'warn' "Test Failed: ${1}";
};
export -f error_and_proceed;

export TFENV_HELPERS=1;
