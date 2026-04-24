# 002. Standalone Script Execution

**Status:** Accepted

**Date:** 2025-04-24

## Context

Every script in `libexec/` contains its own boilerplate for resolving
`TFENV_ROOT` and sourcing shared helpers (`lib/helpers.sh`, `lib/bashlog.sh`,
etc.). This appears to be duplication and a natural refactoring target.

However, the design is intentional. Each `libexec/` script must be
independently executable for:
- Direct invocation during testing (`./libexec/tfenv-install`)
- Debugging individual commands in isolation
- Compatibility with the rbenv-inspired architecture where commands are
  standalone executables discovered by the dispatcher

## Decision

Every `libexec/` script will continue to contain its own `TFENV_ROOT`
resolution and helper sourcing boilerplate. This will NOT be refactored
into a shared loader, even though it creates apparent duplication.

## Consequences

### Positive

- Each command can be tested and debugged independently
- No hidden coupling between the dispatcher and individual commands
- Matches the established rbenv pattern that tfenv is modelled after

### Negative

- Boilerplate is duplicated across ~15 scripts
- Changes to the sourcing pattern must be applied to every script

## Alternatives Considered

### Alternative 1: Shared Init Script

Create a single `lib/init.sh` that all scripts source. Rejected because
it creates a dependency that breaks standalone execution — if a script
cannot find the init script, it cannot run at all.
