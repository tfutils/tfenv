#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: Environment variables';

log 'info' '## TFENV_AUTO_INSTALL=false: prevents auto-install';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.6.1' > .terraform-version;
  # With auto-install disabled, running terraform should fail since 1.6.1 is not installed
  TFENV_AUTO_INSTALL=false terraform version 2>&1 && exit 1;
  exit 0;
) && log 'info' '## TFENV_AUTO_INSTALL=false: passed' \
  || error_and_proceed 'TFENV_AUTO_INSTALL=false did not prevent auto-install';

log 'info' '## TFENV_AUTO_INSTALL=true: triggers auto-install (default)';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.6.1' > .terraform-version;
  terraform version || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## TFENV_AUTO_INSTALL=true: passed' \
  || error_and_proceed 'TFENV_AUTO_INSTALL=true did not auto-install';

log 'info' '## TFENV_SKIP_REMOTE_CHECK: installs known version without remote check';
cleanup || log 'error' 'Cleanup failed?!';
(
  # This should fail because we cannot validate a version exists without remote
  # But we can test that it doesn't contact remote by trying to install an exact
  # version that is already downloaded (needs to be set up first)
  tfenv install 1.6.1 || exit 1;
  # Uninstall and re-install with skip remote check
  rm -rf "${TFENV_CONFIG_DIR}/versions/1.6.1";
  TFENV_SKIP_REMOTE_CHECK=1 tfenv install 1.6.1 || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## TFENV_SKIP_REMOTE_CHECK: passed' \
  || error_and_proceed 'TFENV_SKIP_REMOTE_CHECK did not work as expected';

log 'info' '## TFENV_CONFIG_DIR: uses custom config directory';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare custom_config;
  custom_config="$(mktemp -d 2>/dev/null || mktemp -d -t 'tfenv_cfg_test')";
  mkdir -p "${custom_config}/versions";
  TFENV_CONFIG_DIR="${custom_config}" tfenv install 1.6.1 || exit 1;
  [ -f "${custom_config}/versions/1.6.1/terraform" ] || exit 1;
  rm -rf "${custom_config}";
) && log 'info' '## TFENV_CONFIG_DIR: passed' \
  || error_and_proceed 'TFENV_CONFIG_DIR did not install to custom directory';

log 'info' '## TFENV_BASHLOG=0: lightweight logging works';
cleanup || log 'error' 'Cleanup failed?!';
(
  # With TFENV_BASHLOG=0, commands should still work (just lighter logging)
  declare output;
  output="$(TFENV_BASHLOG=0 tfenv list-remote 2>&1)";
  echo "${output}" | grep -q '1.0.0' || exit 1;
) && log 'info' '## TFENV_BASHLOG=0: passed' \
  || error_and_proceed 'TFENV_BASHLOG=0 broke normal operation';

log 'info' '## TFENV_BASHLOG=1: full logging works';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare output;
  output="$(TFENV_BASHLOG=1 tfenv list-remote 2>&1)";
  echo "${output}" | grep -q '1.0.0' || exit 1;
) && log 'info' '## TFENV_BASHLOG=1: passed' \
  || error_and_proceed 'TFENV_BASHLOG=1 broke normal operation';

log 'info' '## V-prefix stripping: v1.6.1 resolves to 1.6.1';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv install v1.6.1 || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## V-prefix stripping: passed' \
  || error_and_proceed 'V-prefix stripping did not work for v1.6.1';

finish_tests 'env_vars';

exit 0;
