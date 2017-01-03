#!/usr/bin/env bash

[ -n "$TFENV_DEBUG" ] && set -x

TFENV_BIN_SYM=/tmp/tfenv
rm -rf ${TFENV_BIN_SYM}
ln -s ${PWD}/bin/tfenv ${TFENV_BIN_SYM}

echo "### Test supporting symlink"
${TFENV_BIN_SYM} --help
