#!/usr/bin/env bash

# Determines the engine to use based on the variable $TFENV_ENGINE or, if not
# set, using a heuristic based on the discovered version files.
#
# If TFENV_ENGINE is not set, we determine the version file if the engine is
# terraform and tofu and if the files are the same, default to the terraform
# engine, otherwise, take the file that has the longest directory path.  This
# assumes that files are found in the same directory path and the one closest to
# where we are looking is the desired engine.

set -uo pipefail;


function tfenv-resolve-engine() {
  if [[ -n "${TFENV_ENGINE:-""}" ]]; then
    case "${TFENV_ENGINE}" in
      'terraform'|'tofu')
        echo -n "${TFENV_ENGINE}"
        ;;
      *)
        log 'error' "Unknown TFENV_ENGINE value ${TFENV_ENGINE}.  Must be either terraform or tofu"
        ;;
    esac
  else
    export TFENV_ENGINE=terraform
    local terraform_version_file
    terraform_version_file="$(tfenv-version-file)"
    export TFENV_ENGINE=tofu
    local tofu_version_file
    tofu_version_file="$(tfenv-version-file)"

    local terraform_version_dir
    local tofu_version_dir
    terraform_version_dir="$(dirname "${terraform_version_file}")"
    tofu_version_dir="$(dirname "${tofu_version_file}")"

    if [ "${terraform_version_file}" = "${tofu_version_file}" ]; then
      echo -n "terraform"
    elif [ "${#terraform_version_dir}" -lt "${#tofu_version_dir}" ]; then
      echo -n "tofu"
    else
      echo -n "terraform"
    fi
  fi
};
export -f tfenv-resolve-engine;
