#!/usr/bin/env bash

set -uo pipefail;

if [ -z "${TFENV_ROOT:-""}" ]; then
  # http://stackoverflow.com/questions/1055671/how-can-i-get-the-behavior-of-gnus-readlink-f-on-a-mac
  readlink_f() {
    local target_file="${1}";
    local file_name;

    while [ "${target_file}" != "" ]; do
      cd "${target_file%/*}" || early_death "Failed to 'cd \$(${target_file%/*})' while trying to determine TFENV_ROOT";
      file_name="${target_file##*/}" || early_death "Failed to '\"${target_file##*/}\"' while trying to determine TFENV_ROOT";
      target_file="$(readlink "${file_name}")";
    done;

    echo "$(pwd -P)/${file_name}";
  };
  TFENV_SHIM=$(readlink_f "${0}")
  TFENV_ROOT="${TFENV_SHIM%/*/*}";
  [ -n "${TFENV_ROOT}" ] || early_death "Failed to determine TFENV_ROOT";
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

function load_bashlog () {
  source "${TFENV_ROOT}/lib/bashlog.sh";
};
export -f load_bashlog;

if [ "${TFENV_DEBUG:-0}" -gt 0 ] ; then
  # our shim below cannot be used when debugging is enabled
  load_bashlog;
else
  # Shim that understands to no-op for debug messages, and defers to
  # full bashlog for everything else.
  function log () {
    if [ "$1" != 'debug' ] ; then
      # Loading full bashlog will overwrite the `log` function
      load_bashlog;
      log "$@";
    fi;
  };
  export -f log;
fi;

# Curl wrapper to switch TLS option for each OS
function curlw () {
  local TLS_OPT="--tlsv1.2";

  # Check if curl is 10.12.6 or above
  if [[ -n "$(command -v sw_vers 2>/dev/null)" && ("$(sw_vers)" =~ 10\.12\.([6-9]|[0-9]{2}) || "$(sw_vers)" =~ 10\.1[3-9]) ]]; then
    TLS_OPT="";
  fi;

  if [[ ! -z "${TFENV_NETRC_PATH:-""}" ]]; then
    NETRC_OPT="--netrc-file ${TFENV_NETRC_PATH}";
  else
    NETRC_OPT="";
  fi;

  curl ${TLS_OPT} ${NETRC_OPT} "$@";
};
export -f curlw;

function check_active_version() {
  local v="${1}";
  local maybe_chdir=;
  if [ -n "${2:-}" ]; then
      maybe_chdir="-chdir=${2}";
  fi;

  local active_version="$(${TFENV_ROOT}/bin/terraform ${maybe_chdir} version | grep '^Terraform')";

  if ! grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?( is already installed)?\$" <(echo "${active_version}"); then
    log 'debug' "Expected version ${v} but found ${active_version}";
    return 1;
  fi;

  log 'debug' "Active version ${v} as expected";
  return 0;
};
export -f check_active_version;

function check_installed_version() {
  local v="${1}";
  local bin="${TFENV_CONFIG_DIR}/versions/${v}/terraform";
  [ -n "$(${bin} version | grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?$")" ];
};
export -f check_installed_version;

function check_default_version() {
  local v="${1}";
  local def="$(cat "${TFENV_CONFIG_DIR}/version")";
  [ "${def}" == "${v}" ];
};
export -f check_default_version;

function cleanup() {
  log 'info' 'Performing cleanup';
  local pwd="$(pwd)";
  
  # Safety check to ensure TFENV_CONFIG_DIR is set and not empty
  if [ -z "${TFENV_CONFIG_DIR:-""}" ]; then
    log 'error' 'TFENV_CONFIG_DIR is not set, cannot perform cleanup safely';
    return 1;
  fi;
  
  log 'debug' "Deleting ${TFENV_CONFIG_DIR}/version";
  rm -rf "${TFENV_CONFIG_DIR}/version";
  log 'debug' "Deleting ${TFENV_CONFIG_DIR}/versions";
  rm -rf "${TFENV_CONFIG_DIR}/versions";
  log 'debug' "Deleting ${pwd}/.terraform-version";
  rm -rf ./.terraform-version;
  log 'debug' "Deleting ${pwd}/latest_allowed.tf";
  rm -rf ./latest_allowed.tf;
  log 'debug' "Deleting ${pwd}/min_required.tf";
  rm -rf ./min_required.tf;
  log 'debug' "Deleting ${pwd}/chdir-dir";
  rm -rf ./chdir-dir;
};
export -f cleanup;

function error_and_proceed() {
  errors+=("${1}");
  log 'warn' "Test Failed: ${1}";
};
export -f error_and_proceed;

function check_dependencies() {
  if [[ $(uname) == 'Darwin' ]]; then
    # If installed through brew, it will be available as `ggrep`, so we alias it to `grep`
    if command -v ggrep >/dev/null 2>&1; then
      shopt -s expand_aliases;
      alias grep=ggrep;

      # The alias can't be defined and used in the same parsing unit. But
      # since we know the correct package is installed, we can exit early.
      exit
    fi;

    if ! grep --version 2>&1 | grep -q "GNU grep"; then
      log 'error' 'GNU Grep is a requirement and your Mac does not have it. Consider installing it with `brew install grep` or `nix profile install nixpkgs#gnugrep`';
    fi;
  fi;
};
export -f check_dependencies;

source "$TFENV_ROOT/lib/tfenv-exec.sh";
source "$TFENV_ROOT/lib/tfenv-min-required.sh";
source "$TFENV_ROOT/lib/tfenv-version-file.sh";
source "$TFENV_ROOT/lib/tfenv-version-name.sh";

export TFENV_HELPERS=1;
