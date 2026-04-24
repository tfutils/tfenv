#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### List local versions';
cleanup || log 'error' "Cleanup failed?!";

for v in 0.7.2 0.7.13 0.9.1 0.9.2 v0.9.11 0.14.6; do
  log 'info' "## Installing version ${v} to construct list";
  tfenv install "${v}" \
    && log 'debug' "Install of version ${v} succeeded" \
    || error_and_proceed "Install of version ${v} failed";
done;

log 'info' '## Ensuring tfenv list success with no default set';
tfenv list \
  && log 'debug' "List succeeded with no default set" \
  || error_and_proceed "List failed with no default set";

tfenv use 0.14.6;

log 'info' '## Comparing "tfenv list" with default set';
result="$(tfenv list)";

# Determine expected arch from host platform
case "$(uname -m)" in
  x86_64)       expected_arch='amd64';;
  aarch64|arm64) expected_arch='arm64';;
  i386|i686)    expected_arch='386';;
  *)            expected_arch='unknown';;
esac;

expected="$(cat << EOS
* 0.14.6 (${expected_arch}) (set by $(tfenv version-file))
  0.9.11 (${expected_arch})
  0.9.2 (${expected_arch})
  0.9.1 (${expected_arch})
  0.7.13 (${expected_arch})
  0.7.2 (${expected_arch})
EOS
)";

[ "${expected}" == "${result}" ] \
  && log 'info' 'List matches expectation.' \
  || error_and_proceed "List mismatch.\nExpected:\n${expected}\nGot:\n${result}";

finish_tests 'list';

exit 0;
