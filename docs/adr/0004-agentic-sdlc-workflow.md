# 004. Agentic SDLC Workflow

**Status:** Accepted

**Date:** 2026-04-24

## Context

AI coding agents (GitHub Copilot, Claude, etc.) are increasingly used for
development tasks. Without structured guidance, agents make inconsistent
decisions about coding style, project conventions, and workflow. Multiple
agents operating concurrently on the same repository risk conflicts.

## Decision

We adopt a structured agentic SDLC with:

1. **11 specialist agents** defined in `.github/agents/` — each with a
   clear scope, constraints, and workflow
2. **Worktree-first convention** — all code changes happen in git worktrees
   under `.worktrees/`, never in the main working tree
3. **Claim protocol** — agents claim issues with `agent:in-progress` label
   before starting work, preventing duplicate effort
4. **`gh` CLI as primary GitHub interface** — REST API via `gh api` for
   reliability over the GitHub MCP server
5. **Agent ownership table** in `AGENTS.md` — maps work types to owning
   agents for clear routing

## Consequences

### Positive

- Multiple agents can work concurrently without conflicts
- Clear scope boundaries prevent agents from overstepping
- Structured workflow produces consistent, reviewable output
- Human maintains control via PR review — agents never merge

### Negative

- Significant upfront investment in agent definitions
- Agent definitions require maintenance as the project evolves
- Overhead is an upfront investment, but the framework is designed to be
  replicable across the organisation's repositories

## Alternatives Considered

### Alternative 1: Single General-Purpose Agent

Use one agent definition with broad scope. Rejected because it provides
no concurrency safety and the agent has no clear constraints.

### Alternative 2: No Agent Support

Rely on default Copilot behaviour without custom agents. Rejected because
the default behaviour does not respect project-specific conventions
(standalone boilerplate, no shellcheck, etc.).
