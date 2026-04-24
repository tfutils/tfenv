#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: Commands (help, version, error paths)';

log 'info' '## tfenv help: verify output lists expected commands';
(
  declare help_output;
  help_output="$(tfenv help)";
  echo "${help_output}" | grep -q 'install' || exit 1;
  echo "${help_output}" | grep -q 'use' || exit 1;
  echo "${help_output}" | grep -q 'uninstall' || exit 1;
  echo "${help_output}" | grep -q 'list' || exit 1;
  echo "${help_output}" | grep -q 'list-remote' || exit 1;
  echo "${help_output}" | grep -q 'pin' || exit 1;
  echo "${help_output}" | grep -q 'init' || exit 1;
  echo "${help_output}" | grep -q 'version-name' || exit 1;
) && log 'info' '## tfenv help: passed' \
  || error_and_proceed 'tfenv help output missing expected commands';

log 'info' '## tfenv --version: verify output format';
(
  declare version_output;
  version_output="$(tfenv --version)";
  # Output starts with 'tfenv ' followed by a version string (semver or git describe)
  echo "${version_output}" | grep -qE '^tfenv .+' || exit 1;
) && log 'info' '## tfenv --version: passed' \
  || error_and_proceed 'tfenv --version output does not match expected format';

log 'info' '## tfenv -v: verify same as --version';
(
  declare version_output;
  version_output="$(tfenv -v)";
  echo "${version_output}" | grep -qE '^tfenv .+' || exit 1;
) && log 'info' '## tfenv -v: passed' \
  || error_and_proceed 'tfenv -v output does not match expected format';

log 'info' '## tfenv (no args): verify shows version and help, exits non-zero';
(
  tfenv 2>&1 && exit 1 || true;
  declare output;
  output="$(tfenv 2>&1 || true)";
  echo "${output}" | grep -q 'tfenv' || exit 1;
  echo "${output}" | grep -q 'install' || exit 1;
) && log 'info' '## tfenv (no args): passed' \
  || error_and_proceed 'tfenv with no args did not show expected output';

log 'info' '## tfenv invalid-command: verify error message';
(
  declare output;
  output="$(tfenv nonexistent-command 2>&1 || true)";
  echo "${output}" | grep -qi 'no such command' || exit 1;
) && log 'info' '## tfenv invalid-command: passed' \
  || error_and_proceed 'tfenv with invalid command did not show error';

log 'info' '## tfenv --help: verify exits zero and shows help';
(
  declare output;
  output="$(tfenv --help)";
  echo "${output}" | grep -q 'install' || exit 1;
) && log 'info' '## tfenv --help: passed' \
  || error_and_proceed 'tfenv --help did not show help text';

finish_tests 'commands';

exit 0;
