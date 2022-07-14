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

declare -a errors=();

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required normal version (#.#.#)';

minv='0.8.0';

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

minv='0.13.0-rc1'

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

minv='0.12';

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

minv='1.0.0';

echo "terraform {
  required_version = \">=${minv}\"
}" >> min_required.tf;
echo 'min-required' > .terraform-version;

(
  TFENV_AUTO_INSTALL=true terraform version;
  check_active_version "${minv}";
) || error_and_proceed 'Min required auto-installed version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install min-required with TFENV_AUTO_INSTALL & -chdir';

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
