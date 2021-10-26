#!/usr/bin/env bash

set -uo pipefail;

function tfenv-min-required() {
  local path="${1:-${TFENV_DIR:-.}}";

  local versions="$( echo $(cat ${path}/{*.tf,*.tf.json} 2>/dev/null | grep -Eh '^\s*[^#]*\s*required_version') | grep -o '[~=!<>]\{0,2\}\s*\([0-9]\+\.\?\)\{2,3\}\(-[a-z]\+[0-9]\+\)\?')";

  if [[ "${versions}" =~ ([~=!<>]{0,2}[[:blank:]]*)([0-9]+[0-9.]+)[^0-9]*(-[a-z]+[0-9]+)? ]]; then
    qualifier="${BASH_REMATCH[1]}";
    found_min_required="${BASH_REMATCH[2]}${BASH_REMATCH[3]}";
    if [[ "${qualifier}" =~ ^!= ]]; then
      log 'debug' "required_version is a negation - we cannot guess the desired one, skipping.";
    else
      local min_required_file="$(grep -Hn required_version ${path}/{*.tf,*.tf.json} 2>/dev/null | xargs)";

      # Probably not an advisable way to choose a terraform version,
      # but this is the way this functionality works in terraform:
      # add .0 to versions without a minor and/or patch version (e.g. 12.0)
      while ! [[ "${found_min_required}" =~ [0-9]+\.[0-9]+\.[0-9]+ ]]; do
        found_min_required="${found_min_required}.0";
      done;

      log 'debug' "Determined min-required to be '${found_min_required}' (${min_required_file})";
      echo -en "${found_min_required}";
      return;
    fi;
  fi;

  log 'debug' 'Appropriate required_version not found, skipping min-required';
};
export -f tfenv-min-required;
