## 2.0.0 (Unreleased)

 * New logging and debugging library
 * Massive testing, logging and loading refactoring
 * Fix to 'use' logic: don't overwrite .terraform-version files
 * Fix #167 - Never invoke use automatically on install - multiple code and testing changes for new logic
 * Fix to not use 0.12.22 during testing which reports its version incorrectly
 * Introduce tfenv-resolve-version to deduplicate translation of requested version into actual version
 * README.md updates

## 1.0.2 (October 29, 2019)

 * Fix failing 0.11.15-oci version, Add additional tests for 0.11.15-oci and alphas, betas and rcs #145
 * Fix to README section link for .terraform-version file #146

## 1.0.1 (June 22, 2019)

 * Fix '--version' flag to return version from CHANGELOG.md when not running from a git checkout.

## 1.0.0 (June 9, 2019)

 * A number of bugfixes and basic script improvements
 * latest keyword doesn't include unstable releases unless specified by regex
 * Support GnuPG tools for signature verification #109
 * Add support for Cygwin #81

## 0.6.0 (November 1, 2017)

 * Support msys2 bash.exe #75
 * Version from sources #73
 * run tfenv as a neighbour with full path (to keep vscode and whoever doesn't respect %path, happy) #72
 * Add current version in list command #69
 * Version file #67
 * [DOC] Add link to puppet module for automatic tfenv installation #64

## 0.5.2 (July 23, 2017)

 * Switch TLS option for curl depending on the version of macOS

## 0.5.1 (July 21, 2017)

 * Fix version detection
 * Add support for ARM architecture

## 0.5.0 (=0.4.4) (July 14, 2017)

 * Immediately switch version after installation
 * Add uninstall command

## 0.4.3 (April 12, 2017)

 * Move temporary directory from /tmp to mktemp
 * Upgrade tfenv-install logging
 * Prevent interactive prompting from keybase

## 0.4.2 (April 9, 2017)

 * Add support for verifying downloads of Terraform

## 0.4.1 (March 8, 2017)

 * Update error_and_die functions to better report their source location
 * libexec/tfenv-version-name: Respect `latest` & `latest:<regex>` syntax in .terraform-version
 * Extension and development of test suite standards

## 0.4.0 (March 6, 2017)

 * Add capability to specify `latest` or `latest:<regex>` in the `use` and `install` actions.
 * Add error_and_die functions to standardise error output
 * Specify --tlsv1.2 to curl requests to remote repo. TLSv1.2 required; supported by but not auto-selected by older NSS.
