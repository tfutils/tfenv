#!/usr/bin/env bash

set -uo pipefail;

# Ensure we can execute standalone
if [ -n "${TFENV_ROOT:-""}" ]; then
  if [ "${TFENV_DEBUG:-0}" -gt 1 ]; then
    [ -n "${TFENV_HELPERS:-""}" ] \
      && log 'debug' "TFENV_ROOT already defined as ${TFENV_ROOT}" \
      || echo "[DEBUG] TFENV_ROOT already defined as ${TFENV_ROOT}" >&2;
  fi;
else
  export TFENV_ROOT="$(cd "$(dirname "${0}")/.." && pwd)";
  if [ "${TFENV_DEBUG:-0}" -gt 1 ]; then
    [ -n "${TFENV_HELPERS:-""}" ] \
      && log 'debug' "TFENV_ROOT declared as ${TFENV_ROOT}" \
      || echo "[DEBUG] TFENV_ROOT declared as ${TFENV_ROOT}" >&2;
  fi;
fi;

if [ -n "${TFENV_HELPERS:-""}" ]; then
  log 'debug' 'TFENV_HELPERS is set, not sourcing helpers again';
else
  [ "${TFENV_DEBUG:-0}" -gt 1 ] && echo "[DEBUG] Sourcing helpers from ${TFENV_ROOT}/lib/helpers.sh" >&2;
  if source "${TFENV_ROOT}/lib/helpers.sh"; then
    log 'debug' 'Helpers sourced successfully';
  else
    echo "[ERROR] Failed to source helpers from ${TFENV_ROOT}/lib/helpers.sh" >&2;
    exit 1;
  fi;
fi;

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

declare test_num=1;
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
