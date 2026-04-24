#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: version-file resolution';

log 'info' '## version-file: .terraform-version in current dir takes precedence';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.0.0' > .terraform-version;
  declare vf;
  vf="$(tfenv version-file)";
  [ "${vf}" == "$(pwd)/.terraform-version" ] || exit 1;
) && log 'info' '## version-file: current dir passed' \
  || error_and_proceed 'version-file did not find .terraform-version in current dir';

log 'info' '## version-file: search up directory tree to parent';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.0.0' > .terraform-version;
  mkdir -p subdir/nested;
  cd subdir/nested || exit 1;
  declare vf;
  vf="$(tfenv version-file)";
  # Should find the .terraform-version two levels up
  echo "${vf}" | grep -q '.terraform-version' || exit 1;
  declare content;
  content="$(cat "${vf}")";
  [ "${content}" == '1.0.0' ] || exit 1;
) && log 'info' '## version-file: parent dir traversal passed' \
  || error_and_proceed 'version-file did not traverse to parent .terraform-version';

log 'info' '## version-file: HOME fallback';
cleanup || log 'error' 'Cleanup failed?!';
(
  # Ensure no .terraform-version in current tree
  rm -f .terraform-version;
  # Write to HOME
  declare backup='';
  if [ -f "${HOME}/.terraform-version" ]; then
    backup="$(cat "${HOME}/.terraform-version")";
  fi;
  echo '0.15.0' > "${HOME}/.terraform-version";
  declare vf;
  vf="$(tfenv version-file)";
  [ "${vf}" == "${HOME}/.terraform-version" ] || exit 1;
  # Restore
  if [ -n "${backup}" ]; then
    echo "${backup}" > "${HOME}/.terraform-version";
  else
    rm -f "${HOME}/.terraform-version";
  fi;
) && log 'info' '## version-file: HOME fallback passed' \
  || error_and_proceed 'version-file did not fall back to HOME/.terraform-version';

log 'info' '## version-file: TFENV_CONFIG_DIR/version fallback';
cleanup || log 'error' 'Cleanup failed?!';
(
  # No .terraform-version anywhere, no HOME version
  rm -f .terraform-version;
  declare backup='';
  if [ -f "${HOME}/.terraform-version" ]; then
    backup="$(cat "${HOME}/.terraform-version")";
    rm -f "${HOME}/.terraform-version";
  fi;
  declare vf;
  vf="$(tfenv version-file)";
  [ "${vf}" == "${TFENV_CONFIG_DIR}/version" ] || exit 1;
  # Restore
  if [ -n "${backup}" ]; then
    echo "${backup}" > "${HOME}/.terraform-version";
  fi;
) && log 'info' '## version-file: TFENV_CONFIG_DIR fallback passed' \
  || error_and_proceed 'version-file did not fall back to TFENV_CONFIG_DIR/version';

log 'info' '## version-file: carriage return stripping';
cleanup || log 'error' 'Cleanup failed?!';
(
  # Write version with Windows CRLF line ending
  printf '1.6.1\r\n' > .terraform-version;
  tfenv install 1.6.1 || exit 1;
  tfenv use || exit 1;
  check_active_version 1.6.1 || exit 1;
) && log 'info' '## version-file: CR stripping passed' \
  || error_and_proceed 'version-file did not strip carriage returns from .terraform-version';

log 'info' '## version-file: respects working directory for resolution';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare tmpdir;
  tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t 'tfenv_vf_test')";
  echo '1.6.1' > "${tmpdir}/.terraform-version";
  cd "${tmpdir}" || exit 1;
  declare vf;
  vf="$(tfenv version-file)";
  [ "${vf}" == "${tmpdir}/.terraform-version" ] || exit 1;
  rm -rf "${tmpdir}";
) && log 'info' '## version-file: working directory resolution passed' \
  || error_and_proceed 'version-file did not find .terraform-version from working directory';


log 'info' '## version-file: comments and blank lines are stripped';
cleanup || log 'error' 'Cleanup failed?!';
(
  # Write a version file with comments and blank lines
  printf '# This is a comment\n\n# Use 1.0 series\nlatest:^1\\.0\\.\n\n' > .terraform-version;
  declare vn;
  vn="$(tfenv version-name)";
  # Should resolve to the version spec, not the comment
  [[ "${vn}" =~ ^1\.0\. ]] || { echo "Expected 1.0.x but got: ${vn}"; exit 1; };
) && log 'info' '## version-file: comments stripped passed' \
  || error_and_proceed 'version-file did not strip comments correctly';

log 'info' '## version-file: inline comments are stripped';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '1.0.0 # pinned for compatibility' > .terraform-version;
  declare vn;
  vn="$(tfenv version-name)";
  [ "${vn}" == '1.0.0' ] || { echo "Expected 1.0.0 but got: ${vn}"; exit 1; };
) && log 'info' '## version-file: inline comments passed' \
  || error_and_proceed 'version-file did not strip inline comments correctly';

finish_tests 'version_file';

exit 0;
