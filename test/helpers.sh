#!/usr/bin/env bash

check_version() {
  v=${1}
  [ -n "$(terraform --version | grep "Terraform v${v}")" ]
}

cleanup() {
  rm -rf ./versions
  rm -rf ./.terraform-version
}
