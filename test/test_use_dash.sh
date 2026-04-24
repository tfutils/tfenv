#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: tfenv use -';
cleanup || log 'error' 'Cleanup failed?!';

log 'info' '## Test: tfenv use - switches to previous version';
(
  tfenv install 0.12.1 || exit 1;
  tfenv install 0.12.2 || exit 1;

  tfenv use 0.12.1 || exit 1;
  check_active_version 0.12.1 || exit 1;

  tfenv use 0.12.2 || exit 1;
  check_active_version 0.12.2 || exit 1;

  # Switch back to previous
  tfenv use - || exit 1;
  check_active_version 0.12.1 || exit 1;

  # Switch back again
  tfenv use - || exit 1;
  check_active_version 0.12.2 || exit 1;
) && log 'info' '## Test passed: tfenv use - switches to previous version' \
  || error_and_proceed 'tfenv use - switches to previous version';

log 'info' '## Test: tfenv use - fails with no previous version';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv install 0.12.1 || exit 1;
  # No previous version set yet — use - should fail
  tfenv use - 2>/dev/null && exit 1 || exit 0;
) && log 'info' '## Test passed: tfenv use - fails with no previous version' \
  || error_and_proceed 'tfenv use - fails with no previous version';

finish_tests 'use-dash';
exit 0;
# vim: set ts=2 sw=2 et:
