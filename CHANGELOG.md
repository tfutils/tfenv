## 0.4.4 (July 14, 2017)

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
