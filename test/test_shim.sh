#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: terraform shim';

log 'info' '## terraform shim: passes through to terraform (without args)';
cleanup || log 'error' 'Cleanup failed?!';
(
  tfenv install 1.6.1 || exit 1;
  tfenv use 1.6.1 || exit 1;
  declare output;
  output="$(terraform 2>&1 || true)";
  # https://github.com/tfutils/tfenv/issues/532
  echo "${output}" | grep -q 'unbound variable' && exit 1;
  echo "${output}" | grep -q 'Usage: terraform' || exit 1;
) && log 'info' '## terraform (no args): passed' \
  || error_and_proceed 'terraform with no args crashed or did not show usage';

log 'info' '## terraform shim: passes through to terraform (with args)';
(
  tfenv install 1.6.1 || exit 1;
  tfenv use 1.6.1 || exit 1;
  declare output;
  output="$(terraform version -json)" || exit 1;
  echo "${output}" | grep -q 'terraform_version' || exit 1;
  echo "${output}" | grep -q '1.6.1' || exit 1;
) && log 'info' '## terraform shim: arguments are passed through: passed' \
  || error_and_proceed 'terraform arguments were not passed through';

finish_tests 'shim';

exit 0;
