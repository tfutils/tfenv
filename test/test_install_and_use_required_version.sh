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


log 'info' '### Install required_version normal version (#.#.#)';

reqv='0.14.11';

echo "terraform {
  required_version = \">=${reqv}\"
}" > required_version.tf;

(
  tfenv install;
  tfenv use;
  check_active_version "${reqv}";
) || error_and_proceed 'required_version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install required_version tagged version (#.#.#-tag#)'

reqv='0.14.0-rc1'

echo "terraform {
    required_version = \">=${reqv}\"
}" > required_version.tf;

(
  tfenv install;
  tfenv use;
  check_active_version "${reqv}";
) || error_and_proceed 'required_version tagged-version does not match';

cleanup || log 'error' 'Cleanup failed?!';


log 'info' '### Install required_version incomplete version (#.#.<missing>)'

reqv='0.14';

echo "terraform {
  required_version = \">=${reqv}\"
}" > required_version.tf;

(
  tfenv install;
  tfenv use;
  check_active_version "${reqv}.0";
) || error_and_proceed 'required_version incomplete-version does not match';

cleanup || log 'error' 'Cleanup failed?!';

log 'info' '### Test auto-installing when running terraform';

reqv='0.14.11';

echo "terraform {
  required_version = \">=${reqv}\"
}" > required_version.tf;

(
  check_active_version "${reqv}";
) || error_and_proceed 'required_version does not match';

cleanup || log 'error' 'Cleanup failed?!';

if [ "${#errors[@]}" -gt 0 ]; then
  log 'warn' '===== The following required_version tests failed =====';
  for error in "${errors[@]}"; do
    log 'warn' "\t${error}";
  done;
  log 'error' 'required_version test failure(s)';
else
  log 'info' 'All required_version tests passed.';
fi;

exit 0;
