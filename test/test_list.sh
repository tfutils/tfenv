#!/usr/bin/env bash

rm -rf ${TFENV_ROOT}/versions

for v in 0.6.1 0.6.2 0.6.16 0.7.0; do
  tfenv install ${v}
done

result=$(tfenv list)
expected="$(cat << EOS
0.6.1
0.6.2
0.6.16
0.7.0
EOS
)"
if [ "${expected}" != "${result}" ]; then
  echo "Expected: ${expected}, Got: ${result}" 1>&2
  exit 1
fi
