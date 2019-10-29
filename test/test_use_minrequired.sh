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

log 'info' '### Install not min-required version';
cleanup || log 'error' 'Cleanup failed?!';

v='0.8.8';
minv='0.8.0';
(
  tfenv install "${v}" || true;
  tfenv use "${v}" || exit 1;
  check_version "${v}" || exit 1;
) || error_and_proceed "Installing specific version ${v}";

echo "terraform {

  required_version = \">=${minv}\"
}" >> min_required.tf;

tfenv install min-required;
tfenv use min-required;

check_version "${minv}" || error_and_proceed 'Min required version does not match';

cleanup || log 'error' 'Cleanup failed?!';

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following use_minrequired tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'use_minrequired test failure(s)';
else
  log 'info' 'All use_minrequired tests passed.';
fi;

exit 0;
