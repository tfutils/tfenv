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


log 'info' '### Install latest-allowed-stable normal version (#.#.#)';

echo "terraform {
  required_version = \"~> 1.1.0\"
}" > latest_allowed_stable.tf;

(
  tfenv install latest-allowed-stable;
  tfenv use latest-allowed-stable;
  check_active_version 1.1.9;
) || error_and_proceed 'Latest allowed stable version does not match';

cleanup || log 'error' 'Cleanup failed?!';

log 'info' '### Install latest-allowed-stable tagged version (#.#.#-tag#)'

echo "terraform {
    required_version = \"<=0.13.0-rc1\"
}" > latest_allowed_stable.tf;

(
  tfenv install latest-allowed-stable;
  tfenv use latest-allowed-stable;
  check_active_version 0.12.31;
) || error_and_proceed 'Latest allowed stable tagged-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


echo "terraform {
    required_version = \"<=1.1.0-alpha20211006\"
}" > latest_allowed_stable.tf;

(
  tfenv install latest-allowed-stable;
  tfenv use latest-allowed-stable;
  check_active_version 1.0.11;
) || error_and_proceed 'Latest allowed stable tagged-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed-stable incomplete version (#.#.<missing>)'

echo "terraform {
  required_version = \"~> 0.12\"
}" >> latest_allowed_stable.tf;

(
  tfenv install latest-allowed-stable;
  tfenv use latest-allowed-stable;
  check_active_version 0.15.5;
) || error_and_proceed 'Latest allowed stable incomplete-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed-stable with TFENV_AUTO_INSTALL';

echo "terraform {
  required_version = \"~> 1.0.0\"
}" >> latest_allowed_stable.tf;
echo 'latest-allowed-stable' > .terraform-version;

(
  TFENV_AUTO_INSTALL=true terraform version;
  check_active_version 1.0.11;
) || error_and_proceed 'Latest allowed stable auto-installed version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install latest-allowed with TFENV_AUTO_INSTALL & -chdir';

mkdir -p chdir-dir
echo "terraform {
  required_version = \"~> 0.14.3\"
}" >> chdir-dir/latest_allowed.tf;
echo 'latest-allowed-stable' > chdir-dir/.terraform-version

(
  TFENV_AUTO_INSTALL=true terraform -chdir=chdir-dir version;
  check_active_version 0.14.11 chdir-dir;
) || error_and_proceed 'Latest allowed version from -chdir does not match';

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
