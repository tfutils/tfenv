# 003. No ShellCheck

**Status:** Accepted

**Date:** 2026-04-24

## Context

ShellCheck is the de facto static analysis tool for shell scripts. Many
bash projects use it as part of their CI pipeline. However, tfenv has
historically not used ShellCheck and its coding patterns sometimes
diverge from ShellCheck's recommendations.

## Decision

tfenv will NOT adopt ShellCheck. This is a deliberate choice, not an
oversight. The project's bash style is self-consistent and enforced
through code review and documented coding standards rather than automated
linting.

Adding ShellCheck at this stage would require significant refactoring of
existing code to satisfy its rules, many of which are stylistic rather
than correctness-related. The cost outweighs the benefit for a mature
codebase of this size.

## Consequences

### Positive

- No churn from retrofitting existing code to satisfy a linter
- Coding style remains under human control, not tool-dictated
- No additional CI dependency

### Negative

- Some classes of shell bugs that ShellCheck catches must be caught by
  code review instead
- Contributors familiar with ShellCheck-enforced projects may find the
  style unfamiliar

## Alternatives Considered

### Alternative 1: Adopt ShellCheck with Exclusions

Enable ShellCheck with a curated set of disabled rules. Rejected because
the exclusion list would be large and maintaining it adds ongoing overhead
for marginal benefit.
