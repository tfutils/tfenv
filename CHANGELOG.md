## 3.2.0 (April 24, 2026)

 * NEW FEATURE: Support comments in `.terraform-version` files — lines starting with `#` and inline `#` comments are stripped, fixing #391 and #283 (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Support uninstalling multiple versions in a single `tfenv uninstall` invocation, fixing #449 (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Show binary architecture in `tfenv list` output (e.g. `amd64`, `arm64`), fixing #321 (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Support `tfenv use -` to switch to the previously active version, fixing #378 (Mike Peachey <mike.peachey@bjss.com>)

## 3.1.0 (April 24, 2026)

 * NEW FEATURE: Add `latest-allowed` version resolution from `required_version` constraint in Terraform files, supporting `=`, `>=`, `>`, `<=`, and `~>` operators (Oliver Ford <dev@ojford.com>, Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Add `TFENV_SKIP_REMOTE_CHECK` to skip checking remote versions before installing (David Forster <david.forster@capgemini.com>)
 * NEW FEATURE: Add per-version file locking to prevent concurrent install races, fixing #422 (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Load bashlog conditionally based on shell interactivity via `TFENV_BASHLOG`, fixing #229 (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Improve custom remote compatibility for Artifactory and similar mirrors (Mike Peachey <mike.peachey@bjss.com>)
 * NEW FEATURE: Add Dockerfile for containerised usage and end-to-end testing (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Filter pre-release versions from `latest` resolution; use `latest:` (with colon) or alphabetic regex to include pre-releases, fixing #375 and #204 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: CLI argument takes precedence over `TFENV_TERRAFORM_VERSION` for install, fixing #441 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Route tfenv log output to stderr in terraform shim context, fixing #374 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Do not assume brew is available for GNU grep detection, fixing #442 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Operator precedence bug and unbound variable in version-name resolution, fixing #406 and #431 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Double-quoted trap in tfenv-install vulnerable to word-splitting, fixing #455 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Unquoted curl options in curlw() causing failures with paths containing spaces, fixing #454 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Unquoted argument iteration in tfenv-exec causing word-splitting, fixing #453 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Stray closing paren in syslog tag default in bashlog, fixing #451 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Remove dead code and fix indentation in install script, fixing #452 and #460 (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix `realpath` not available on macOS (Oliver Ford <dev@ojford.com>)
 * FIX: Fix use of `-chdir` with an absolute path, fixing #354 (Oliver Ford <dev@ojford.com>)
 * FIX: Cope with different line endings in `.terraform-version` files (Adam Christie <fractos@gmail.com>)
 * FIX: Fix Windows OS selection with underscore handling for Cygwin-like tools (Jack Blower <Jack@elvenspellmaker.co.uk>)
 * FIX: Fix regex for darwin arm64 to amd64 workaround (Bryan Hiestand <bryanhiestand@users.noreply.github.com>)
 * FIX: Fix macOS architecture selection for TFENV_ARCH (Deniz Genç <25902330+denizgenc@users.noreply.github.com>)
 * FIX: Replace use of `rev` for Windows compatibility (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix latest-allowed version matching for partial versions (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix test runner to exit with correct status codes reflecting actual failures (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix test infrastructure bugs including undefined functions and missing quotes (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Revert premature optimisation that skipped release list checks (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix GitHub Actions matrix.os strategy (TAKANO Mitsuhiro <takano32@gmail.com>)
 * FIX: Update README.md for zsh users (EC2 Default User <kuredev@users.noreply.github.com>)
 * FIX: Add fish shell PATH example to README.md (sato-s <s.sato.desu@gmail.com>)
 * FIX: Add zprofile example to README.md (Kasper Christensen <kasper@friischristensen.com>)
 * FIX: Update terraform docs link in README.md (Bob Idle <102661087+bobidle@users.noreply.github.com>)
 * FIX: Update CI badge in README.md (David Beitey <david@davidjb.com>)
 * FIX: Correct TFENV_ARCH defaults in README.md (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Update Windows support status in documentation (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Fix legacy `ENV` syntax in Dockerfile to use `=` format (Mike Peachey <mike.peachey@bjss.com>)
 * TEST: Complete test coverage overhaul — 7 new test suites (commands, list-remote, pin, version-file, pre-release filtering, env vars, resolve-version), infrastructure improvements, and isolation enhancements (Mike Peachey <mike.peachey@bjss.com>)
 * MAJOR THANKS: Oliver Ford for the `latest-allowed` feature and macOS fixes
 * MAJOR THANKS: All community contributors for bug reports, fixes, and patience during the long gap between releases

## 3.0.0 (July 15, 2022)

 * BREAKING CHANGE: Don't install ggrep on macs automatically; it's just not good practice. ggrep now a pre-requisite for Mac.
 * BREAKING CHANGE: When TFENV_AUTO_INSTALL=true, tfenv use will now attempt to install a matching version when no matching version is installed already, including defaulting to latest if no version is specified at the command line or in overrides (Mike Peachey <mike.peachey@bjss.com>)
 * MAJOR THANKS: Extreme performance improvements for bashlog (Zoltán Reegn <zoltan.reegn@gmail.com> & Andrew Bradley <abradley@brightcove.com>)
 * NEW FEATURE: Add support for netrc for private mirrors (Jack Parsons <jack.parsons@test-and-trace.nhs.uk>)
 * NEW FEATURE: Support reversed list of versions from remote (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Try to fix RC sorting order in use. Do not invoke version on use as it's meaningless and confusing when defaults are overridden (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Improve mktemp error message (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: #282 - Update warning message when no PGP in use (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: #299 - When AUTO_INSTALL, tfenv use will attempt to install, including default latest (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: #315 - ugly 'no such file or directory' message when setting terraform version for the first time (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: #318 - Updated documentation regarding files instructing PGP usage (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Add test of auto-installing min-required version (Oliver Ford <dev@ojford.com>)
 * FIX: #300: `min-required` not evaluated for auto-install (Oliver Ford <dev@ojford.com>)
 * FIX: #340 - Unbound variable (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Install ggrep during mac tests (Mike Peachey <mike.peachey@bjss.com>)
 * FIX: Update README.md for WSL users (Panu Valtanen <p4nu@users.noreply.github.com>)
 * FIX: Look up correct versions location in tfenv-pin and set tfenv-pin as executable (Aaron Madlon-Kay <aaron@madlon-kay.com>)
 * FIX: Handle new alpha versioning syntax (Dylan Turner <dylan.turner@bjss.com>)
 * FIX: Add missing exports (Andrew Bradley <abradley@brightcove.com>)
 * FIX: Extract some libexec commands into helper functions (Andrew Bradley <abradley@brightcove.com>)
 * FIX: ARM/ARM64 support (Volodymyr Samusia <samusia.vs@gmail.com>)
 * FIX: Docs clarification for arm achitecture (Kuba <jakub.glapa@adverity.com>)
 * FIX: Use macos-11 tests (Stephen <stephengroat@users.noreply.github.com>)
 * FIX: README.md - Suggest shallow git clone for regular usage (Aurélien Joga <aurelienjoga@gmail.com>)
 * FIX: Support redirects on downloads (Troy Ready <troy@troyready.com>)
 * FIX: README.md - Support Apple Silicon (Greg Dubicki <grzegorz.dubicki@gmail.com>)

## 2.2.3 (Feb 9, 2022)

 * Fix: mktemp not working correctly on Alpine Linux (#285)
 * Add support of ARM64 (#280)
 * Add support for tf.json files on min-required (#277)
 * Fix issue #210 - allow non-numeric values for DEBUG (#274)
 * Download latest version if user uses regex and TFENV_AUTO_INSTALL is true (#272)
 * Add tfenv pin command (#270)

## 2.2.2 (May 6, 2021)

 * remove trust from revoked signing key as of hcsec-2021-12
 * fix installation of versions signed by revoked key by forcing to use the new key

## 2.2.1 (April 29, 2021)

 * hcsec-2021-12 (#257)

## 2.2.0 (February 06, 2021)

 * Convert GitHub CI from Travis CI to Github Actions
 * Fix min-required after it was broken by 2.1.0 (#235)
 * Min-required recursive lookup was dangerously broken. Removed the recursion that should never have been (#237)
 * Fix the failure of tfenv list when no default was set (#236)
 * Add init command (#240)
 * Use ggrep on Mac with Homebrew (#218)

## 2.1.0 (January 30, 2021)

 * Update tfenv-min-required to search root before recursing (#203)  
 * Terraform 0.13.0 support (#191)
 * Add Arch Linux install instructions via Arch User Repository (AUR) (#201)
 * min-required correctly finds tagged release versions (#206)
 * install: make keybase a fall-through verification variant (#213)
 * Feature/add TFENV_TERRAFORM_VERSION env var (#222)
 * Document version-name command (#224)
 * Fix signature verification bypass due to insufficient hashsum checking (#212)
 * Fix keybase login exit code handling (#188)
 * Fix bug on MacOS when using CLICOLOR=1 (#152)
 * Improved error handling in tfenv-list-remote-curl (#186)
 * Test in Windows (#140)
 * force tfenv to write over existing zip if it exists (#169)
 * Remove the versions directory when the last version is uninstalled (#128)
 * Add support for sha256sum command (#170)
 * Adding freebsd support (#133)
 * Improve shell script synatx (#174)
 * Begrudging Bash 3.x Compatability because of macOS (#181)

## 2.0.0 (April 20, 2020)

 * New logging and debugging library
 * Massive testing, logging and loading refactoring
 * Fix to 'use' logic: don't overwrite .terraform-version files
 * Fix #167 - Never invoke use automatically on install - multiple code and testing changes for new logic
 * Fix to not use 0.12.22 during testing which reports its version incorrectly
 * Introduce tfenv-resolve-version to deduplicate translation of requested version into actual version
 * README.md updates
 * Fix #176 - New parameter TFENV_AUTO_INSTALL to handle the version specified by `use` or a `.terraform-version` file not being installed

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
