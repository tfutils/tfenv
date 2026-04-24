#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: list-remote';

log 'info' '## list-remote: returns version list';
(
  declare output;
  output="$(tfenv list-remote)";
  [ -n "${output}" ] || exit 1;
  # Should contain known stable versions
  echo "${output}" | grep -q '1.0.0' || exit 1;
  echo "${output}" | grep -q '0.12.0' || exit 1;
) && log 'info' '## list-remote: basic output passed' \
  || error_and_proceed 'list-remote did not return expected version list';

log 'info' '## list-remote: output contains pre-release versions';
(
  declare output;
  output="$(tfenv list-remote)";
  # Should include rc/beta/alpha versions
  echo "${output}" | grep -qE '(rc|beta|alpha)' || exit 1;
) && log 'info' '## list-remote: pre-release versions present' \
  || error_and_proceed 'list-remote output missing pre-release versions';

log 'info' '## list-remote: output is one version per line';
(
  declare output;
  output="$(tfenv list-remote)";
  # Every line should match a version pattern
  declare invalid;
  invalid="$(echo "${output}" | grep -cvE '^[0-9]+\.[0-9]+\.[0-9]+(-(rc|beta|alpha|oci)-?[0-9]*)?$' || true)";
  [ "${invalid}" -eq 0 ] || exit 1;
) && log 'info' '## list-remote: format validation passed' \
  || error_and_proceed 'list-remote output contains non-version lines';

log 'info' '## list-remote: rejects arguments';
(
  tfenv list-remote extra-arg 2>&1 && exit 1;
  exit 0;
) && log 'info' '## list-remote: argument rejection passed' \
  || error_and_proceed 'list-remote did not reject extra arguments';

log 'info' '## list-remote: TFENV_REVERSE_REMOTE reverses output';
(
  declare normal_first normal_last reversed_first reversed_last;
  normal_first="$(tfenv list-remote | head -n 1)";
  normal_last="$(tfenv list-remote | tail -n 1)";
  reversed_first="$(TFENV_REVERSE_REMOTE=1 tfenv list-remote | head -n 1)";
  reversed_last="$(TFENV_REVERSE_REMOTE=1 tfenv list-remote | tail -n 1)";
  [ "${normal_first}" == "${reversed_last}" ] || exit 1;
  [ "${normal_last}" == "${reversed_first}" ] || exit 1;
) && log 'info' '## list-remote: TFENV_REVERSE_REMOTE passed' \
  || error_and_proceed 'TFENV_REVERSE_REMOTE did not reverse output';

finish_tests 'list_remote';

exit 0;
