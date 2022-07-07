function tfenv-version-name() {
  if [[ -z "${TFENV_TERRAFORM_VERSION:-""}" ]]; then
    TFENV_VERSION_FILE="$(tfenv-version-file)" \
      && log 'debug' "TFENV_VERSION_FILE retrieved from tfenv-version-file: ${TFENV_VERSION_FILE}" \
      || log 'error' 'Failed to retrieve TFENV_VERSION_FILE from tfenv-version-file';

    if [ "${TFENV_VERSION_FILE}" = "${TFENV_CONFIG_DIR}/version" ]; then
      log 'debug' 'Tryng to set version from "required_version" under "terraform" section'

      versions="$( echo $(cat {*.tf,*.tf.json} 2>/dev/null | grep -h required_version) | grep  -o '\([0-9]\+\.\?\)\{2,3\}\(-[a-z]\+[0-9]\+\)\?')";
      if [[ "${versions}" =~ ([~=!<>]{0,2}[[:blank:]]*[0-9]+[0-9.]+)[^0-9]*(-[a-z]+[0-9]+)? ]]; then
        found_min_required="${BASH_REMATCH[1]}${BASH_REMATCH[2]}"
        if [[ "${found_min_required}" =~ ^!=.+ ]]; then
          log 'debug' "required_version is a negation - we cannot guess the desired one, skipping.";
        else
          found_min_required="$(echo "$found_min_required")";

          # Probably not an advisable way to choose a terraform version,
          # but this is the way this functionality works in terraform:
          # add .0 to versions without a minor and/or patch version (e.g. 12.0)
          while ! [[ "${found_min_required}" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; do
            found_min_required="${found_min_required}.0";
          done;
          TFENV_VERSION="${found_min_required}";
        fi;
      fi;

      TFENV_VERSION_SOURCE='terraform{required_version}'
    fi;

    if [[ -z "${TFENV_VERSION:-""}" ]]; then
      TFENV_VERSION="$(cat "${TFENV_VERSION_FILE}" || true)" \
        && log 'debug' "TFENV_VERSION specified in TFENV_VERSION_FILE: ${TFENV_VERSION}";

      TFENV_VERSION_SOURCE="${TFENV_VERSION_FILE}";
    fi;
  else
    TFENV_VERSION="${TFENV_TERRAFORM_VERSION}" \
      && log 'debug' "TFENV_VERSION specified in TFENV_TERRAFORM_VERSION: ${TFENV_VERSION}";

    TFENV_VERSION_SOURCE='TFENV_TERRAFORM_VERSION';
  fi;

  if [[ "${TFENV_VERSION}" =~ ^latest.*$ ]]; then
    log 'debug' "TFENV_VERSION uses 'latest' keyword: ${TFENV_VERSION}";

    [ -d "${TFENV_CONFIG_DIR}/versions" ] \
      || log 'error' 'No versions of terraform installed. Please install one with: tfenv install';

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
    fi;

    if [[ "${TFENV_AUTO_INSTALL:-true}" == "true" ]]; then
      log 'debug' "Trying to find the remote version using the regex: ${regex}";
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
      log 'debug' "Version Requested is prefixed with a v. Stripping the v."
      TFENV_VERSION="${TFENV_VERSION#v*}";
    fi;
  fi;

  if [[ -z "${TFENV_VERSION}" ]]; then
    log 'error' "Version could not be resolved (set by ${TFENV_VERSION_SOURCE} or tfenv use <version>)";
  fi;

  if [[ ! -d "${TFENV_CONFIG_DIR}/versions/${TFENV_VERSION}" ]]; then
    log 'debug' "version '${TFENV_VERSION}' is not installed (set by ${TFENV_VERSION_SOURCE})";
  fi;

  echo "${TFENV_VERSION}";
}
export -f tfenv-version-name;

