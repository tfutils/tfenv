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

# Detect arch from each installed binary — can't assume host arch
# because older versions may only have amd64 builds (e.g. macOS arm64 via Rosetta)
detect_arch() {
  local binary="${1}";
  local file_output;
  file_output="$(file -b "${binary}" 2>/dev/null)" || { echo 'unknown'; return; };
  case "${file_output}" in
    *x86-64*|*x86_64*)  echo 'amd64';;
    *aarch64*|*arm64*)   echo 'arm64';;
    *386*|*i386*|*80386*) echo '386';;
    *ARM,*)              echo 'arm';;
    *)                   echo 'unknown';;
  esac;
};

expected="$(cat << EOS
* 0.14.6 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.14.6/terraform")) (set by $(tfenv version-file))
  0.9.11 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.9.11/terraform"))
  0.9.2 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.9.2/terraform"))
  0.9.1 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.9.1/terraform"))
  0.7.13 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.7.13/terraform"))
  0.7.2 ($(detect_arch "${TFENV_CONFIG_DIR}/versions/0.7.2/terraform"))
EOS
)";

[ "${expected}" == "${result}" ] \
  && log 'info' 'List matches expectation.' \
  || error_and_proceed "List mismatch.\nExpected:\n${expected}\nGot:\n${result}";

finish_tests 'list';

exit 0;
