#!/usr/bin/env bash

[ -n "$TFENV_DEBUG" ] && set -x
source $(dirname $0)/helpers.sh

echo "### List local versions"
cleanup

for v in 0.6.16 0.7.0-rc4 0.7.2 0.7.13 0.8.0-beta2; do
  tfenv install ${v}
done

result=$(tfenv list)
expected="$(cat << EOS
0.8.0-beta2
0.7.13
0.7.2
0.7.0-rc4
0.6.16
EOS
)"
if [ "${expected}" != "${result}" ]; then
  echo "Expected: ${expected}, Got: ${result}" 1>&2
  exit 1
fi
