#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Testing symlink functionality';

TFENV_BIN_DIR="$(mktemp -d)";
log 'info' "## Using temporary directory ${TFENV_BIN_DIR}";
trap 'rm -rf "${TFENV_BIN_DIR}"' EXIT;
log 'info' "## Symlinking ${TFENV_ROOT}/bin/* into ${TFENV_BIN_DIR}";
ln -s "${TFENV_ROOT}"/bin/* "${TFENV_BIN_DIR}";

cleanup || log 'error' 'Cleanup failed?!';

log 'info' '## Installing 1.6.1';
${TFENV_BIN_DIR}/tfenv install 1.6.1 || error_and_proceed 'Install failed';

log 'info' '## Using 1.6.1';
${TFENV_BIN_DIR}/tfenv use 1.6.1 || error_and_proceed 'Use failed';

log 'info' '## Check-Version for 1.6.1';
check_active_version 1.6.1 || error_and_proceed 'Version check failed';

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
