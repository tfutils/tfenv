# Agent Guidelines for tfutils/tfenv

Ground rules for AI coding agents contributing to this repository.

## Project Overview

tfenv is a Terraform version manager written in Bash, modelled after rbenv.
~2.5k LOC across `bin/`, `lib/`, `libexec/`, `test/`, and `share/`.

- **Language:** Bash (no shellcheck — deliberate choice)
- **Default branch:** `master`
- **Last release:** v3.0.0 (July 2022) — there is a large unreleased backlog
- **Maintainer:** Mike Peachey (Jaz)

## Repository Structure

```
bin/            Entry points (terraform shim, tfenv command)
lib/            Shared libraries sourced by multiple scripts
libexec/        Subcommands (tfenv-install, tfenv-use, etc.)
test/           Integration tests (download real Terraform binaries)
share/          Static assets (HashiCorp PGP keys)
.github/        CI workflows
```

### Design Principle: Standalone Execution

Every `libexec/` script contains its own boilerplate for resolving
`TFENV_ROOT` and sourcing helpers. This is **intentional** — each script
must be executable in isolation for independent testing. Do not refactor
this into a shared loader.

## Branch and PR Workflow

1. **Never commit directly to `master`.** Always use a feature branch.
2. **Branch naming:** `fix/<short-description>` for bugs,
   `feat/<short-description>` for features, `chore/<short-description>`
   for maintenance. Reference the issue number where applicable
   (e.g. `fix/406-unbound-requested-variable`).
3. **One logical change per branch/PR.** Do not bundle unrelated fixes.
4. **PRs target `master`.** There is no develop or staging branch.
5. **Merge strategy:** The repo uses merge commits (not squash).
   Do not force-push or rebase shared branches.

## Commit Messages

Freeform — no conventional commits standard is enforced. Be descriptive.
Reference issue numbers where applicable (e.g. `Fix #406: ...`).

Historical examples:
- `Fix realpath not available on macOS`
- `Cope with different line endings in .terraform-version`
- `Reduce duplication, and add safety`

## Testing

### What the tests are

Integration tests that **download and install real Terraform binaries**
from HashiCorp. They are not fast, they are not unit tests, and they
require network access.

### How to run tests

```bash
# Install dependencies (macOS only — installs ggrep via Homebrew)
./test/install_deps.sh

# Run all test suites
./test/run.sh

# Run a specific test suite
./test/run.sh test_install_and_use.sh
```

### Test framework

- `test/test_common.sh` provides setup and helpers
- `test/run.sh` discovers and runs all `test_*` files
- Helper functions: `error_and_proceed` (log failure, continue),
  `check_active_version`, `check_installed_version`,
  `check_default_version`, `cleanup`
- **Do not use `error_and_die`** — it does not exist in the test context

### Known issue: test runner always exits 0

`test/run.sh` ends with `exit 0` regardless of test failures. It logs
failures but the exit code does not reflect them. This means CI may not
catch test regressions. Be aware and check test output manually.

### CI matrix

| Trigger | Platforms |
|---------|-----------|
| Pull request | `macos-latest`, `ubuntu-latest` |
| Push to master | `ubuntu-24.04`, `ubuntu-22.04`, `macos-14`, `macos-13`, `windows-2025`, `windows-2022` |

CI also builds the Dockerfile and validates it on Ubuntu runners.

## Before Submitting a PR

1. **Run the test suite locally** and verify all tests pass. Do not rely
   solely on CI — the exit code bug means CI may report green on failure.
2. **Read the diff carefully.** Bash quoting and operator precedence bugs
   are the most common class of defect in this codebase.
3. **Test on at least one platform.** If you only have Linux, note that
   in the PR. macOS and Windows have different `readlink`, `grep`, and
   `sed` behaviours.
4. **Do not modify the test runner exit code bug** as a drive-by fix in
   an unrelated PR. It should be its own tracked change.

## Release Process

Releases are manual and infrequent. The historical process is:

1. Accumulate changes on `master`
2. Update `CHANGELOG.md` with a new version section using the format:
   ```
   ## X.Y.Z (Month Day, Year)

    * CATEGORY: Description (Contributor Name <email>)
   ```
   Categories: `BREAKING CHANGE`, `NEW FEATURE`, `FIX`, `MAJOR THANKS`
3. Create an annotated tag: `git tag -a vX.Y.Z -m "tfenv vX.Y.Z"`
4. Push the tag: `git push origin vX.Y.Z`
5. Update `Dockerfile` if it hardcodes the version

**Agents must not create releases or tags without explicit instruction
from the maintainer.**

## Things Agents Must Not Do

- **Do not run `git push` to `master` or any remote branch** without
  explicit approval from Jaz.
- **Do not create GitHub releases or tags.**
- **Do not close or lock issues** — only the maintainer triages.
- **Do not refactor the standalone boilerplate** in `libexec/` scripts.
- **Do not add shellcheck** directives, configs, or CI steps.
- **Do not add new dependencies** (external tools, packages) without
  discussion. The project deliberately has minimal dependencies
  (bash, curl, grep/ggrep, unzip).
- **Do not modify `.github/workflows/`** without explicit approval —
  CI changes affect all platforms and have outsized blast radius.

## Common Pitfalls in This Codebase

These are the most frequent sources of bugs. Check for them in every change:

1. **Shell operator precedence:** `cmd1 || cmd2 | cmd3` means
   `cmd1 || (cmd2 | cmd3)`, not `(cmd1 || cmd2) | cmd3`.
2. **Unquoted variables:** Always quote `"${var}"` in arguments,
   conditions, and assignments. Unquoted variables cause word-splitting.
3. **Double-quoted traps:** `trap "rm ${var}" EXIT` expands at definition
   time and is vulnerable to word-splitting. Use functions or single quotes.
4. **`$@` in for-loops:** Always quote: `for arg in "$@"`.
5. **Regex anchoring:** `^1.1` matches `1.10.x` because `.` is a regex
   wildcard. Use `^1\.1\.` for exact prefix matching.
6. **Cross-platform differences:** macOS uses BSD `sed`/`grep`/`readlink`.
   GNU extensions may not be available. The project installs `ggrep` on
   macOS but uses the system `sed` and `readlink`.
7. **`set -uo pipefail`:** All scripts run with strict mode. Undefined
   variables cause immediate crashes. Use `${var:-default}` for optional
   variables.
8. **Carriage returns:** `.terraform-version` files from Windows/WSL may
   contain `\r`. Always strip with `tr -d '\r'` after reading.
