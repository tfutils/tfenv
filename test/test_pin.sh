#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: pin';

log 'info' '## pin: pins current version to .terraform-version';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv install 1.6.1 || exit 1;
  tfenv use 1.6.1 || exit 1;
  tfenv pin || exit 1;
  [ -f .terraform-version ] || exit 1;
  declare pinned;
  pinned="$(cat .terraform-version)";
  [ "${pinned}" == '1.6.1' ] || exit 1;
) && log 'info' '## pin: basic pin passed' \
  || error_and_proceed 'pin did not write correct version to .terraform-version';

log 'info' '## pin: rejects arguments';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv pin extra-arg 2>&1 && exit 1;
  exit 0;
) && log 'info' '## pin: argument rejection passed' \
  || error_and_proceed 'pin did not reject extra arguments';

log 'info' '## pin: fails with no versions installed';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv pin 2>&1 && exit 1;
  exit 0;
) && log 'info' '## pin: no-versions error passed' \
  || error_and_proceed 'pin did not fail with no versions installed';

log 'info' '## pin: overwrites existing .terraform-version';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv install 1.6.1 || exit 1;
  tfenv use 1.6.1 || exit 1;
  # Write a bogus version, then remove it so pin reads from use'd default
  echo '0.0.0' > .terraform-version;
  rm -f .terraform-version;
  tfenv pin || exit 1;
  declare pinned;
  pinned="$(cat .terraform-version)";
  [ "${pinned}" == '1.6.1' ] || exit 1;
) && log 'info' '## pin: overwrite passed' \
  || error_and_proceed 'pin did not write correct version to .terraform-version';

finish_tests 'pin';

exit 0;
