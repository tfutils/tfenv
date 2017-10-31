#!/usr/bin/env bash

check_version() {
  v="${1}"
  [ -n "$(terraform --version | grep -E "^Terraform v${v}(-dev)?$")" ]
}

cleanup() {
  rm -rf ./versions
  rm -rf ./.terraform-version
  rm -rf ./min_required.tf
}
