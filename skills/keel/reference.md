# Keel Reference

## Integration with External Systems

Keel's `--refs` flag links decisions to external tracking systems. Use any ID format.

### Beads Integration

[Beads](https://github.com/steveyegge/beads) is a lightweight issue tracker for AI agents.

```bash
# Link decision to a Beads issue
keel decide \
  --type product \
  --problem "Need to handle rate limiting" \
  --choice "Token bucket algorithm" \
  --refs "bd-auth-123"

# Query decisions for a Beads issue
keel context --ref bd-auth-123
```

**Workflow with Beads:**
- `bd` tracks *what* to do (tasks, dependencies, status)
- `keel` tracks *why* it was done (decisions, rationale)

### Jira Integration

```bash
keel decide \
  --type constraint \
  --problem "Security audit requirement" \
  --choice "All API endpoints require authentication" \
  --refs "PROJ-456"
```

### GitHub Issues

```bash
keel decide \
  --type learning \
  --problem "Tried WebSockets for notifications" \
  --choice "Switched to SSE - simpler for our use case" \
  --refs "gh-123"
```

### Multiple References

Link to multiple systems:

```bash
keel decide \
  --type product \
  --problem "User requested dark mode" \
  --choice "CSS variables with prefers-color-scheme" \
  --refs "bd-ui-456,JIRA-789,gh-42"
```

## Decision Type Guidelines

| Type | Use When | Examples |
|------|----------|----------|
| `product` | Affects user-facing behavior | Feature limits, UI choices, pricing |
| `process` | Affects how code is written | Code style, patterns, conventions |
| `constraint` | Hard limit that can't be violated | Security, compliance, performance |
| `learning` | Knowledge from trying something | Failed approaches, discoveries |

## CLI Quick Reference

```bash
# Record
keel decide --type <type> --problem "..." --choice "..." [--rationale "..."] [--files "..."] [--refs "..."] [--agent]

# Query
keel context <file>              # Decisions for a file
keel context --ref <id>          # Decisions for a reference
keel search "query"              # Full-text search
keel search --type constraint    # Filter by type
keel why <id>                    # Full decision details

# Manage
keel supersede <id> --choice "..." # Replace a decision
keel validate                      # Check file references exist
keel curate --older-than 30        # Get old decisions for summarization

# Maintain
keel upgrade                       # Update to latest version
keel upgrade --check               # Check for updates
```

## JSON Output

All commands support `--json` for programmatic use:

```bash
keel context --json src/auth.ts
keel search --json "authentication"
keel why --json DEC-a1b2
```
