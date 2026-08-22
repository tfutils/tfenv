#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: dependencies';

log 'info' '## grep_is_sufficient: accepts the grep already on PATH';
(
  grep_is_sufficient grep || exit 1;
  exit 0;
) && log 'info' '## grep_is_sufficient: system grep accepted' \
  || error_and_proceed 'system grep rejected; tfenv needs only `\+` in a BRE and `-o` multi-match';

log 'info' '## grep_is_sufficient: rejects a grep whose -o returns only the first match';
(
  # Satisfies the BRE probe but truncates -o output, so only the multi-match check rejects it
  declare stub_dir;
  stub_dir="$(mktemp -d)" || exit 1;
  trap 'rm -rf "${stub_dir}"' EXIT;
  cat > "${stub_dir}/grep" <<'STUB';
#!/usr/bin/env bash
if [ "${1}" = '-o' ]; then
  /usr/bin/grep "${@}" | head -n 1;
else
  /usr/bin/grep "${@}";
fi;
STUB
  chmod +x "${stub_dir}/grep" || exit 1;
  grep_is_sufficient "${stub_dir}/grep" && exit 1;
  exit 0;
) && log 'info' '## grep_is_sufficient: truncating grep rejected' \
  || error_and_proceed 'grep_is_sufficient accepted a grep whose -o returns only the first match';

log 'info' '## grep_is_sufficient: rejects a grep that cannot be run';
(
  grep_is_sufficient /nonexistent/grep && exit 1;
  exit 0;
) && log 'info' '## grep_is_sufficient: missing binary rejected' \
  || error_and_proceed 'grep_is_sufficient accepted a grep binary that does not exist';

log 'info' '## check_dependencies: succeeds with a capable grep on PATH';
(
  check_dependencies || exit 1;
  exit 0;
) && log 'info' '## check_dependencies: passed' \
  || error_and_proceed 'check_dependencies failed despite a capable grep being on PATH';

finish_tests 'dependencies';

exit 0;
