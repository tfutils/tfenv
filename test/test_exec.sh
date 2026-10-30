#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: terraform shim (tfenv-exec)';

cleanup || log 'error' 'Cleanup failed?!';
tfenv install 1.6.1 || early_death 'Failed to install 1.6.1 for shim tests';
tfenv use 1.6.1 || early_death 'Failed to use 1.6.1 for shim tests';

# zsh's bashcompinit runs `complete -C` commands with no arguments, and Bash 3.2
# (macOS /bin/bash) treats "$@" as unbound under `set -u` when it is empty (#532).
# Run the shim under each available bash so the system bash is exercised on macOS
# even when a newer bash is first on PATH.
for shell in bash /bin/bash; do
  command -v "${shell}" >/dev/null 2>&1 || continue;
  log 'info' "## terraform (no args) under ${shell}: passes through to terraform usage";
  (
    declare output;
    output="$("${shell}" "${TFENV_ROOT}/bin/terraform" 2>&1 || true)";
    echo "${output}" | grep -q 'unbound variable' && exit 1;
    echo "${output}" | grep -q 'Usage: terraform' || exit 1;
  ) && log 'info' "## terraform (no args) under ${shell}: passed" \
    || error_and_proceed "terraform with no args under ${shell} did not reach terraform usage";
done;

log 'info' '## terraform (with args): arguments are passed through to terraform';
(
  declare output;
  output="$(terraform version -json)" || exit 1;
  echo "${output}" | grep -q 'terraform_version' || exit 1;
  echo "${output}" | grep -q '1.6.1' || exit 1;
) && log 'info' '## terraform (with args): passed' \
  || error_and_proceed 'terraform arguments were not passed through';

log 'info' '## terraform -chdir with a space in the path: argument is not word-split';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.6.1' > .terraform-version;
  mkdir -p 'chdir-dir/with space';
  echo '1.6.0' > 'chdir-dir/with space/.terraform-version';
  declare output;
  output="$(terraform '-chdir=chdir-dir/with space' version 2>&1)" || exit 1;
  echo "${output}" | grep -q 'Terraform v1.6.0' || exit 1;
) && log 'info' '## terraform -chdir with a space in the path: passed' \
  || error_and_proceed 'terraform -chdir with a space in the path did not resolve the version from that directory';

cleanup || log 'error' 'Cleanup failed?!';

finish_tests 'exec';

exit 0;
# vim: set syntax=bash tabstop=2 softtabstop=2 shiftwidth=2 expandtab smarttab :
