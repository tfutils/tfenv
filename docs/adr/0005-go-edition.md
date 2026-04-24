# 005. Go Edition — Dual-Implementation Strategy

**Status:** Proposed

**Date:** 2026-04-24

## Context

tfenv has been a Bash-only project since inception. This has served the
community well — Bash is universally available on Unix-like systems and
keeps the project zero-dependency for most users. However, several
persistent pain points stem directly from the Bash implementation:

1. **Cross-platform friction.** macOS requires GNU grep (`ggrep`) via
   Homebrew because BSD grep lacks required features. Windows support is
   best-effort via git-bash with symlink workarounds. FreeBSD, Solaris,
   and other platforms require manual adaptation.

2. **Fragile external tool dependencies.** PGP verification relies on a
   three-tier fallback (keybase → gpgv → gpg), SHA256 verification
   differs between `shasum` and `sha256sum`, TLS handling varies by curl
   version, and archive extraction depends on system `unzip`.

3. **Limited testability.** The test suite is integration-only — every
   test downloads real Terraform binaries over the network. There are no
   unit tests because Bash makes isolation impractical. The test runner
   exits 0 regardless of failures.

4. **Distribution burden.** Users must clone the repo or use a package
   manager that understands the Bash script layout. There is no single
   downloadable artifact.

5. **Version constraint parsing.** Semantic version comparison and HCL
   `required_version` constraint parsing are implemented via regex and
   `sort` pipelines — functional but fragile and hard to extend.

Meanwhile, the Go ecosystem provides battle-tested solutions for every
one of these problems: `hashicorp/go-version` for semver,
`hashicorp/hcl/v2` for HCL parsing, `crypto` stdlib for PGP/SHA256,
`archive/zip` for extraction, `net/http` for downloads, and
cross-compilation to every platform from a single build.

A third-party project (`tofuutils/tenv`) has already produced a Go-based
tool that covers some of the same ground. However, tenv diverged
significantly from tfenv: it supports five tools (Terraform, OpenTofu,
Terragrunt, Terramate, Atmos), changed the command structure, added scope
beyond version management (interactive TUI, YAML remote configs), and
relicensed under Apache 2.0. It is not a drop-in replacement for tfenv
and does not maintain behavioural parity.

tfenv's value proposition is — and always has been — **doing one thing
well**. The question is not "Bash or Go?" but "why not both?"

## Decision

tfenv will adopt a **dual-implementation strategy**:

1. **Bash remains the canonical, primary implementation.** It is the
   reference specification. All features are designed and proven in Bash
   first. The Bash edition lives in the existing `bin/`, `lib/`,
   `libexec/` directory structure unchanged.

2. **A Go edition is introduced as a secondary implementation** providing
   identical user-facing behaviour in a single compiled binary. The Go
   edition lives under a new top-level `go/` directory within the same
   repository.

3. **Perfect feature parity is enforced** via a shared acceptance test
   suite. Both implementations must pass the same behavioural tests. If
   one implementation adds a feature, the other must follow before the
   next release.

4. **The Go edition is a compiled alternative, not a replacement.** Users
   choose their preferred distribution: source the Bash scripts as today,
   or download a single Go binary. Both are first-class.

### Clean-Room Implementation Constraint

The Go edition MUST be developed as a **clean-room implementation**
derived exclusively from:

- **tfenv's own Bash source code** (MIT, we own it)
- **tfenv's own documentation and test suite** (the behavioural spec)
- **Public API documentation** from HashiCorp (releases.hashicorp.com
  URL patterns, SHA256SUMS format, PGP signature format)
- **Go standard library and dependency documentation** (go-version,
  hcl/v2, gopenpgp — all independently licensed)

Contributors to the Go edition **MUST NOT**:

- Read, reference, copy, or adapt source code from `tofuutils/tenv`
  or `tofuutils/tofuenv` (Apache 2.0)
- Read, reference, copy, or adapt source code from
  `hashicorp/hc-install` (MPL 2.0)
- Use any third-party tfenv-like tool's source as a reference for
  implementation decisions
- Use AI tools trained on or prompted with code from the above projects
  to generate code for the Go edition

The sole authoritative reference for the Go edition's behaviour is
tfenv's Bash source code and its test suite. If the Bash code does X,
the Go code must also do X — but the *how* must be independently
conceived using Go idioms and Go ecosystem libraries.

**Rationale:** `tofuutils/tenv` is licensed under Apache 2.0.
`hashicorp/hc-install` is licensed under MPL 2.0. Both are
copyleft-adjacent licences with attribution and/or source-disclosure
requirements that are incompatible with tfenv's MIT licence. Even
incidental structural similarity derived from reading their code could
constitute a derivative work and expose the project to licence
contamination claims. A clean-room discipline eliminates this risk
entirely.

This constraint applies to all contributors — human and AI agents
alike. AI coding agents working on the Go edition must be instructed
not to reference external implementations. Code review must verify
that implementations are idiomatically Go and not transliterations of
another project's approach.

### Testing Strategy

The Go edition introduces a proper, enterprise-grade test suite that
replaces the existing Bash test scripts as the authoritative parity
gate. The current Bash tests are integration-only, require network
access, lack proper assertions, and the runner exits 0 regardless of
failures — they are not fit for purpose as a quality gate.

The new test architecture has two tiers:

**Tier 1: Acceptance tests (black-box, CLI-level).** Written in Go
using `os/exec` to invoke whichever `tfenv` binary is under test. A
build tag or environment variable (`TFENV_TEST_BINARY`) selects the
target:

```
TFENV_TEST_BINARY=/path/to/bash/bin/tfenv   → test Bash edition
TFENV_TEST_BINARY=/path/to/go/tfenv         → test Go edition
```

CI runs the acceptance suite twice — once per edition — in a matrix.
If either edition fails, the build is red. This is the parity
enforcement mechanism.

These tests use `httptest.Server` to mock HashiCorp's release
infrastructure where possible, eliminating network dependency for
most scenarios. A subset of tests tagged `integration` hit the real
remote for end-to-end validation.

**Tier 2: Unit tests (Go internals only).** Cover version parsing,
semver constraint evaluation, `.terraform-version` file resolution,
HCL `required_version` extraction, platform detection, and download
verification logic. These are fast, isolated, and parallelised.
Table-driven tests ensure comprehensive edge-case coverage for areas
that Bash could never practically unit-test.

```
         ┌────────────────────────────────────────┐
         │   Acceptance Tests (Go, os/exec)       │
         │   Black-box CLI parity enforcement     │
         │   httptest.Server for mocked remote    │
         │   TFENV_TEST_BINARY selects edition    │
         └──────────┬───────────┬─────────────────┘
                    │           │
            ┌───────▼──┐  ┌────▼───────┐
            │  Bash    │  │  Go        │
            │  Edition │  │  Edition   │
            │ (bin/,   │  │ (go/,      │
            │  lib/,   │  │  single    │
            │  libexec)│  │  binary)   │
            └──────────┘  └──┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  Unit Tests (Go) │
                    │  Internal logic  │
                    │  Fast, isolated  │
                    └──────────────────┘
```

The existing Bash tests in `test/` are retained as legacy smoke tests
but are no longer the quality gate. They may be retired once the Go
acceptance suite reaches full coverage.

**Test capabilities gained:**

- Real pass/fail exit codes (not silent exit 0)
- `t.Run()` subtests and table-driven patterns
- `t.Parallel()` for concurrent execution
- `httptest.Server` for network-free download testing
- `t.TempDir()` with automatic cleanup
- Built-in coverage reporting and `-race` detection
- Proper assertion libraries (`testify/assert` or stdlib)
- CI integration with granular failure reporting

### Go Edition Scope

The Go binary must support:

- All existing `tfenv` subcommands with identical arguments
- All `TFENV_*` environment variables with identical semantics
- The `terraform` shim via multi-call binary pattern (see below)
- `.terraform-version` file resolution with identical precedence
- Version keywords: `latest`, `latest:<regex>`, `latest-allowed`,
  `min-required`
- PGP signature and SHA256 hash verification
- Custom remote mirrors via `TFENV_REMOTE`
- Platform detection and architecture override via `TFENV_ARCH`
- Concurrent install locking
- Identical exit codes for scripted consumption

**Output and UX are NOT constrained to match the Bash edition.** The
Go edition is free to — and should — significantly improve output
formatting, progress indicators, error messages, and general UX.
Functional parity (same commands, same results) is required;
presentational parity is not. Where the Go edition discovers better
UX patterns, those insights should flow back to improve the Bash
edition where practical.

### State Directory

Both editions share the same state layout via `TFENV_CONFIG_DIR`:

```
~/.tfenv/                         # Default TFENV_CONFIG_DIR
├── versions/                     # Installed Terraform binaries
│   ├── 1.5.0/
│   └── 1.6.1/
└── version                       # Default version file
```

The Go edition defaults `TFENV_CONFIG_DIR` to `~/.tfenv/` (the binary
has no "install directory" to conflate with state). The Bash edition
continues to default `TFENV_CONFIG_DIR` to `TFENV_ROOT` for backwards
compatibility.

When both editions point at the same `TFENV_CONFIG_DIR`, they share
installed versions and default version seamlessly. This enables
scenarios like Bash on a dev laptop and Go binary in CI, managing the
same versions directory.

This also resolves the longstanding multi-user system issue: system-wide
binary install, per-user `~/.tfenv/` state.

### Multi-Call Binary and Installation

The Go edition uses a **multi-call binary** pattern. A single compiled
binary serves as both `tfenv` and the `terraform` shim, determined by
`os.Args[0]`:

- Invoked as `tfenv` → command mode (install, use, list, etc.)
- Invoked as `terraform` → shim mode (resolve version, exec terraform)

**Distribution creates both entry points:**

- **Package managers** (Homebrew, deb, rpm): install both `tfenv` and
  `terraform` binaries (hardlinks to the same file)
- **Release tarballs/zips**: contain both `tfenv` and `terraform`
- **Single binary download**: `tfenv init` creates a `terraform`
  hardlink (or copy on filesystems without hardlink support) next to
  itself. Idempotent, safe to re-run.

Hardlinks are preferred over symlinks to avoid Windows symlink
privilege requirements.

### PGP Key Handling

The Go edition **embeds** the HashiCorp PGP public key at compile time
via `//go:embed`, sourced from `share/hashicorp-keys.pgp` in the
repository. This makes the binary fully self-contained with no runtime
file dependencies for verification.

An override mechanism allows users to supply a custom key path via
environment variable or config file, supporting private mirrors with
different signing keys. The embedded key is the default; the override
takes precedence when set.

### Logging

The Go edition's logging system is a **ground-up design** — it does
not replicate bashlog. The `BASHLOG_*` environment variables are not
supported and are not part of the Go edition's interface.

The logging system should be best-in-class: the most capable,
flexible, and beautifully presented logging of any CLI tool in the
ecosystem. The designer has full latitude to research and specify the
optimal Go logging libraries, output formats, and UX patterns.
Requirements:

- Colourised, human-friendly output by default when stderr is a TTY
- Clean, machine-parseable output when piped or in CI
- Structured JSON mode available via environment variable
- Verbosity levels (error/warn/info/debug) via `TFENV_LOG_LEVEL`
- Fast — logging must never be a performance bottleneck

### Tool Scope

The Go edition manages **Terraform only** (HashiCorp releases). This
is tfenv's core value proposition: do one thing well.

However, the internal architecture should be designed with extensibility
in mind — clean interfaces for "a tool that is downloaded from a remote,
verified, and version-managed" — so that adding support for other tools
in the future would not require an architectural rewrite. This is a
structural consideration, not a feature commitment.

### Delivery Approach

The Go edition ships with **full feature parity or not at all.** There
is no MVP or phased rollout. The first release that includes the Go
binary must pass the complete acceptance test suite — every command,
every feature, every edge case. Until then, the Go edition is
unreleased development work.

This ensures users never encounter a half-finished alternative. When
the Go edition ships, it is a complete, drop-in choice.

### Repository Layout

```
tfenv/
├── bin/              # Bash entry points (unchanged)
├── lib/              # Bash libraries (unchanged)
├── libexec/          # Bash subcommands (unchanged)
├── go/               # Go edition — all Go code lives here
│   ├── cmd/
│   │   └── tfenv/    # Main binary entry point (multi-call)
│   ├── internal/     # Internal packages
│   │   ├── cli/      # Command dispatch
│   │   ├── config/   # Environment & file config, state dir
│   │   ├── install/  # Download, verify, extract
│   │   ├── list/     # Local & remote listing
│   │   ├── logging/  # Logging framework
│   │   ├── resolve/  # Version resolution & constraints
│   │   ├── shim/     # Terraform proxy (multi-call handler)
│   │   └── platform/ # OS/arch detection
│   ├── test/
│   │   ├── acceptance/  # Black-box CLI tests (both editions)
│   │   └── testdata/    # Fixtures (.tf files, mock responses)
│   ├── go.mod
│   └── go.sum
├── test/             # Legacy Bash tests (retained, not the gate)
├── share/            # Static assets (PGP keys — shared, embedded by Go)
└── docs/adr/         # Architecture decisions
```

### Clone-Install Compatibility

Many users install tfenv by cloning the repository and adding `bin/`
to their `PATH`. CI pipelines, Docker images, and automation scripts
rely on this pattern. The Go edition **must not break or degrade this
workflow in any way.**

Concretely:

- The `go/` directory is inert source code. It contains no scripts,
  binaries, or artefacts that the Bash edition references or executes.
  A `git clone` that includes `go/` is functionally identical to one
  that does not — the Bash edition does not know `go/` exists.
- No Go toolchain is required to use the Bash edition. The Go source
  is only relevant to contributors building the Go binary.
- No files in `bin/`, `lib/`, `libexec/`, or `share/` may be modified,
  moved, or removed to accommodate the Go edition. These directories
  are the Bash edition's contract with clone-install users.
- The `go.mod` and `go.sum` files live inside `go/`, not at the
  repository root, to avoid polluting the top-level directory that
  clone-install users interact with.
- Repository size impact is negligible — Go source is plain text, and
  the clone is already small (~2.5k LOC).

If a future change would impact clone-install users (e.g. repository
restructuring, CI changes that affect clone behaviour, or build
artefacts appearing in tracked directories), it requires its own ADR
and explicit maintainer approval.

### Dependencies (Go Edition)

| Dependency | Purpose |
|------------|---------|
| `hashicorp/go-version` | Semver parsing and constraint evaluation |
| `hashicorp/hcl/v2` | Parse `required_version` from `.tf` files |
| `ProtonMail/gopenpgp` (or `golang.org/x/crypto/openpgp`) | PGP signature verification |
| stdlib `crypto/sha256` | Hash verification |
| stdlib `archive/zip` | Binary extraction |
| stdlib `net/http` | Downloads |
| stdlib `os`, `path/filepath` | File operations, version-file walk |
| TBD — logging library | Designer to specify based on research |
| TBD — CLI framework (if warranted) | Designer to evaluate need |

### Release Strategy

- Bash edition: continues to ship as the Git-cloned source tree
- Go edition: cross-compiled via `goreleaser`, distributed as platform
  binaries alongside GitHub Releases, with Homebrew tap, `.deb`, `.rpm`,
  and `.zip` artifacts. Each package contains both `tfenv` and
  `terraform` entry points (hardlinks to the same binary).
- Both editions share the same version number and CHANGELOG
- A release is only cut when both editions pass the full acceptance
  test suite — no exceptions, no partial Go releases

## Consequences

### Positive

- Users on macOS no longer need `ggrep` or Homebrew for basic operation
- Windows becomes a first-class platform (native `.exe`, no git-bash
  dependency)
- PGP and SHA256 verification work everywhere without external tools
  (keys embedded in binary)
- Single-binary distribution via GitHub Releases, Homebrew, etc.
- Multi-user systems work naturally: system-wide binary, per-user
  `~/.tfenv/` state directory
- Enterprise-grade test suite: proper assertions, subtests, parallel
  execution, coverage reporting, race detection, HTTP mocking — all
  enforcing parity between editions
- Unit tests become possible for internal logic (version parsing,
  constraint evaluation, platform detection)
- Network-free testing for most scenarios via `httptest.Server`
- Output and UX can be dramatically improved in the Go edition, with
  successful patterns flowing back to improve the Bash edition
- World-class logging system unconstrained by bashlog's limitations
- The Go edition can be embedded as a library by other Go tools
  (similar to `hashicorp/hc-install`)
- Internal architecture supports future tool extensibility without
  committing to multi-tool scope
- Both editions can share state, enabling mixed-edition workflows
- tfenv remains the authoritative, Terraform-focused version manager
  under MIT license
- The Bash edition continues to serve users who prefer script-based
  tools with zero compilation step

### Negative

- Two implementations must be kept in sync — every feature change is
  double the work
- Go introduces a build toolchain requirement for contributors to the
  Go edition (though not for users — they get pre-built binaries)
- The repository becomes more complex with two codebases and a
  multi-tier test architecture
- Edge-case behavioural divergence is possible despite shared tests —
  regex dialect differences (ERE vs RE2), platform-specific path
  handling, and error message formatting require vigilance
- CI build time increases (acceptance tests × 2 editions + Go unit
  tests + Go cross-compilation)
- The acceptance test suite must be maintained as a third artefact
  alongside the two implementations

## Future Capabilities

The Go edition unlocks a category of features that are impractical or
impossible in Bash. These are not commitments — they are the vision
that justifies the investment. Each would be specified via its own
issue and designed by `feature-designer` before implementation.

**Note:** tfenv is a download facilitator, not a distribution channel.
All features respect HashiCorp's licence — we point users to
HashiCorp's own release infrastructure (or configured mirrors) for
binaries. We never host, embed, bundle, or redistribute Terraform.

### Trust & Supply Chain Security

- **Cosign-signed releases.** Every tfenv Go binary release is signed
  with Sigstore cosign, with SLSA provenance attestations and an SBOM
  (Software Bill of Materials). Reproducible builds where possible.
  Positions tfenv as the most supply-chain-secure version manager in
  the ecosystem.
- **Proxy binary integrity check.** Before exec'ing a resolved
  `terraform` binary, optionally verify its SHA256 against the cached
  checksum from the original install. Detects post-install tampering.
  Off by default, enabled via policy. For high-security environments.

### Debugging & Observability

- **`tfenv why`** — explain the full decision trace for version
  selection. "1.6.1 was selected because `.terraform-version` at
  `/home/jaz/project/.terraform-version` contains `latest-allowed`,
  and `required_version >= 1.5.0, < 1.7.0` in `versions.tf` resolved
  to 1.6.1." Eliminates the #1 debugging pain point.
- **`tfenv doctor`** — diagnostic command: is the shim on PATH
  correctly? Is `TFENV_CONFIG_DIR` writable? Can we reach the remote?
  Is the PGP key valid? Are installed versions known-vulnerable? One
  command, full health check.
- **`tfenv config`** — dump effective configuration: all env vars,
  resolved state directory, active version file path, remote URL,
  architecture, auto-install setting. JSON-compatible for scripting.
- **`tfenv audit`** — scan a project and report: selected version,
  active constraint, version file path, installed status, latest
  available. One-shot project health check. CI-friendly.
- **Built-in timing.** Every operation logs duration at debug level.
  "Downloaded 1.6.1 in 3.2s", "Version resolution took 12ms", "PGP
  verification took 45ms." Instant performance diagnosis.

### Enterprise & Team Features

- **Policy files.** `.tfenv.toml` or `.tfenv.yaml` in a project root
  with team-wide settings: allowed version ranges, required
  verification, forbidden pre-release versions, custom mirror URL.
  Checked into git, shared by the team. Overrides individual env vars.
- **`TFENV_ALLOWED_VERSIONS` constraint.** Env var or policy setting
  restricting which versions can be installed or used. "Only 1.5.x and
  1.6.x are approved." Out-of-range attempts fail with a clear error.
- **Air-gapped / enterprise-first features.** Local binary caching
  (don't re-download verified versions), mirror fallback chains
  (primary → secondary → tertiary), proxy authentication
  (HTTPS_PROXY with auth), offline install from a local directory.
- **Version health intelligence.** Know which Terraform versions have
  known CVEs, are EOL, or are deprecated. `tfenv list` shows warnings
  next to vulnerable versions. `tfenv install latest` skips known-bad
  versions. Data from HashiCorp security advisories or a curated list.

### Automation & Integration

- **JSON output mode for every command.** `tfenv list --json`,
  `tfenv list-remote --json`, `tfenv version-name --json`. Structured
  machine-readable output. CI pipelines, dashboards, automation
  scripts — they all want JSON, not grep-able text.
- **Shell completions from the binary.** `tfenv completion bash`,
  `tfenv completion zsh`, `tfenv completion fish`,
  `tfenv completion powershell`. Built into the binary, always
  up-to-date, one command to install. Version-aware: `tfenv use <TAB>`
  completes from installed versions.
- **Git hook integration.** `tfenv hook install` drops a
  `.git/hooks/post-checkout` that auto-switches version when changing
  branches with different `.terraform-version` files. Optional,
  user-initiated, non-invasive.
- **Terraform lock file awareness.** Read `.terraform.lock.hcl` to
  cross-reference provider versions with Terraform version
  compatibility. Warn before switching to incompatible versions.

### Robustness & Resilience

- **Atomic installs with rollback.** Download to a temp directory,
  verify, then atomic rename into `versions/`. If anything fails —
  network, verification, disk — previous state is untouched. No
  half-installed versions.
- **Retry with exponential backoff.** Network failures during download
  retry with backoff and jitter. Flaky corporate proxies, transient
  DNS failures — handled gracefully.
- **Checksum cache.** After verifying a downloaded binary, cache the
  SHA256 alongside it. On subsequent installs of the same version,
  skip download if cached binary matches. Never re-download what
  you've already verified.

### UX Polish

- **Self-update.** `tfenv update` — the Go binary updates itself
  in-place. Download new version, verify signature, replace. No
  package manager round-trip for direct installs.
- **Update notifications.** When a newer tfenv version exists, print a
  subtle one-liner (once per day, cached, non-blocking). Opt-out via
  env var. Awareness, not nagging.
- **Fuzzy version matching.** `tfenv install 1.6` installs latest
  `1.6.x`. `tfenv install 1` installs latest `1.x.x`. Intuitive
  shorthand beyond the existing `latest:^1.6` syntax.
- **Interactive version picker.** When `tfenv install` is run in a TTY
  without arguments and no `.terraform-version` exists, show a
  filterable list. Up/down, type to filter, enter to install. Only
  interactive — never in CI.
- **Colourised diff on version switch.** `tfenv use 1.6.1` shows
  `Switched: 1.5.0 → 1.6.1`. Communicates the change, not just the
  result.
- **Download progress with ETA.** Native progress bar with bytes
  transferred, speed, and estimated time remaining. Not curl's output
  proxied through stderr — a polished, responsive, beautiful
  indicator.
- **Published performance benchmarks.** Time-to-first-terraform-run,
  version resolution latency, install throughput. If we're faster —
  and with Go vs tenv's 5-tool abstraction layers, we will be — make
  it undeniable and public.

### Strategic Positioning

These capabilities collectively tell a clear story: tfenv has
**depth** where tenv has breadth. tfenv is the most polished, secure,
observable, enterprise-ready Terraform version manager that exists.
One tool, done exceptionally well.

### Cross-Pollination: Go as R&D Lab for Bash

The Go edition is not a walled garden. It is an **innovation pipeline**
that feeds improvements back into the Bash edition wherever practical.

Not every Go feature can be backported — some depend on Go's
capabilities (embedded keys, interactive TUI, self-update). But many
can:

- **`tfenv why`** — decision trace logic is pure text output; Bash
  can implement the same walkthrough
- **`tfenv doctor`** — health checks are shell-native (test PATH,
  check permissions, probe HTTP endpoints)
- **`tfenv config`** — dumping effective config is trivial in Bash
- **`tfenv audit`** — project scanning is grep and file walking
- **JSON output** — Bash can emit JSON for structured commands
  (version-name, list, config)
- **Fuzzy version matching** — `tfenv install 1.6` resolving to
  latest `1.6.x` is regex shorthand Bash already nearly supports
- **Colourised version switch diff** — Bash already has colour via
  bashlog; showing `1.5.0 → 1.6.1` is a one-line improvement
- **Policy files** — Bash can read a simple config file and enforce
  allowed version ranges
- **Atomic installs** — Bash can download-to-temp-then-rename; it
  partially does this already
- **Retry with backoff** — curl retry flags exist; wrapping them
  properly is straightforward
- **Built-in timing** — Bash can timestamp operations at debug level

The workflow is: Go edition pioneers a feature → proves the UX and
design → Bash edition adopts a compatible implementation where the
feature is achievable in shell. The acceptance test suite (written in
Go) validates both. Features that genuinely require Go capabilities
remain Go-only — that's fine, it's a reason to choose the Go edition.

This creates a **virtuous cycle**: the Go edition pushes the boundary,
the Bash edition rises with it, and both editions get better because
the other exists.

## Alternatives Considered

### Alternative 1: Full Rewrite — Replace Bash With Go

Abandon the Bash implementation entirely and go all-in on Go.

Rejected because: The Bash edition is the soul of the project. It works,
it has a decade of battle-testing, and many users value its
transparency — they can read and modify every line. Dropping it would
alienate a significant portion of the user base and discard a working,
proven implementation for no good reason.

### Alternative 2: Adopt tenv as Successor

Recommend `tofuutils/tenv` as the Go-based successor and put tfenv into
maintenance mode.

Rejected because: tenv diverged from tfenv's design philosophy. It
supports five tools, changed the command structure, expanded scope
beyond version management (TUI menus, YAML remote configs), and
relicensed under Apache 2.0. It is not a drop-in replacement for
tfenv and serves a different audience with different goals.

### Alternative 2b: Fork tenv and Strip It Down

Fork `tofuutils/tenv`, remove non-Terraform tools, and rebrand as
tfenv's Go edition.

Rejected because: tenv is Apache 2.0. Forking it would require tfenv
to either adopt Apache 2.0 for the Go edition (licence fragmentation
within the project) or relicense the fork to MIT (which Apache 2.0
does not permit without the original authors' consent). Additionally,
tenv's architecture is built around multi-tool abstractions that would
be expensive to remove cleanly. A clean-room implementation from our
own Bash spec is both legally cleaner and architecturally simpler.

### Alternative 3: Keep Bash Only

Continue with the Bash implementation and address platform issues
incrementally.

Not fully rejected — this is the status quo and remains viable. However,
some problems (Windows support, external tool dependencies, testability)
are structural to Bash and cannot be incrementally solved. The Go
edition addresses these without sacrificing the Bash edition.

### Alternative 4: Wrapper/Transpiler Approach

Auto-generate Go from Bash or vice versa.

Rejected because: The languages are too different for meaningful
transpilation. The result would be worse than either hand-written
implementation. Maintaining a transpiler would be harder than maintaining
two implementations.
