#!/usr/bin/env bash

set -uo pipefail;

if [ "${TFENV_DEBUG:-0}" -gt 0 ]; then
  [ "${DEBUG:-0}" -gt "${TFENV_DEBUG:-0}" ] || export DEBUG="${TFENV_DEBUG:-0}";
  if [[ "${TFENV_DEBUG}" -gt 2 ]]; then
    export PS4='+ [${BASH_SOURCE##*/}:${LINENO}] ';
    set -x;
  fi;
fi;

[ -n "${TFENV_ROOT:-""}" ] || export TFENV_ROOT="$(cd "$(dirname "${0}")/.." && pwd)"
source "${TFENV_ROOT}/lib/bashlog.sh";

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

check_version() {
  v="${1}";
  [ -n "$(terraform --version | grep -E "^Terraform v${v}((-dev)|( \([a-f0-9]+\)))?$")" ];
}
export -f check_version;

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
