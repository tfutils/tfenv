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


log 'info' '### Detect min-required when working directory path contains spaces (#537)';

spaced_dir="${PWD}/project with spaces";
mkdir -p "${spaced_dir}";
minv='1.6.0';
echo "terraform {
  required_version = \">=${minv}\"
}" > "${spaced_dir}/versions.tf";

(
  cd "${spaced_dir}" || exit 1;
  # bin/tfenv exports TFENV_DIR="${PWD}" before dispatching. Paths with spaces
  # used to break unquoted ${path}/{*.tf,*.tf.json} globs in tfenv-min-required,
  # so install/resolve-version min-required failed with an empty version.
  export TFENV_DIR="${PWD}";
  detected="$(tfenv min-required)" || exit 1;
  [ "${detected}" = "${minv}" ] || exit 1;
  resolved="$(tfenv-resolve-version min-required)" || exit 1;
  [ "${resolved}" = "${minv}:^${minv}$" ] || exit 1;
) || error_and_proceed 'Min-required detection failed when path contains spaces';

rm -rf "${spaced_dir}";
cleanup || log 'error' 'Cleanup failed?!';

finish_tests 'use_minrequired';

exit 0;
