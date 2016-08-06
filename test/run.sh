#!/usr/bin/env bash

errors=()
for t in `ls -1 $(dirname $0) | grep 'test_'`; do
  bash $(dirname $0)/${t} || errors+=( ${t} )
done

if [ ${#errors[@]} -ne 0 ];then
  echo -e "\033[0;31m===== Following tests failed ====\033[0;39m" 1>&2
  echo "${errors[@]}"
  exit 1
fi
