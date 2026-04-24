# Copilot Instructions — tfenv

This file is intentionally minimal. The canonical source for project
architecture, conventions, and agent guidelines is [`AGENTS.md`](../../AGENTS.md)
at the repository root.

## Auto-Loaded Instructions

Language-specific coding standards are in `.github/instructions/` and use
`applyTo` frontmatter to load automatically when editing matching files:

| File | Applies To | Purpose |
| ---- | ---------- | ------- |
| `bash.instructions.md` | `**/*.sh` | Bash coding standards |

## Agent Definitions

Agent definitions live in `.github/agents/` as `.agent.md` files. Each agent
has a YAML frontmatter block with `description`, `tools`, `model`, and
`argument-hint` fields.

## Key References

- **[AGENTS.md](../../AGENTS.md)** — master context: project overview, core
  principles, conventions, work type ownership, essential commands
- **[SECURITY.md](../../SECURITY.md)** — vulnerability reporting policy
- **[docs/adr/](../../docs/adr/)** — Architecture Decision Records
