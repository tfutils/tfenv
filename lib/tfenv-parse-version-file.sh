#!/usr/bin/env bash

set -uo pipefail;

# Ensure GNU grep
check_dependencies;

function tfenv-parse-version-file() {
  local version_file="$1"
  log 'debug' "Parsing version from file ${version_file}"

  grep --invert-match '^#' "${version_file}";
}
export -f tfenv-parse-version-file;
