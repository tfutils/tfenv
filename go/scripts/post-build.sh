#!/bin/sh
# Create the terraform entry point alongside tfenv after a manual local build.
#
# goreleaser uses a dual-build approach (see .goreleaser.yml) so this script
# is only needed for manual `go build` workflows outside goreleaser.
#
# The terraform binary is a copy of tfenv — multi-call dispatch uses the
# invoked basename to decide whether to act as tfenv CLI or terraform shim.

set -eu

src="${1:-tfenv}"
dir="$(dirname "${src}")"

if [ -f "${src}" ]; then
  cp "${src}" "${dir}/terraform"
  echo "Created ${dir}/terraform"
elif [ -f "${src}.exe" ]; then
  cp "${src}.exe" "${dir}/terraform.exe"
  echo "Created ${dir}/terraform.exe"
else
  echo "Error: ${src} not found. Build tfenv first:" >&2
  echo "  go build -o tfenv ./cmd/tfenv" >&2
  exit 1
fi
