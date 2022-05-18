![CI](https://github.com/tfutils/tfenv/workflows/CI/badge.svg)

# tfenv

[Terraform](https://www.terraform.io/) version manager inspired by [rbenv](https://github.com/rbenv/rbenv)

## Support

Currently tfenv supports the following OSes

- Mac OS X (64bit)
- Linux
  - 64bit
  - Arm
- Windows (64bit) - only tested in git-bash - currently presumed failing due to symlink issues in git-bash

## Installation

### Automatic

Install via Homebrew

```console
$ brew install tfenv
```

Install via Arch User Repository (AUR)
   
```console
$ yay --sync tfenv
```

Install via puppet

Using puppet module [sergk-tfenv](https://github.com/SergK/puppet-tfenv)

```puppet
include ::tfenv
```

### Manual

1. Check out tfenv into any path (here is `${HOME}/.tfenv`)

```console
$ git clone https://github.com/tfutils/tfenv.git ~/.tfenv
```

2. Add `~/.tfenv/bin` to your `$PATH` any way you like

```console
$ echo 'export PATH="$HOME/.tfenv/bin:$PATH"' >> ~/.bash_profile
```

  OR you can make symlinks for `tfenv/bin/*` scripts into a path that is already added to your `$PATH` (e.g. `/usr/local/bin`) `OSX/Linux Only!`

```console
$ ln -s ~/.tfenv/bin/* /usr/local/bin
```

  On Ubuntu/Debian touching `/usr/local/bin` might require sudo access, but you can create `${HOME}/bin` or `${HOME}/.local/bin` and on next login it will get added to the session `$PATH`
  or by running `. ${HOME}/.profile` it will get added to the current shell session's `$PATH`.

```console
$ mkdir -p ~/.local/bin/
$ . ~/.profile
$ ln -s ~/.tfenv/bin/* ~/.local/bin
$ which tfenv
```

## Usage

### tfenv install [version]

Install a specific version of Terraform.

If no parameter is passed, the version to use is resolved automatically via [TFENV\_TERRAFORM\_VERSION environment variable](#tfenv_terraform_version), [.terraform-version files](#terraform-version-file), or [required_version in "terraform" section of any .tf or .tf.json file](#min-required), in that order of precedence, i.e. TFENV\_TERRAFORM\_VERSION, then .terraform-version, and then required_version in .tf. The default is 'latest' if none are found.

If a parameter is passed, available options:

- `x.y.z` [Semver 2.0.0](https://semver.org/) string specifying the exact version to install
- `latest` is a syntax to install latest version
- `latest:<regex>` is a syntax to install latest version matching regex (used by grep -e)
- `min-required` is a syntax to recursively scan your Terraform files to detect which version is minimally required. See [required_version](https://www.terraform.io/docs/configuration/terraform.html) docs. Also [see min-required](#min-required) section below.

```console
$ tfenv install
$ tfenv install 0.7.0
$ tfenv install latest
$ tfenv install latest:^0.8
$ tfenv install min-required
```

If `shasum` is present in the path, tfenv will verify the download against Hashicorp's published sha256 hash.
If [keybase](https://keybase.io/) is available in the path it will also verify the signature for those published hashes using Hashicorp's published public key.

You can opt-in to using GnuPG tools for PGP signature verification if keybase is not available:

```console
$ echo 'trust-tfenv: yes' > ~/.tfenv/use-gpgv
$ tfenv install
```

The `trust-tfenv` directive means that verification uses a copy of the
Hashicorp OpenPGP key found in the tfenv repository.  Skipping that directive
means that the Hashicorp key must be in the existing default trusted keys.
Use the file `~/.tfenv/use-gnupg` to instead invoke the full `gpg` tool and
see web-of-trust status; beware that a lack of trust path will not cause a
validation failure.

#### .terraform-version

If you use a [.terraform-version](#terraform-version-file) file, `tfenv install` (no argument) will install the version written in it.

#### min-required

Please note that we don't do semantic version range parsing but use first ever found version as the candidate for minimally required one. It is up to the user to keep the definition reasonable. I.e.

```terraform
// this will detect 0.12.3
terraform {
  required_version  = "<0.12.3, >= 0.10.0"
}
```

```terraform
// this will detect 0.10.0
terraform {
  required_version  = ">= 0.10.0, <0.12.3"
}
```

### Environment Variables

#### TFENV

##### `TFENV_ARCH`

String (Default: amd64)

Specify architecture. Architecture other than the default amd64 can be specified with the `TFENV_ARCH` environment variable

```console
$ TFENV_ARCH=arm tfenv install 0.7.9
```

##### `TFENV_AUTO_INSTALL`

String (Default: true)

Should tfenv automatically install terraform if the version specified by defaults or a .terraform-version file is not currently installed.

```console
$ TFENV_AUTO_INSTALL=false terraform plan
```

##### `TFENV_CURL_OUTPUT`

Integer (Default: 2)

Set the mechanism used for displaying download progress when downloading terraform versions from the remote server.

* 2: v1 Behaviour: Pass `-#` to curl
* 1: Use curl default
* 0: Pass `-s` to curl

##### `TFENV_DEBUG`

Integer (Default: 0)

Set the debug level for TFENV.

* 0: No debug output
* 1: Simple debug output
* 2: Extended debug output, with source file names and interactive debug shells on error
* 3: Debug level 2 + Bash execution tracing

##### `TFENV_REMOTE`

String (Default: https://releases.hashicorp.com)

To install from a remote other than the default

```console
$ TFENV_REMOTE=https://example.jfrog.io/artifactory/hashicorp
```

##### `TFENV_CONFIG_DIR`

Path (Default: `$TFENV_ROOT`)

The path to a directory where the local terraform versions and configuration files exist.

```console
TFENV_CONFIG_DIR="$XDG_CONFIG_HOME/tfenv"
```

##### `TFENV_TERRAFORM_VERSION`

String (Default: "")

If not empty string, this variable overrides Terraform version, specified in [.terraform-version files](#terraform-version-file).
`latest` and `latest:<regex>` syntax are also supported.
[`tfenv install`](#tfenv-install-version) and [`tfenv use`](#tfenv-use-version) command also respects this variable.

e.g.

```console
$ TFENV_TERRAFORM_VERSION=latest:^0.11. terraform --version
```

#### Bashlog Logging Library

##### `BASHLOG_COLOURS`

Integer (Default: 1)

To disable colouring of console output, set to 0.


##### `BASHLOG_DATE_FORMAT`

String (Default: +%F %T)

The display format for the date as passed to the `date` binary to generate a datestamp used as a prefix to:

* `FILE` type log file lines.
* Each console output line when `BASHLOG_EXTRA=1`

##### `BASHLOG_EXTRA`

Integer (Default: 0)

By default, console output from tfenv does not print a date stamp or log severity.

To enable this functionality, making normal output equivalent to FILE log output, set to 1.

##### `BASHLOG_FILE`

Integer (Default: 0)

Set to 1 to enable plain text logging to file (FILE type logging).

The default path for log files is defined by /tmp/$(basename $0).log
Each executable logs to its own file.

e.g.

```console
$ BASHLOG_FILE=1 tfenv use latest
```

will log to `/tmp/tfenv-use.log`

##### `BASHLOG_FILE_PATH`

String (Default: /tmp/$(basename ${0}).log)

To specify a single file as the target for all FILE type logging regardless of the executing script.

##### `BASHLOG_I_PROMISE_TO_BE_CAREFUL_CUSTOM_EVAL_PREFIX`

String (Default: "")

*BE CAREFUL - MISUSE WILL DESTROY EVERYTHING YOU EVER LOVED*

This variable allows you to pass a string containing a command that will be executed using `eval` in order to produce a prefix to each console output line, and each FILE type log entry.

e.g.

```console
$ BASHLOG_I_PROMISE_TO_BE_CAREFUL_CUSTOM_EVAL_PREFIX='echo "${$$} "'
```
will prefix every log line with the calling process' PID.

##### `BASHLOG_JSON`

Integer (Default: 0)

Set to 1 to enable JSON logging to file (JSON type logging).

The default path for log files is defined by /tmp/$(basename $0).log.json
Each executable logs to its own file.

e.g.

```console
$ BASHLOG_JSON=1 tfenv use latest
```

will log in JSON format to `/tmp/tfenv-use.log.json`

JSON log content:

`{"timestamp":"<date +%s>","level":"<log-level>","message":"<log-content>"}`

##### `BASHLOG_JSON_PATH`

String (Default: /tmp/$(basename ${0}).log.json)

To specify a single file as the target for all JSON type logging regardless of the executing script.

##### `BASHLOG_SYSLOG`

Integer (Default: 0)

To log to syslog using the `logger` binary, set this to 1.

The basic functionality is thus:

```console
$ local tag="${BASHLOG_SYSLOG_TAG:-$(basename "${0}")}";
$ local facility="${BASHLOG_SYSLOG_FACILITY:-local0}";
$ local pid="${$}";
$ logger --id="${pid}" -t "${tag}" -p "${facility}.${severity}" "${syslog_line}"
```

##### `BASHLOG_SYSLOG_FACILITY`

String (Default: local0)

The syslog facility to specify when using SYSLOG type logging.

##### `BASHLOG_SYSLOG_TAG`

String (Default: $(basename $0))

The syslog tag to specify when using SYSLOG type logging.

Defaults to the PID of the calling process.



### tfenv use [version]

Switch a version to use

If no parameter is passed, the version to use is resolved automatically via [.terraform-version files](#terraform-version-file) or [TFENV\_TERRAFORM\_VERSION environment variable](#tfenv_terraform_version) (TFENV\_TERRAFORM\_VERSION takes precedence), defaulting to 'latest' if none are found.

`latest` is a syntax to use the latest installed version

`latest:<regex>` is a syntax to use latest installed version matching regex (used by grep -e)

`min-required` will switch to the version minimally required by your terraform sources (see above `tfenv install`)

```console
$ tfenv use
$ tfenv use min-required
$ tfenv use 0.7.0
$ tfenv use latest
$ tfenv use latest:^0.8
```

### tfenv uninstall &lt;version>

Uninstall a specific version of Terraform
`latest` is a syntax to uninstall latest version
`latest:<regex>` is a syntax to uninstall latest version matching regex (used by grep -e)

```console
$ tfenv uninstall 0.7.0
$ tfenv uninstall latest
$ tfenv uninstall latest:^0.8
```

### tfenv list

List installed versions

```console
$ tfenv list
* 0.10.7 (set by /opt/tfenv/version)
  0.9.0-beta2
  0.8.8
  0.8.4
  0.7.0
  0.7.0-rc4
  0.6.16
  0.6.2
  0.6.1
```

### tfenv list-remote

List installable versions

```console
$ tfenv list-remote
0.9.0-beta2
0.9.0-beta1
0.8.8
0.8.7
0.8.6
0.8.5
0.8.4
0.8.3
0.8.2
0.8.1
0.8.0
0.8.0-rc3
0.8.0-rc2
0.8.0-rc1
0.8.0-beta2
0.8.0-beta1
0.7.13
0.7.12
...
```

## .terraform-version file

If you put a `.terraform-version` file on your project root, or in your home directory, tfenv detects it and uses the version written in it. If the version is `latest` or `latest:<regex>`, the latest matching version currently installed will be selected.

Note, that [TFENV\_TERRAFORM\_VERSION environment variable](#tfenv_terraform_version) can be used to override version, specified by `.terraform-version` file.

```console
$ cat .terraform-version
0.6.16

$ terraform version
Terraform v0.6.16

Your version of Terraform is out of date! The latest version
is 0.7.3. You can update by downloading from www.terraform.io

$ echo 0.7.3 > .terraform-version

$ terraform version
Terraform v0.7.3

$ echo latest:^0.8 > .terraform-version

$ terraform version
Terraform v0.8.8

$ TFENV_TERRAFORM_VERSION=0.7.3 terraform --version
Terraform v0.7.3
```

## Upgrading

```console
$ git --git-dir=~/.tfenv/.git pull
```

## Uninstalling

```console
$ rm -rf /some/path/to/tfenv
```

## LICENSE

- [tfenv itself](https://github.com/tfutils/tfenv/blob/master/LICENSE)
- [rbenv](https://github.com/rbenv/rbenv/blob/master/LICENSE)
  - tfenv partially uses rbenv's source code
