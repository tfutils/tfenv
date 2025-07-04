#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed normal version (#.#.#)';

echo "terraform {
  required_version = \"~> 1.1.0\"
}" > latest_allowed.tf;

(
  tfenv install latest-allowed;
  tfenv use latest-allowed;
  check_active_version 1.1.9;
) || error_and_proceed 'Latest allowed version does not match. Requested: "~> 1.1.0", Expected: 1.1.9';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed tagged version (#.#.#-tag#)'

echo "terraform {
    required_version = \"<=0.13.0-rc1\"
}" > latest_allowed.tf;

(
  tfenv install latest-allowed;
  tfenv use latest-allowed;
  check_active_version 0.13.0-rc1;
) || error_and_proceed 'Latest allowed tagged-version does not match. Requested: "<=0.13.0-rc1", Expected: 0.13.0-rc1';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed incomplete version (#.#.<missing>)'

echo "terraform {
  required_version = \"~> 0.12\"
}" >> latest_allowed.tf;

(
  tfenv install latest-allowed;
  tfenv use latest-allowed;
  check_active_version 0.15.5;
) || error_and_proceed 'Latest allowed incomplete-version does not match. Requested: "~> 0.12", Expected: 0.15.5';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed with TFENV_AUTO_INSTALL';

echo "terraform {
  required_version = \"~> 1.0.0\"
}" >> latest_allowed.tf;
echo 'latest-allowed' > .terraform-version;

(
  TFENV_AUTO_INSTALL=true terraform version;
  check_active_version 1.0.11;
) || error_and_proceed 'Latest allowed auto-installed version does not match. Requested: "~> 1.0.0", Expected: 1.0.11';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed with TFENV_AUTO_INSTALL & -chdir';

mkdir -p chdir-dir
echo "terraform {
  required_version = \"~> 0.14.3\"
}" >> chdir-dir/latest_allowed.tf;
echo 'latest-allowed' > chdir-dir/.terraform-version

(
  TFENV_AUTO_INSTALL=true terraform -chdir=chdir-dir version;
  check_active_version 0.14.11 chdir-dir;
) || error_and_proceed 'Latest allowed version from -chdir does not match. Requested: "~> 0.14.3", Expected: 0.14.11';

cleanup || log 'error' 'Cleanup failed?!';

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following use_latestallowed tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'use_latestallowed test failure(s)';
else
  log 'info' 'All use_latestallowed tests passed.';
fi;

exit 0;
