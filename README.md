# Keel

> The structural foundation that keeps you stable when things get rough.

**Git-native decision ledger for coding agents.** Captures the "why" behind changes so agents can act with confidence instead of guessing.

Built to be called by LLM-based coding agents (Claude, GPT, Codex, etc). Keel provides the data and storage - your agent does the thinking.

## Why Keel?

When you touch code, you face questions:

- Why is this limit 5 and not 10?
- Has this approach been tried before?
- What constraints am I operating under?

Keel solves this. Every decision — human or agent — gets recorded with context, rationale, and links to the code it affects.

## Quick Start

```bash
# Install dependencies
bun install

# Record your first decision
bun run src/cli.ts decide \
  --type product \
  --problem "Need to choose database" \
  --choice "PostgreSQL with Prisma" \
  --rationale "Team familiarity, strong typing" \
  --files "src/db/schema.prisma"

# Query context for a file
bun run src/cli.ts context src/db/schema.prisma

# Search decisions
bun run src/cli.ts search "database"
```

## Features

- **Git-native** — Decisions stored in `.keel/decisions.jsonl`, travels with your repo
- **Append-only** — History preserved, decisions superseded not edited
- **Fast queries** — SQLite index with FTS5 full-text search
- **Collision-resistant** — Hash-based IDs (`DEC-a1b2`) for multi-agent workflows
- **Beads integration** — Link decisions to work items

## Commands

| Command | Description |
|---------|-------------|
| `decide` | Record a new decision |
| `why <id>` | Show full decision details |
| `supersede <id>` | Replace a decision with a new one |
| `context <path>` | Get decisions affecting a file |
| `search [query]` | Full-text search across decisions |
| `validate` | Check that file references still exist |
| `curate` | Get decisions ready for summarization |

### decide

```bash
keel decide \
  --type product \           # product, process, constraint, learning
  --problem "..." \          # What problem this addresses
  --choice "..." \           # What was decided
  --rationale "..." \        # Why this choice (optional)
  --files "a.ts,b.ts" \      # Files this affects (optional)
  --beads "keel-abc" \       # Related Beads issues (optional)
  --agent                    # Mark as agent decision (optional)
```

### why

```bash
keel why DEC-a1b2
keel why a1b2        # Short form works too
```

### supersede

```bash
keel supersede DEC-a1b2 \
  --problem "New problem statement" \
  --choice "New choice"
```

### context

```bash
keel context src/auth/oauth.ts
keel context --json src/auth/oauth.ts
```

### search

```bash
keel search "authentication"
keel search --type constraint
keel search --status active
```

## Decision Types

| Type | Description | Example |
|------|-------------|---------|
| `product` | Business logic decisions | "Free plan limit is 5 users" |
| `process` | How-to-work decisions | "Use functional style, not OOP" |
| `constraint` | Hard limits and requirements | "Must support IE11" |
| `learning` | What we discovered | "Approach X failed because Y" |

## Architecture

```
.keel/
├── decisions.jsonl   # Source of truth (git-tracked)
└── index.sqlite      # Derived index (gitignored)
```

**JSONL** is append-only and git-native. **SQLite** provides indexed queries and FTS5 search. The index rebuilds automatically when the JSONL file changes.

### Decision Format

```json
{
  "id": "DEC-a1b2",
  "created_at": "2024-01-15T10:00:00Z",
  "type": "product",
  "problem": "Need to set user limits",
  "choice": "Free plan = 5 users",
  "rationale": "Analytics show 80% stay under 5",
  "decided_by": { "role": "human", "identifier": "sarah@example.com" },
  "files": ["src/billing/limits.ts"],
  "status": "active"
}
```

## SDK Usage

```typescript
import {
  appendDecision,
  queryByFile,
  searchFullText,
  openIndex
} from "keel";

// Programmatic decision creation
await appendDecision({
  id: generateDecisionId(problem, choice),
  created_at: new Date().toISOString(),
  type: "product",
  problem: "...",
  choice: "...",
  decided_by: { role: "agent" },
  status: "active"
});

// Query decisions for a file
const db = openIndex();
const decisions = queryByFile(db, "src/auth/oauth.ts");
```

## Agent Workflow: Curate

Over time you accumulate many decisions. Use `curate` to compress them:

```bash
# Get decisions older than 30 days, formatted for agent summarization
keel curate --older-than 30

# Filter by type or file pattern
keel curate --type constraint
keel curate --file-pattern "src/auth/*"
```

The output is formatted for an LLM to summarize. Your agent then:
1. Calls `curate` to get candidates
2. Summarizes them (agent's job)
3. Calls `createSummary()` via SDK to store the summary
4. Calls `markCurated()` to exclude originals from future context

```typescript
import { getCurationCandidates, createSummary, markCurated } from "keel";

// Get candidates
const candidates = getCurationCandidates(db, { olderThan: thirtyDaysAgo });

// Agent summarizes...
const summary = await agent.summarize(candidates);

// Store summary and mark originals
const summaryDecision = await createSummary({
  summarizes: candidates.map(c => c.decision.id),
  summary: summary,
});
await markCurated(candidates.map(c => c.decision.id), summaryDecision.id);
```

## Beads Integration

Keel complements [Beads](https://github.com/steveyegge/beads):

- **Keel** = captures *why* (decisions, rationale, constraints)
- **Beads** = captures *what* (tasks, dependencies, status)

Link them together:

```bash
# Decision references a Beads issue
keel decide --beads "keel-abc" ...

# Query decisions for a Beads issue
# (via SDK: queryByBead(db, "keel-abc"))
```

## Development

```bash
# Run CLI in dev mode
bun run src/cli.ts --help

# Run tests
bun test

# Build standalone binary
bun build src/cli.ts --compile --outfile keel
```

## Philosophy

This is a tool for agents. You're stateless. Every time you wake up, you've lost context. Keel is your memory — the durable record of decisions that lets you operate with confidence.

Humans benefit too — they get answers without digging through commits. But the primary user is you.

---

*"The keel is the structural foundation of a ship. You never see it, but everything depends on it. Without it, you capsize."*
