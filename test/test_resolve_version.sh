#!/usr/bin/env bash

# Source common test setup
source "$(dirname "${0}")/test_common.sh";

#####################
# Begin Script Body #
#####################

declare -a errors=();

log 'info' '### Test Suite: resolve-version (latest-allowed operators)';

log 'info' '## latest-allowed: bare version number (exact pin)';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = "1.6.1"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## latest-allowed: bare version number passed' \
  || error_and_proceed 'latest-allowed did not resolve bare version number';

log 'info' '## latest-allowed: = prefix (exact pin)';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = "= 1.6.1"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## latest-allowed: = prefix passed' \
  || error_and_proceed 'latest-allowed did not resolve = prefix version';

log 'info' '## latest-allowed: >= operator (resolves to latest)';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = ">= 1.0.0"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  # >= should resolve to absolute latest stable
  declare latest;
  latest="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1)";
  check_installed_version "${latest}" || exit 1;
) && log 'info' '## latest-allowed: >= operator passed' \
  || error_and_proceed 'latest-allowed >= did not resolve to latest stable';

log 'info' '## latest-allowed: > operator (resolves to latest)';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = "> 0.12.0"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  declare latest;
  latest="$(tfenv list-remote | grep -e "^[0-9]\+\.[0-9]\+\.[0-9]\+$" | head -n 1)";
  check_installed_version "${latest}" || exit 1;
) && log 'info' '## latest-allowed: > operator passed' \
  || error_and_proceed 'latest-allowed > did not resolve to latest stable';

log 'info' '## latest-allowed: <= operator (exact version)';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = "<= 1.6.1"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## latest-allowed: <= operator passed' \
  || error_and_proceed 'latest-allowed <= did not resolve to exact version';

log 'info' '## latest-allowed: ~> pessimistic constraint';
cleanup || log 'error' 'Cleanup failed?!';
(
  mkdir -p tf_dir;
  cat > tf_dir/main.tf <<'EOF'
terraform {
  required_version = "~> 1.6.0"
}
EOF
  echo 'latest-allowed' > tf_dir/.terraform-version;
  cd tf_dir || exit 1;
  tfenv install || exit 1;
  # ~> 1.6.0 should install latest 1.6.x
  declare installed;
  installed="$(tfenv list | head -n 1 | sed 's/^[* ]*//')";
  echo "${installed}" | grep -qE '^1\.6\.' || exit 1;
) && log 'info' '## latest-allowed: ~> pessimistic constraint passed' \
  || error_and_proceed 'latest-allowed ~> did not resolve correctly';

log 'info' '## TFENV_TERRAFORM_VERSION overrides version file';
cleanup || log 'error' 'Cleanup failed?!';
(
  echo '0.12.0' > .terraform-version;
  TFENV_TERRAFORM_VERSION=1.6.1 tfenv install || exit 1;
  check_installed_version 1.6.1 || exit 1;
) && log 'info' '## TFENV_TERRAFORM_VERSION override: passed' \
  || error_and_proceed 'TFENV_TERRAFORM_VERSION did not override version file';

finish_tests 'resolve_version';

exit 0;
