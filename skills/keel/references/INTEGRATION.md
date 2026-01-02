# Integration Guide

How to integrate keel with external systems.

## External References

The `--refs` flag links decisions to external tracking systems. Keel stores these as plain strings - it doesn't validate or fetch from external systems.

### Beads Integration

[Beads](https://github.com/steveyegge/beads) is a lightweight issue tracker for AI agents.

**Link decisions to Beads issues:**
```bash
keel decide \
  --type product \
  --problem "Need to handle rate limiting" \
  --choice "Token bucket algorithm" \
  --refs "bd-auth-123" \
  --agent
```

**Query decisions for a Beads issue:**
```bash
keel context --ref bd-auth-123
```

**Workflow:**
- `bd` tracks *what* to do (tasks, dependencies, status)
- `keel` tracks *why* it was done (decisions, rationale)

### Jira Integration

```bash
keel decide \
  --type constraint \
  --problem "Security audit requirement" \
  --choice "All API endpoints require authentication" \
  --refs "PROJ-456" \
  --files "src/middleware/auth.ts"
```

Query by Jira ticket:
```bash
keel context --ref PROJ-456
```

### GitHub Issues

```bash
keel decide \
  --type product \
  --problem "Tried WebSockets for notifications" \
  --choice "Switched to SSE - simpler for our use case" \
  --refs "gh-123"
```

### Linear

```bash
keel decide \
  --type product \
  --problem "User requested dark mode" \
  --choice "CSS variables with prefers-color-scheme" \
  --refs "LIN-abc-123"
```

### Multiple References

Link to multiple systems at once:

```bash
keel decide \
  --type product \
  --problem "Authentication redesign" \
  --choice "OAuth2 with PKCE" \
  --refs "bd-auth-epic,JIRA-789,gh-42"
```

## JSON Output

All commands support `--json` for programmatic use:

```bash
keel context --json src/auth.ts
keel sql "SELECT raw_json FROM decisions WHERE problem LIKE '%auth%'" --json
keel why --json DEC-a1b2
keel curate --json --older-than 30
```

## CI/CD Integration

### Validate on PR

```yaml
# .github/workflows/keel.yml
name: Check Constraints
on: [pull_request]
jobs:
  constraints:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: |
          curl -fsSL https://raw.githubusercontent.com/TYRONEMICHAEL/keel/main/scripts/install.sh | bash
          keel sql "SELECT raw_json FROM decisions WHERE type = 'constraint'" --json
```

### Audit Decisions in CI

```bash
# List all constraints
keel sql "SELECT raw_json FROM decisions WHERE type = 'constraint'" --json

# List recent decisions
keel sql "SELECT raw_json FROM decisions ORDER BY created_at DESC LIMIT 20" --json
```

## IDE Integration

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Keel: Context for current file",
      "type": "shell",
      "command": "keel context ${file}"
    }
  ]
}
```

### Terminal Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias kd='keel decide'
alias kc='keel context'
alias kw='keel why'
```
