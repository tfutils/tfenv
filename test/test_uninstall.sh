#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

function test_uninstall() {
  local k="${1}";
  local v="${2}";
  tfenv install "${v}" || return 1;
  tfenv uninstall "${v}" || return 1;
  log 'info' 'Confirming uninstall success; an error indicates success:';
  check_active_version "${v}" && return 1 || return 0;
};

log 'info' '### Test Suite: Uninstall Local Versions';
cleanup || log 'error' 'Cleanup failed?!';

tests__keywords=(
  '0.9.1'
  '0.11.15-oci'
  'latest'
  'latest:^0.8'
  'v0.14.6'
);

tests__versions=(
  '0.9.1'
  '0.11.15-oci'
  "$(tfenv list-remote | head -n1)"
  "$(tfenv list-remote | grep -e "^0.8" | head -n1)"
  '0.14.6'
);

tests_count=${#tests__keywords[@]};

for ((test_num=0; test_num<${tests_count}; ++test_num )) ; do
  keyword=${tests__keywords[${test_num}]};
  version=${tests__versions[${test_num}]};
  log 'info' "Test $(( ${test_num} + 1 ))/${tests_count}: Testing uninstall of version ${version} via keyword ${keyword}";
  test_uninstall "${keyword}" "${version}" \
    && log info "Test uninstall of version ${version} (via ${keyword}) succeeded" \
    || error_and_proceed "Test uninstall of version ${version} (via ${keyword}) failed";
done;

echo "### Uninstall removes versions directory"
cleanup || error_and_die "Cleanup failed?!"
(
  tfenv install 0.12.1 || exit 1
  tfenv install 0.12.2 || exit 1
  [ -d "${TFENV_CONFIG_DIR}/versions" ] || exit 1
  tfenv uninstall 0.12.1 || exit 1
  [ -d "${TFENV_CONFIG_DIR}/versions" ] || exit 1
  tfenv uninstall 0.12.2 || exit 1
  [ -d "${TFENV_CONFIG_DIR}/versions" ] && exit 1 || exit 0
) || error_and_proceed "Removing last version deletes versions directory"

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' "===== The following list tests failed =====";
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'List test failure(s)';
else
  log 'info' 'All list tests passed.';
fi;
exit 0;
