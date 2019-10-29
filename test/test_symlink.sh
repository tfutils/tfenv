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

log 'info' '### Testing symlink functionality';

TFENV_BIN_DIR='/tmp/tfenv-test';
log 'info' "## Creating/clearing ${TFENV_BIN_DIR}"
rm -rf "${TFENV_BIN_DIR}" && mkdir "${TFENV_BIN_DIR}";
log 'info' "## Symlinking ${PWD}/bin/* into ${TFENV_BIN_DIR}";
ln -s "${PWD}"/bin/* "${TFENV_BIN_DIR}";
log 'info' "## Adding ${TFENV_BIN_DIR} to \$PATH";
export PATH="${TFENV_BIN_DIR}:${PATH}";

cleanup || log 'error' 'Cleanup failed?!';

log 'info' '## Installing 0.8.2';
tfenv install 0.8.2 || error_and_proceed 'Install failed';

log 'info' '## Using 0.8.2';
tfenv use 0.8.2 || error_and_proceed 'Use failed';

log 'info' '## Check-Version for 0.8.2';
check_version 0.8.2 || error_and_proceed 'Version check failed';

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following symlink tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'Symlink test failure(s)';
  exit 1;
else
  log 'info' 'All symlink tests passed.';
fi;

exit 0;
