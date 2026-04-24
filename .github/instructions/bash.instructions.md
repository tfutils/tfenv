---
applyTo: '**/*.sh'
---
# Bash Coding Standards — tfenv

These standards apply to ALL bash scripts in the tfenv repository. Read
this entire document BEFORE writing or modifying any bash code.

## Mandatory Pre-Write Verification

Before writing ANY bash code, verify you understand these rules. Violations
caught after writing waste time — get it right first.

## Shell Setup

Every script in `libexec/` contains its own boilerplate for resolving
`TFENV_ROOT` and sourcing helpers. This is **intentional** — each script
must be executable in isolation. Do NOT refactor this into a shared loader.

```bash
#!/usr/bin/env bash
set -uo pipefail;
```

- Shebang: `#!/usr/bin/env bash` (exact format)
- Strict mode: `set -uo pipefail;` — NEVER use `set -e`
- The project does NOT use shellcheck — this is a deliberate choice

## Indentation and Formatting

- 2-space indentation throughout (no tabs)
- Terminate ALL statements with a semicolon (`;`) where syntactically correct
- Include vim modeline at end of new scripts:
  `# vim: set syntax=bash tabstop=2 softtabstop=2 shiftwidth=2 expandtab smarttab :`

## Variables

- ALL variable references use braces: `${variable}` (never `$variable`)
- Script-local variables use lowercase: `${version}`, `${remote}`
- Environment variables use UPPERCASE: `${TFENV_ROOT}`, `${TFENV_ARCH}`
- Always quote variable expansions: `"${variable}"`
- Quote entire strings containing variables: `"${dir}/file"` NOT `"${dir}"/file`
- Use `${var:-default}` for optional variables (scripts run with `set -u`)

## Quoting

- Single-quote strings unless variable expansion is needed
- Double-quote strings that contain variable expansion
- NEVER use unescaped `!` inside double-quoted strings (bash history expansion)

## Command Substitution

- Use `$(...)` NOT backticks

## Error Handling

- NEVER use `set -e` — handle errors explicitly
- Use the project's existing error patterns:
  - `error_and_die` for fatal errors (available via `lib/helpers.sh`)
  - Inline `||` pattern for simple cases:
    ```bash
    some_command || error_and_die "Failed to do thing";
    ```
- Note: `error_and_die` does NOT exist in the test context — tests use
  `error_and_proceed` instead

## Functions

- Define as: `function name() { ... };`
- Use `local` for variables inside functions

## Cross-Platform Considerations

- macOS uses BSD `sed`/`grep`/`readlink` — GNU extensions may not be available
- The project installs `ggrep` on macOS but uses system `sed` and `readlink`
- `.terraform-version` files may contain `\r` from Windows/WSL — always strip
  with `tr -d '\r'` after reading

## Common Pitfalls

1. **Shell operator precedence:** `cmd1 || cmd2 | cmd3` means
   `cmd1 || (cmd2 | cmd3)`, not `(cmd1 || cmd2) | cmd3`
2. **Unquoted variables:** Always quote `"${var}"` — unquoted variables cause
   word-splitting
3. **Double-quoted traps:** `trap "rm ${var}" EXIT` expands at definition time.
   Use functions or single quotes.
4. **`$@` in for-loops:** Always quote: `for arg in "$@"`
5. **Regex anchoring:** `^1.1` matches `1.10.x` because `.` is a regex
   wildcard. Use `^1\.1\.` for exact prefix matching.

## Pre-Submission Checklist

- [ ] `#!/usr/bin/env bash` shebang
- [ ] `set -uo pipefail;` (not `set -e`)
- [ ] 2-space indentation, no tabs
- [ ] All variables use `${braces}`
- [ ] All variable expansions double-quoted
- [ ] All statements terminated with `;`
- [ ] No unescaped `!` in double-quoted strings
- [ ] `$(...)` for command substitution (no backticks)
- [ ] `local` used inside functions
- [ ] Cross-platform compatibility considered
