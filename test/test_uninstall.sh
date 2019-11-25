#!/usr/bin/env bash

set -uo pipefail;

####################################
# Ensure we can execute standalone #
####################################

function early_death() {
  echo "[FATAL] ${0}: ${1}" >&2;
  exit 1;
};

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

if [ -n "${TFENV_HELPERS:-""}" ]; then
  log 'debug' 'TFENV_HELPERS is set, not sourcing helpers again';
else
  [ "${TFENV_DEBUG:-0}" -gt 0 ] && echo "[DEBUG] Sourcing helpers from ${TFENV_ROOT}/lib/helpers.sh";
  if source "${TFENV_ROOT}/lib/helpers.sh"; then
    log 'debug' 'Helpers sourced successfully';
  else
    early_death "Failed to source helpers from ${TFENV_ROOT}/lib/helpers.sh";
  fi;
fi;

#####################
# Begin Script Body #
#####################

declare -a errors=();

function test_uninstall() {
  local k="${1}";
  local v="${2}";
  tfenv install "${v}" || return 1;
  tfenv uninstall "${v}" || return 1;
  log 'info' 'Confirming uninstall success; an error indicates success:';
  check_version "${v}" && return 1 || return 0;
}

log 'info' '### Test Suite: Uninstall Local Versions'
cleanup || log 'error' 'Cleanup failed?!';

declare -A tests;
tests['0.9.1']='0.9.1';
tests['0.11.15-oci']='0.11.15-oci';
tests['latest']="$(tfenv list-remote | head -n1)";
tests['latest:^0.8']="$(tfenv list-remote | grep -e "^0.8" | head -n1)";

declare -i test_num=1;
for k in "${!tests[@]}"; do
  log 'info' "Test ${test_num}/${#tests[@]}: Testing uninstall of version ${tests[${k}]} via keyword ${k}";
  test_uninstall "${k}" "${tests[${k}]}" \
    && log info "Test uninstall of version ${tests[${k}]} succeeded" \
    || error_and_proceed "Test uninstall of version ${tests[${k}]} failed";
  test_num+=1;
done;

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' "===== The following list tests failed =====";
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'List test failure(s)';
else
  log 'info' 'All list tests passed.';
fi;
exit 0;
