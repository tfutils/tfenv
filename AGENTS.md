# Agent Guidelines for tfutils/tfenv

Master context document for AI coding agents contributing to this repository.
Every agent MUST read this file before starting work.

---

## Project Overview

tfenv is a Terraform version manager written in Bash, modelled after rbenv.
~2.5k LOC across `bin/`, `lib/`, `libexec/`, `test/`, and `share/`.

- **Language:** Bash (no shellcheck — deliberate choice; see
  [ADR-0003](docs/adr/0003-no-shellcheck.md))
- **Default branch:** `master`
- **Current release:** v3.2.0 (April 2026)
- **Maintainer:** Mike Peachey / Jaz (@Zordrak)
- **Minimal dependencies:** bash, curl, grep/ggrep, sort, unzip
- **License:** MIT

## Repository Structure

```
bin/              Entry points (terraform shim, tfenv command)
lib/              Shared libraries sourced by multiple scripts
libexec/          Subcommands (tfenv-install, tfenv-use, etc.)
test/             Integration tests (download real Terraform binaries)
share/            Static assets (HashiCorp PGP keys)
.github/          CI workflows, agent definitions, issue templates
  agents/         Agent definition files (.agent.md)
  instructions/   Auto-loaded coding standards
  ISSUE_TEMPLATE/ Structured issue forms
docs/
  adr/            Architecture Decision Records
```

---

## Core Principles

### 1. Standalone Script Execution

Every `libexec/` script contains its own boilerplate for resolving
`TFENV_ROOT` and sourcing helpers. This is **intentional** — each script
must be executable in isolation. Do NOT refactor this into a shared loader.
See [ADR-0002](docs/adr/0002-standalone-script-execution.md).

### 2. Worktree-First

ALL code changes MUST happen in a git worktree under `.worktrees/`. The main
working tree is shared by all agents and the human — modifying it directly
risks conflicts. `.worktrees/` is gitignored.

```bash
# Create a worktree for your work
git worktree add .worktrees/fix-123 -b fix/123-description master

# Do all work inside the worktree
cd .worktrees/fix-123

# Clean up when done (after PR is created)
cd /path/to/main/tree
git worktree remove .worktrees/fix-123
```

**Exception:** Read-only operations (searching, reading files, running tests)
may use the main working tree.

### 3. Quality Over Speed

- Run the test suite before committing
- Read the diff before pushing
- Check your own work before calling it done
- Fix problems when you find them — do not defer

### 4. Minimal Dependencies

The project deliberately has minimal dependencies: bash, curl, grep/ggrep,
sort, and unzip. Do NOT add new external tools or packages without explicit
approval from the maintainer.

### 5. No ShellCheck

The project does not use ShellCheck. This is a deliberate architectural
decision. Do NOT add shellcheck directives, configs, or CI steps.
See [ADR-0003](docs/adr/0003-no-shellcheck.md).

---

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

---

## Claim Protocol

Before starting work on any issue, an agent MUST claim it:

1. Add the `agent:in-progress` label to the issue
2. Post a comment: `Claimed by <agent-name>. Working on this.`
3. If the label is already present, the issue is claimed — do NOT work on it

This acts as a distributed lock preventing duplicate effort when multiple
agents run concurrently.

When work is complete (PR created), remove `agent:in-progress` and add
`agent:review-requested`.

```bash
# Claim an issue
gh issue edit NNN --add-label 'agent:in-progress'
gh issue comment NNN --body 'Claimed by <agent-name>. Working on this.'

# Release after PR created
gh issue edit NNN --remove-label 'agent:in-progress' --add-label 'agent:review-requested'
```

---

## GitHub CLI Strategy

The `gh` CLI is the **primary and load-bearing** interface to GitHub for all
agents. Use `gh api repos/tfutils/tfenv/...` for operations not covered by
high-level `gh` commands.

The GitHub MCP server may supplement `gh` where proven reliable, but must
never be on the critical path.

---

## Work Type Ownership

Each work type is owned by a specific agent. If you receive a request that
belongs to a different agent, say so and stop.

| Work Type | Owning Agent | Description |
| --------- | ------------ | ----------- |
| Autonomous delivery orchestration | `developer` | Assess board, dispatch specialists |
| Bug hunting and auditing | `bug-finder` | Find defects, file structured issues |
| Bug fixing | `bug-fixer` | Implement fixes with tests |
| Feature design and specification | `feature-designer` | Write detailed specs as issues |
| Feature implementation | `feature-implementer` | Build from specs with tests |
| Code review | `reviewer` | First-pass structured PR review |
| Architecture and design decisions | `architect` | ADRs, decomposition, trade-off analysis |
| Documentation quality | `documenter` | Cross-doc consistency, link integrity |
| Delivery metrics and reporting | `evaluator` | Compute metrics from `gh` data |
| Board management and triage | `pm` | Organise backlog, recommend priorities |
| Release management | `releaser` | CHANGELOG, tags, GitHub releases |

---

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

---

## Release Process

Releases are manual and infrequent. The `releaser` agent handles the
judgement-intensive parts but always requires human confirmation.

1. Accumulate changes on `master`
2. Update `CHANGELOG.md` with a new version section using the format:
   ```
   ## X.Y.Z (Month Day, Year)

    * CATEGORY: Description (Contributor Name <email>)
   ```
   Categories: `BREAKING CHANGE`, `NEW FEATURE`, `FIX`, `MAJOR THANKS`
3. Create an annotated tag: `git tag -a vX.Y.Z -m "tfenv vX.Y.Z"`
4. Push the tag: `git push origin vX.Y.Z`
5. Create a GitHub Release from the tag
6. Update `Dockerfile` if it hardcodes the version

**Agents must not create releases or tags without explicit instruction
from the maintainer.**

---

## Label Taxonomy

Issues use a namespaced label system:

| Namespace | Labels | Purpose |
| --------- | ------ | ------- |
| `type:` | `bug`, `feature`, `chore`, `question` | What kind of work |
| `severity:` | `critical`, `high`, `medium`, `low` | Impact of bugs |
| `priority:` | `critical`, `high`, `medium`, `low` | When to address |
| `complexity:` | `trivial`, `small`, `medium`, `large` | Effort estimate |
| `confidence:` | `confirmed`, `probable`, `speculative` | Bug certainty |
| `category:` | `install`, `use`, `list`, `uninstall`, `version-resolution`, `verification`, `platform`, `logging` | Affected area |
| `agent:` | `in-progress`, `review-requested` | Agent workflow state |

---

## Things Agents Must Not Do

- **Do not run `git push` to `master` or any remote branch** without
  explicit approval from Jaz
- **Do not create GitHub releases or tags** without explicit instruction
- **Do not close or lock issues** — only the maintainer triages
- **Do not refactor the standalone boilerplate** in `libexec/` scripts
- **Do not add shellcheck** directives, configs, or CI steps
- **Do not add new dependencies** without discussion
- **Do not modify `.github/workflows/`** without explicit approval —
  CI changes affect all platforms and have outsized blast radius
- **Do not merge PRs** — create them and leave for human review
- **Do not modify Terraform structure** (if any) without explicit approval
  from Jaz — content changes are fine, structural changes are not
- **Do not write to `/tmp` or `/dev/null`** — use `.tmp/` in the workspace
  root for temporary files
- **Do not remove tool permissions** from agent definitions — granted
  permissions (`tools:` in frontmatter) are the maintainer's decision.
  Agents may recommend changes but must not unilaterally reduce access.

---

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

---

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

---

## Satellite Documentation

| Document | Purpose |
| -------- | ------- |
| [.github/copilot-instructions.md](.github/copilot-instructions.md) | Copilot configuration, auto-loading |
| [.github/instructions/bash.instructions.md](.github/instructions/bash.instructions.md) | Bash coding standards (auto-loaded for `**/*.sh`) |
| [SECURITY.md](SECURITY.md) | Vulnerability reporting policy |
| [docs/adr/](docs/adr/) | Architecture Decision Records |
| [CHANGELOG.md](CHANGELOG.md) | Release history |
