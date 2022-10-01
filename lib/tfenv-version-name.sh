#!/usr/bin/env bash

set -uo pipefail;

function tfenv-version-name() {
  if [[ -z "${TFENV_TERRAFORM_VERSION:-""}" ]]; then
    log 'debug' 'We are not hardcoded by a TFENV_TERRAFORM_VERSION environment variable';

    TFENV_VERSION_FILE="$(tfenv-version-file)" \
      && log 'debug' "TFENV_VERSION_FILE retrieved from tfenv-version-file: ${TFENV_VERSION_FILE}" \
      || log 'error' 'Failed to retrieve TFENV_VERSION_FILE from tfenv-version-file';

    TFENV_VERSION="$(cat "${TFENV_VERSION_FILE}" || true)" \
      && log 'debug' "TFENV_VERSION specified in TFENV_VERSION_FILE: ${TFENV_VERSION}";

    TFENV_VERSION_SOURCE="${TFENV_VERSION_FILE}";

  else
    TFENV_VERSION="${TFENV_TERRAFORM_VERSION}" \
      && log 'debug' "TFENV_VERSION specified in TFENV_TERRAFORM_VERSION environment variable: ${TFENV_VERSION}";

    TFENV_VERSION_SOURCE='TFENV_TERRAFORM_VERSION';
  fi;

  local auto_install="${TFENV_AUTO_INSTALL:-true}";

  if [[ "${TFENV_VERSION}" == "min-required" ]]; then
    log 'debug' 'TFENV_VERSION uses min-required keyword, looking for a required_version in the code';

    local potential_min_required="$(tfenv-min-required)";
    if [[ -n "${potential_min_required}" ]]; then
      log 'debug' "'min-required' converted to '${potential_min_required}'";
      TFENV_VERSION="${potential_min_required}" \
      TFENV_VERSION_SOURCE='terraform{required_version}';
    else
      log 'error' 'Specifically asked for min-required via terraform{required_version}, but none found';
    fi;
  fi;

  if [[ "${TFENV_VERSION}" =~ ^latest.*$ ]]; then
    log 'debug' "TFENV_VERSION uses 'latest' keyword: ${TFENV_VERSION}";

    if [[ "${TFENV_VERSION}" == latest-allowed ]]; then
        TFENV_VERSION="$(tfenv-resolve-version)";
        log 'debug' "Resolved latest-allowed to: ${TFENV_VERSION}";
    fi;

    if [[ "${TFENV_VERSION}" =~ ^latest\:.*$ ]]; then
      regex="${TFENV_VERSION##*\:}";
      log 'debug' "'latest' keyword uses regex: ${regex}";
    else
      regex="^[0-9]\+\.[0-9]\+\.[0-9]\+$";
      log 'debug' "Version uses latest keyword alone. Forcing regex to match stable versions only: ${regex}";
    fi;

    declare local_version='';
    if [[ -d "${TFENV_CONFIG_DIR}/versions" ]]; then
      local_version="$(\find "${TFENV_CONFIG_DIR}/versions/" -type d -exec basename {} \; \
        | tail -n +2 \
        | sort -t'.' -k 1nr,1 -k 2nr,2 -k 3nr,3 \
        | grep -e "${regex}" \
        | head -n 1)";

      log 'debug' "Resolved ${TFENV_VERSION} to locally installed version: ${local_version}";
    elif [[ "${auto_install}" != "true" ]]; then
      log 'error' 'No versions of terraform installed and TFENV_AUTO_INSTALL is not true. Please install a version of terraform before it can be selected as latest';
    fi;

    if [[ "${auto_install}" == "true" ]]; then
      log 'debug' "Using latest keyword and auto_install means the current version is whatever is latest in the remote. Trying to find the remote version using the regex: ${regex}";
      remote_version="$(tfenv-list-remote | grep -e "${regex}" | head -n 1)";
      if [[ -n "${remote_version}" ]]; then
          if [[ "${local_version}" != "${remote_version}" ]]; then
            log 'debug' "The installed version '${local_version}' does not much the remote version '${remote_version}'";
            TFENV_VERSION="${remote_version}";
          else
            TFENV_VERSION="${local_version}";
          fi;
      else
        log 'error' "No versions matching '${requested}' found in remote";
      fi;
    else
      if [[ -n "${local_version}" ]]; then
        TFENV_VERSION="${local_version}";
      else
        log 'error' "No installed versions of terraform matched '${TFENV_VERSION}'";
      fi;
    fi;
  else
    log 'debug' 'TFENV_VERSION does not use "latest" keyword';

    # Accept a v-prefixed version, but strip the v.
    if [[ "${TFENV_VERSION}" =~ ^v.*$ ]]; then
      log 'debug' "Version Requested is prefixed with a v. Stripping the v.";
      TFENV_VERSION="${TFENV_VERSION#v*}";
    fi;
  fi;

  if [[ -z "${TFENV_VERSION}" ]]; then
    log 'error' "Version could not be resolved (set by ${TFENV_VERSION_SOURCE} or tfenv use <version>)";
  fi;

  if [[ "${TFENV_VERSION}" == min-required ]]; then
    TFENV_VERSION="$(tfenv-min-required)";
  fi;

  if [[ ! -d "${TFENV_CONFIG_DIR}/versions/${TFENV_VERSION}" ]]; then
    log 'debug' "version '${TFENV_VERSION}' is not installed (set by ${TFENV_VERSION_SOURCE})";
  fi;

  echo "${TFENV_VERSION}";
};
export -f tfenv-version-name;

