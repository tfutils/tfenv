#!/usr/bin/env bash
set -uo pipefail;

####################################
# Ensure we can execute standalone #
####################################

function early_death() {
  echo "[FATAL] ${0}: ${1}" >&2;
  exit 1;
};

if [ -z "${TFENV_ROOT:-""}" ]; then
  # http://stackoverflow.com/questions/1055671/how-can-i-get-the-behavior-of-gnus-readlink-f-on-a-mac
  readlink_f() {
    local target_file="${1}";
    local file_name;

    while [ "${target_file}" != "" ]; do
      cd "$(dirname ${target_file})" || early_death "Failed to 'cd \$(dirname ${target_file})' while trying to determine TFENV_ROOT";
      file_name="$(basename "${target_file}")" || early_death "Failed to 'basename \"${target_file}\"' while trying to determine TFENV_ROOT";
      target_file="$(readlink "${file_name}")";
    done;

    echo "$(pwd -P)/${file_name}";
  };

  TFENV_ROOT="$(cd "$(dirname "$(readlink_f "${0}")")/.." && pwd)";
  [ -n ${TFENV_ROOT} ] || early_death "Failed to 'cd \"\$(dirname \"\$(readlink_f \"${0}\")\")/..\" && pwd' while trying to determine TFENV_ROOT";
else
  TFENV_ROOT="${TFENV_ROOT%/}";
fi;
export TFENV_ROOT;

if [ -n "${TFENV_HELPERS:-""}" ]; then
  log 'debug' 'TFENV_HELPERS is set, not sourcing helpers again';
else
  [ "${TFENV_DEBUG:-0}" -gt 0 ] && echo "[DEBUG] Sourcing helpers from ${TFENV_ROOT}/lib/helpers.sh";
  if source "${TFENV_ROOT}/lib/helpers.sh"; then
    log 'debug' 'Helpers sourced successfully';
  else
    early_death "Failed to source helpers from ${TFENV_ROOT}/lib/helpers.sh";
  fi;
fi;

#####################
# Begin Script Body #
#####################

test_install_and_use() {
  # Takes a static version and the optional keyword to install it with
  local k="${2-""}";
  local v="${1}";
  tfenv install "${k}" || return 1;
  check_installed_version "${v}" || return 1;
  tfenv use "${k}" || return 1;
  check_active_version "${v}" || return 1;
  return 0;
};

test_install_and_use_overridden() {
  # Takes a static version and the optional keyword to install it with
  local k="${2-""}";
  local v="${1}";
  tfenv install "${k}" || return 1;
  check_installed_version "${v}" || return 1;
  tfenv use "${k}" || return 1;
  check_default_version "${v}" || return 1;
  return 0;
};

declare -a errors=();

log 'info' '### Test Suite: Install and Use';

tests__desc=(
  'latest version'
  'latest possibly-unstable version'
  'latest alpha'
  'latest beta'
  'latest rc'
  'latest possibly-unstable version from 0.11'
  '0.11.15-oci'
  'latest version matching regex'
  'specific version'
  'latest version matching regex, without latest word'
);

tests__kv=(
  "$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1),latest"
  "$(tfenv list-remote | head -n 1),latest:"
  "$(tfenv list-remote | grep 'alpha' | head -n 1),latest:alpha"
  "$(tfenv list-remote | grep 'beta' | head -n 1),latest:beta"
  "$(tfenv list-remote | grep 'rc' | head -n 1),latest:rc"
  "$(tfenv list-remote | grep '^0\.11\.' | head -n 1),latest:^0.11."
  '0.11.15-oci,0.11.15-oci'
  '0.8.8,latest:^0.8'
  "0.7.13,0.7.13"
  '0.8.8,:^0.8'
);

tests_count=${#tests__desc[@]};

declare desc kv k v;

for ((test_num=0; test_num<${tests_count}; ++test_num )) ; do
  cleanup || log 'error' 'Cleanup failed?!';
  desc=${tests__desc[${test_num}]};
  kv="${tests__kv[${test_num}]}";
  v="${kv%,*}";
  k="${kv##*,}";
  log 'info' "## Param Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} )";
  test_install_and_use "${v}" "${k}" \
    && log info "## Param Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} ) succeeded" \
    || error_and_proceed "## Param Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} ) failed";
done;

for ((test_num=0; test_num<${tests_count}; ++test_num )) ; do
  cleanup || log 'error' 'Cleanup failed?!';
  desc=${tests__desc[${test_num}]};
  kv="${tests__kv[${test_num}]}";
  v="${kv%,*}";
  k="${kv##*,}";
  log 'info' "## ./.terraform-version Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} )";
  log 'info' "Writing ${k} to ./.terraform-version";
  echo "${k}" > ./.terraform-version;
  test_install_and_use "${v}" \
    && log info "## ./.terraform-version Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} ) succeeded" \
    || error_and_proceed "## ./.terraform-version Test ${test_num}/${tests_count}: ${desc} ( ${k} / ${v} ) failed";
done;

cleanup || log 'error' 'Cleanup failed?!';
log 'info' '## ${HOME}/.terraform-version Test Preparation';

# 0.12.22 reports itself as 0.12.21 and breaks testing
declare v1="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | grep -v '0.12.22' | head -n 2 | tail -n 1)";
declare v2="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | grep -v '0.12.22' | head -n 1)";

if [ -f "${HOME}/.terraform-version" ]; then
  log 'info' "Backing up ${HOME}/.terraform-version to ${HOME}/.terraform-version.bup";
  mv "${HOME}/.terraform-version" "${HOME}/.terraform-version.bup";
fi;
log 'info' "Writing ${v1} to ${HOME}/.terraform-version";
echo "${v1}" > "${HOME}/.terraform-version";

log 'info' "## \${HOME}/.terraform-version Test 1/3: Install and Use ( ${v1} )";
test_install_and_use "${v1}" \
  && log info "## \${HOME}/.terraform-version Test 1/1: ( ${v1} ) succeeded" \
  || error_and_proceed "## \${HOME}/.terraform-version Test 1/1: ( ${v1} ) failed";

log 'info' "## \${HOME}/.terraform-version Test 2/3: Override Install with Parameter ( ${v2} )";
test_install_and_use_overridden "${v2}" "${v2}" \
  && log info "## \${HOME}/.terraform-version Test 2/3: ( ${v2} ) succeeded" \
  || error_and_proceed "## \${HOME}/.terraform-version Test 2/3: ( ${v2} ) failed";

log 'info' "## \${HOME}/.terraform-version Test 3/3: Override Use with Parameter ( ${v2} )";
(
  tfenv use "${v2}" || exit 1;
  check_default_version "${v2}" || exit 1;
) && log info "## \${HOME}/.terraform-version Test 3/3: ( ${v2} ) succeeded" \
  || error_and_proceed "## \${HOME}/.terraform-version Test 3/3: ( ${v2} ) failed";

log 'info' '## \${HOME}/.terraform-version Test Cleanup';
log 'info' "Deleting ${HOME}/.terraform-version";
rm "${HOME}/.terraform-version";
if [ -f "${HOME}/.terraform-version.bup" ]; then
  log 'info' "Restoring backup from ${HOME}/.terraform-version.bup to ${HOME}/.terraform-version";
  mv "${HOME}/.terraform-version.bup" "${HOME}/.terraform-version";
fi;

log 'info' 'Install invalid specific version';
cleanup || log 'error' 'Cleanup failed?!';

neg_tests__desc=(
  'specific version'
  'latest:word'
);

neg_tests__kv=(
  '9.9.9'
  "latest:word"
);

neg_tests_count=${#neg_tests__desc[@]};

for ((test_num=0; test_num<${neg_tests_count}; ++test_num )) ; do
  cleanup || log 'error' 'Cleanup failed?!';
  desc=${neg_tests__desc[${test_num}]}
  k="${neg_tests__kv[${test_num}]}";
  expected_error_message="No versions matching '${k}' found in remote";
  log 'info' "##  Invalid Version Test ${test_num}/${neg_tests_count}: ${desc} ( ${k} )";
  [ -z "$(tfenv install "${k}" 2>&1 | grep "${expected_error_message}")" ] \
    && error_and_proceed "Installing invalid version ${k}";
done;

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following install_and_use tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done
  log 'error' 'Test failure(s): install_and_use';
else
  log 'info' 'All install_and_use tests passed';
fi;

exit 0;
