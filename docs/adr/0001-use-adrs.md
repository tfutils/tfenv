# 001. Use Architecture Decision Records

**Status:** Accepted

**Date:** 2026-04-24

## Context

tfenv has accumulated design decisions over its 8+ year history that are
not documented anywhere. New contributors (human or agent) have to reverse-
engineer the reasoning behind design choices from code and commit messages.

This makes it difficult to:
- Understand why things are the way they are
- Know which constraints are deliberate vs accidental
- Make informed decisions about changes

## Decision

We will use Architecture Decision Records (ADRs) to document significant
technical decisions. ADRs live in `docs/adr/` and follow a sequential
numbering scheme (`NNN-title.md`). The template is in `0000-template.md`.

An ADR is warranted when:
- A design choice affects multiple components
- The decision is non-obvious or has trade-offs worth documenting
- Future contributors might reasonably question why something works this way

## Consequences

### Positive

- Design rationale is discoverable and permanent
- New contributors can onboard faster
- Agents can load relevant ADRs before making changes

### Negative

- Adds a small overhead to the decision-making process
- Historical decisions remain undocumented until retroactively recorded

## Alternatives Considered

### Alternative 1: Inline Code Comments

Document decisions as comments in the relevant source files. Rejected because
decisions often span multiple files and comments get stale when code moves.

### Alternative 2: Wiki

Use the GitHub wiki for design documentation. Rejected because wikis are
disconnected from the codebase and not version-controlled alongside the code.
