#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: Pre-release version filtering';

log 'info' '## Numeric regex excludes pre-releases: latest:^0.14.';
cleanup || log 'error' 'Cleanup failed?!';
(
  # latest:^0.14. with a numeric-only regex should skip 0.14.0-rc1 etc.
  declare resolved;
  resolved="$(tfenv list-remote | grep -e '^0\.14\.' | grep -v '-' | head -n 1)";
  log 'info' "Expected stable version: ${resolved}";
  tfenv install "latest:^0.14." || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## Numeric regex excludes pre-releases: passed' \
  || error_and_proceed 'latest:^0.14. did not filter pre-releases';

log 'info' '## Alphabetic regex includes pre-releases: latest:alpha';
cleanup || log 'error' 'Cleanup failed?!';
(
  # latest:alpha should match alpha versions
  declare resolved;
  resolved="$(tfenv list-remote | grep 'alpha' | head -n 1)";
  log 'info' "Expected alpha version: ${resolved}";
  tfenv install 'latest:alpha' || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## Alphabetic regex includes pre-releases: passed' \
  || error_and_proceed 'latest:alpha did not include pre-release versions';

log 'info' '## Alphabetic regex includes pre-releases: latest:rc';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare resolved;
  resolved="$(tfenv list-remote | grep 'rc' | head -n 1)";
  log 'info' "Expected rc version: ${resolved}";
  tfenv install 'latest:rc' || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## latest:rc includes rc versions: passed' \
  || error_and_proceed 'latest:rc did not include rc versions';

log 'info' '## Alphabetic regex includes pre-releases: latest:beta';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare resolved;
  resolved="$(tfenv list-remote | grep 'beta' | head -n 1)";
  log 'info' "Expected beta version: ${resolved}";
  tfenv install 'latest:beta' || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## latest:beta includes beta versions: passed' \
  || error_and_proceed 'latest:beta did not include beta versions';

log 'info' '## Numeric regex on 0.11 excludes oci suffix';
cleanup || log 'error' 'Cleanup failed?!';
(
  # latest:^0.11. should NOT match 0.11.15-oci since regex is numeric-only
  declare resolved;
  resolved="$(tfenv list-remote | grep -e '^0\.11\.' | grep -v '-' | head -n 1)";
  log 'info' "Expected stable version: ${resolved}";
  tfenv install "latest:^0.11." || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## Numeric regex on 0.11 excludes oci: passed' \
  || error_and_proceed 'latest:^0.11. did not exclude oci-suffixed versions';

log 'info' '## Bare latest (no regex) gives stable version';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare resolved;
  resolved="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1)";
  log 'info' "Expected latest stable: ${resolved}";
  tfenv install latest || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## Bare latest gives stable version: passed' \
  || error_and_proceed 'latest did not resolve to latest stable version';

log 'info' '## latest: (with colon, no regex) includes pre-releases';
cleanup || log 'error' 'Cleanup failed?!';
(
  declare resolved;
  resolved="$(tfenv list-remote | head -n 1)";
  log 'info' "Expected latest including pre-release: ${resolved}";
  tfenv install 'latest:' || exit 1;
  check_installed_version "${resolved}" || exit 1;
) && log 'info' '## latest: includes pre-releases: passed' \
  || error_and_proceed 'latest: did not include pre-release versions';

finish_tests 'prerelease_filter';

exit 0;
