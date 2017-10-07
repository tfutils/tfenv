#!/usr/bin/env bash

declare -a errors

function error_and_proceed() {
  errors+=("${1}")
  echo -e "tfenv: ${0}: Test Failed: ${1}" >&2
}

function error_and_die() {
  echo -e "tfenv: ${0}: ${1}" >&2
  exit 1
}

[ -n "${TFENV_DEBUG}" ] && set -x
source $(dirname $0)/helpers.sh \
  || error_and_die "Failed to load test helpers: $(dirname $0)/helpers.sh"

echo "### List local versions"
cleanup || error_and_die "Cleanup failed?!"

for v in 0.7.2 0.7.13 0.9.1 0.9.2 0.9.11; do
  tfenv install ${v} || error_and_proceed "Install of version ${v} failed"
done

result=$(tfenv list)
expected="$(cat << EOS
* 0.9.11 (set by $(tfenv version-file))
  0.9.2
  0.9.1
  0.7.13
  0.7.2
EOS
)"

if [ "${expected}" != "${result}" ]; then
  error_and_proceed "List mismatch.\nExpected:\n${expected}\nGot:\n${result}"
fi

if [ ${#errors[@]} -gt 0 ]; then
  echo -e "\033[0;31m===== The following list tests failed =====\033[0;39m" >&2
  for error in "${errors[@]}"; do
    echo -e "\t${error}"
  done
  exit 1
else
  echo -e "\033[0;32mAll list tests passed.\033[0;39m"
fi;
exit 0
