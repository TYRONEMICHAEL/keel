# Keel

> The structural foundation that keeps you stable when things get rough.

**Git-native decision ledger for coding agents.** Captures the "why" behind changes so agents can act with confidence instead of guessing.

## Quickstart

**1. Install**

```bash
curl -fsSL https://raw.githubusercontent.com/TYRONEMICHAEL/keel/main/scripts/install.sh | bash
```

**2. Initialize in your repo** (humans do this once)

```bash
keel init
```

**3. Install Claude Code plugin** (optional - for automatic skill loading)

```bash
claude plugin install --path /path/to/keel
```

Or manually tell your agent:

```bash
echo "Use 'keel' for decision tracking. Run 'keel context <file>' before editing." >> CLAUDE.md
```

**Upgrade**

```bash
keel upgrade
```

## Essential Commands

| Command | Purpose |
|---------|---------|
| `keel decide --type product --problem "..." --choice "..."` | Record a decision |
| `keel context <file>` | Get decisions affecting a file |
| `keel sql "SELECT ..."` | Query decisions with SQL |
| `keel why DEC-xxxx` | Show decision details |
| `keel graph` | Output decision graph as Mermaid |

## Why Keel?

When you touch code, you face questions:

- Why is this limit 5 and not 10?
- Has this approach been tried before?
- What constraints am I operating under?

Keel solves this. Every decision — human or agent — gets recorded with context, rationale, and links to the code it affects.

## Commands

### decide

Record a new decision:

```bash
keel decide \
  --type product \           # product, process, constraint
  --problem "..." \          # What problem this addresses
  --choice "..." \           # What was decided
  --rationale "..." \        # Why this choice (optional)
  --files "a.ts,b.ts" \      # Files this affects (optional)
  --refs "JIRA-123,bd-abc" \ # External refs: Jira, Beads, GitHub, etc. (optional)
  --agent                    # Mark as agent decision (optional)
```

### context

Get decisions for a file or reference:

```bash
keel context src/auth/oauth.ts
keel context --ref bd-abc      # Query by Beads issue, Jira ticket, etc.
keel context --json src/auth/oauth.ts
```

### sql

Query decisions with raw SQL:

```bash
keel sql "SELECT raw_json FROM decisions WHERE status = 'active'"
keel sql "SELECT * FROM decisions WHERE type = 'constraint'"
keel sql "SELECT raw_json FROM decisions WHERE problem LIKE '%auth%'"
```

### why

Show full decision details:

```bash
keel why DEC-a1b2
keel why a1b2        # Short form works too
```

### supersede

Replace a decision with a new one:

```bash
keel supersede DEC-a1b2 \
  --problem "New problem statement" \
  --choice "New choice"
```

### graph

Output decision relationships as Mermaid diagram:

```bash
keel graph              # Decisions with supersession chains and bead links
keel graph --files      # Also include file associations
```

Output can be pasted into any Mermaid viewer (GitHub, Notion, mermaid.live).

## Decision Types

| Type | When to Use | Example |
|------|-------------|---------|
| `product` | Business logic, features, architecture | "Free plan = 5 users" |
| `process` | How we work, patterns, style | "Use functional style, not OOP" |
| `constraint` | Hard limits, requirements | "Must support IE11" |

## Architecture

```
.keel/
├── decisions.jsonl   # Source of truth (git-tracked)
└── index.sqlite      # Derived index (gitignored)
```

**JSONL** is append-only and git-native. **SQLite** provides indexed queries. The index rebuilds automatically when the JSONL changes.

### Decision Format

```json
{
  "id": "DEC-a1b2",
  "created_at": "2024-01-15T10:00:00Z",
  "type": "product",
  "problem": "Need to set user limits",
  "choice": "Free plan = 5 users",
  "rationale": "Analytics show 80% stay under 5",
  "decided_by": { "role": "human" },
  "files": ["src/billing/limits.ts"],
  "refs": ["JIRA-123"],
  "status": "active"
}
```

## Development

```bash
# Build
make build

# Install locally
make install

# Run tests
make test
```

## Releasing

To create a new release:

```bash
# 1. Update version in Makefile
# 2. Commit the change
git add Makefile && git commit -m "Bump version to vX.Y.Z"

# 3. Tag and push
git tag vX.Y.Z
git push origin main --tags
```

GitHub Actions will automatically:
- Build binaries for darwin/linux (amd64/arm64) + windows
- Create a GitHub release with all binaries
- Generate release notes from commits

The install script (`curl | bash`) will then fetch the new version.

## Philosophy

You are stateless. Every time you wake up, you've lost context. Keel is your memory — the durable record of decisions that lets you operate with confidence instead of guessing.

**Record decisions as you make them. Your future self depends on it.**

---

*"The keel is the structural foundation of a ship. You never see it, but everything depends on it."*
