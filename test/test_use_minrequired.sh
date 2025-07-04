#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required normal version (#.#.#)';

minv='1.6.0';

echo "terraform {
  required_version = \">=${minv}\"
}" > min_required.tf;

(
  tfenv install min-required;
  tfenv use min-required;
  check_active_version "${minv}";
) || error_and_proceed 'Min required version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required tagged version (#.#.#-tag#)'

minv='1.5.0-rc1'

echo "terraform {
    required_version = \">=${minv}\"
}" > min_required.tf;

(
  tfenv install min-required;
  tfenv use min-required;
  check_active_version "${minv}";
) || error_and_proceed 'Min required tagged-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required incomplete version (#.#.<missing>)'

minv='1.3';

echo "terraform {
  required_version = \">=${minv}\"
}" >> min_required.tf;

(
  tfenv install min-required;
  tfenv use min-required;
  check_active_version "${minv}.0";
) || error_and_proceed 'Min required incomplete-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required with TFENV_AUTO_INSTALL';

minv='1.2.0';

echo "terraform {
  required_version = \">=${minv}\"
}" >> min_required.tf;
echo 'min-required' > .terraform-version;

(
  TFENV_AUTO_INSTALL=true terraform version;
  check_active_version "${minv}";
) || error_and_proceed 'Min required auto-installed version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required with TFENV_AUTO_INSTALL & -chdir with rel path';

minv='1.1.0';

mkdir -p chdir-dir
echo "terraform {
  required_version = \">=${minv}\"
}" >> chdir-dir/min_required.tf;
echo 'min-required' > chdir-dir/.terraform-version

(
  TFENV_AUTO_INSTALL=true terraform -chdir=chdir-dir version;
  check_active_version "${minv}" chdir-dir;
) || error_and_proceed 'Min required version from -chdir does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required with TFENV_AUTO_INSTALL & -chdir with abs path';

minv='1.2.3';

mkdir -p chdir-dir
echo "terraform {
  required_version = \">=${minv}\"
}" >> chdir-dir/min_required.tf;
echo 'min-required' > chdir-dir/.terraform-version

(
  TFENV_AUTO_INSTALL=true terraform -chdir="${PWD}/chdir-dir" version;
  check_active_version "${minv}" chdir-dir;
) || error_and_proceed 'Min required version from -chdir does not match';

cleanup || log 'error' 'Cleanup failed?!';

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following use_minrequired tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'use_minrequired test failure(s)';
else
  log 'info' 'All use_minrequired tests passed.';
fi;

exit 0;
