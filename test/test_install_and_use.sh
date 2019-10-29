#!/usr/bin/env bash

set -uo pipefail;

# Ensure we can execute standalone
if [ -n "${TFENV_ROOT:-""}" ]; then
  if [ "${TFENV_DEBUG:-0}" -gt 1 ]; then
    [ -n "${TFENV_HELPERS:-""}" ] \
      && log 'debug' "TFENV_ROOT already defined as ${TFENV_ROOT}" \
      || echo "[DEBUG] TFENV_ROOT already defined as ${TFENV_ROOT}" >&2;
  fi;
else
  export TFENV_ROOT="$(cd "$(dirname "${0}")/.." && pwd)";
  if [ "${TFENV_DEBUG:-0}" -gt 1 ]; then
    [ -n "${TFENV_HELPERS:-""}" ] \
      && log 'debug' "TFENV_ROOT declared as ${TFENV_ROOT}" \
      || echo "[DEBUG] TFENV_ROOT declared as ${TFENV_ROOT}" >&2;
  fi;
fi;

if [ -n "${TFENV_HELPERS:-""}" ]; then
  log 'debug' 'TFENV_HELPERS is set, not sourcing helpers again';
else
  [ "${TFENV_DEBUG:-0}" -gt 1 ] && echo "[DEBUG] Sourcing helpers from ${TFENV_ROOT}/lib/helpers.sh" >&2;
  if source "${TFENV_ROOT}/lib/helpers.sh"; then
    log 'debug' 'Helpers sourced successfully';
  else
    echo "[ERROR] Failed to source helpers from ${TFENV_ROOT}/lib/helpers.sh" >&2;
    exit 1;
  fi;
fi;

test_install_and_use() {
  # Takes a static version and the optional keyword to install it with
  local k="${2-""}";
  local v="${1}";
  tfenv install "${k}" || return 1;
  check_version "${v}" || return 1;
  return 0;
};

declare -a errors=();

log 'info' '### Test Suite: Install and Use'

declare -A string_tests=();

string_tests['latest version']="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1),latest";
string_tests['latest possibly-unstable version']="$(tfenv list-remote | head -n 1),latest:";
string_tests['latest alpha']="$(tfenv list-remote | grep 'alpha' | head -n 1),latest:alpha";
string_tests['latest beta']="$(tfenv list-remote | grep 'beta' | head -n 1),latest:beta";
string_tests['latest rc']="$(tfenv list-remote | grep 'rc' | head -n 1),latest:rc";
string_tests['latest possibly-unstable version from 0.11']="$(tfenv list-remote | grep '^0\.11\.' | head -n 1),latest:^0.11.";
string_tests['0.11.15-oci']='0.11.15-oci,0.11.15-oci';
string_tests['latest version matching regex']='0.8.8,latest:^0.8';
string_tests['specific version']="0.7.13,0.7.13";

declare kv k v;
declare -i test_num=1;

for desc in "${!string_tests[@]}"; do
  cleanup || log 'error' 'Cleanup failed?!';
  kv="${string_tests[${desc}]}";
  v="${kv%,*}";
  k="${kv##*,}";
  log 'info' "## Param Test ${test_num}/${#string_tests[*]}: ${desc} ( ${k} / ${v} )";
  test_install_and_use "${v}" "${k}" \
    && log info "## Param Test ${test_num}/${#string_tests[*]}: ${desc} ( ${k} / ${v} ) succeeded" \
    || error_and_proceed "## Param Test ${test_num}/${#string_tests[*]}: ${desc} ( ${k} / ${v} ) failed";
  test_num+=1;
done;

test_num=1;
for desc in "${!string_tests[@]}"; do
  cleanup || log 'error' 'Cleanup failed?!';
  kv="${string_tests[${desc}]}";
  v="${kv%,*}";
  k="${kv##*,}";
  log 'info' "## ./.terraform-version Test ${test_num}/${string_tests[*]}: ${desc} ( ${k} / ${v} )";
  log 'info' "Writing ${k} to ./.terraform-version";
  echo "${k}" > ./.terraform-version;
  test_install_and_use "${v}" \
    && log info "## ./.terraform-version Test ${test_num}/${#string_tests[*]}: ${desc} ( ${k} / ${v} ) succeeded" \
    || error_and_proceed "## ./.terraform-version Test ${test_num}/${#string_tests[*]}: ${desc} ( ${k} / ${v} ) failed";
  test_num+=1;
done;

cleanup || log 'error' 'Cleanup failed?!';
log 'info' '## \${HOME}/.terraform-version Test Preparation';
declare v1="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 2 | tail -n 1)";
declare v2="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1)";
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
test_install_and_use "${v2}" "${v2}" \
  && log info "## \${HOME}/.terraform-version Test 2/3: ( ${v2} ) succeeded" \
  || error_and_proceed "## \${HOME}/.terraform-version Test 2/3: ( ${v2} ) failed";

log 'info' "## \${HOME}/.terraform-version Test 3/3: Override Use with Parameter ( ${v2} )";
(
  tfenv use "${v2}" || exit 1;
  check_version "${v2}" || exit 1;
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


declare -A neg_tests=();
neg_tests['specific version']="9.9.9";
neg_tests['latest:word']="latest:word";

test_num=1;

for desc in "${!neg_tests[@]}"; do
  cleanup || log 'error' 'Cleanup failed?!';
  k="${neg_tests[${desc}]}";
  expected_error_message="No versions matching '${k}' found in remote";
  log 'info' "##  Invalid Version Test ${test_num}/${#neg_tests[*]}: ${desc} ( ${k} )";
  [ -z "$(tfenv install "${k}" 2>&1 | grep "${expected_error_message}")" ] \
    && error_and_proceed "Installing invalid version ${k}";
  test_num+=1;
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
