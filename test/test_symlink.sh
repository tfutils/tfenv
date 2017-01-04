#!/usr/bin/env bash

[ -n "$TFENV_DEBUG" ] && set -x
source $(dirname $0)/helpers.sh

TFENV_BIN_DIR=/tmp/tfenv-test
rm -rf ${TFENV_BIN_DIR} && mkdir ${TFENV_BIN_DIR}
ln -s ${PWD}/bin/* ${TFENV_BIN_DIR}
export PATH="${TFENV_BIN_DIR}:${PATH}"

echo "### Test supporting symlink"
cleanup
tfenv install 0.8.2
tfenv use 0.8.2
check_version 0.8.2
