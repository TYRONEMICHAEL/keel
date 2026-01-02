---
name: keel
description: "Decision memory for AI agents. Before building or planning any feature, invoke this skill and check for existing decisions that might conflict. Decisions are constraints - if your plan conflicts, STOP and ask."
allowed-tools: "Read,Bash(keel:*)"
version: "0.1.0"
author: "Tyrone Avnit"
license: "MIT"
---

# Keel - Decision Memory for AI Agents

Git-native decision ledger that captures the "why" behind code changes. Provides persistent rationale that survives conversation compaction and enables confident decision-making.

## Overview

**keel** records decisions with full context: what problem was solved, what choice was made, why that choice, and what files it affects. Unlike comments (scattered, stale) or docs (separate, forgotten), keel keeps decisions queryable and linked to code.

**Core Capabilities**:
- 📝 **Decision Records**: Problem, choice, rationale, tradeoffs in one place
- 🔗 **Code Links**: Decisions linked to specific files and symbols
- 🔍 **Context Queries**: Ask "why is this file like this?" and get answers
- 📊 **Constraint Awareness**: Always know what hard limits apply
- 🐙 **Git Integration**: Decisions versioned in `.keel/decisions.jsonl`
- ⚡ **Fast Queries**: SQLite index with raw SQL access

**When to Record a Decision**:
- ❓ "Would a future agent wonder why this is like this?" → **YES** = record
- ❓ "Did I choose between multiple valid approaches?" → **YES** = record
- ❓ "Is there a limit, constraint, or requirement here?" → **YES** = record
- ❓ "Is this just an implementation detail?" → **NO** = don't record

**Decision Rule**: If you had to think about it, record it.

## Prerequisites

**Required**:
- **keel CLI**: Version 0.1.0 or later installed and in PATH
- **Git Repository**: Current directory must be a git repo
- **Initialization**: `keel init` must be run once (humans do this, not agents)

**Verify Installation**:
```bash
keel --version  # Should return 0.1.0 or later
```

**First-Time Setup** (humans run once):
```bash
cd /path/to/your/repo
keel init  # Creates .keel/ directory with empty ledger
```

## Instructions

### Session Start Protocol

**Every session, before changing code:**

#### Step 0: Check ALL Active Decisions (REQUIRED)

Before planning any feature or making changes, query all active decisions:

```bash
keel sql "SELECT id, type, problem, choice FROM decisions WHERE status = 'active'"
```

Scan for:
- **Constraints** (`type = 'constraint'`) - hard rules you MUST follow
- **Decisions that conflict** with your planned approach

**If your plan conflicts with an existing decision: STOP and inform the user.**

**If there are more than 50 active decisions:** Suggest running `keel curate --older-than 30` to consolidate old decisions into summaries.

#### Step 1: Check Context for Files You'll Touch

```bash
keel context <file-path>
```

Shows:
- Decisions that directly affect this file
- Active constraints that apply globally

**Example**:
```bash
keel context src/billing/limits.ts
```

**Output**:
```
Decisions affecting this file:

DEC-a1b2 [product] active
  Problem: Need to set free plan limits
  Choice: Free plan = 5 users

Active constraints:

  DEC-3957 Append-only ledger: decisions are never edited, only superseded
```

#### Step 2: Query Related Decisions

Use `keel sql` to query the SQLite index directly:

```bash
# Get all active decisions
keel sql "SELECT raw_json FROM decisions WHERE status = 'active'"

# Get all constraints (always relevant)
keel sql "SELECT raw_json FROM decisions WHERE type = 'constraint' AND status = 'active'"

# Search by content
keel sql "SELECT raw_json FROM decisions WHERE problem LIKE '%auth%' OR choice LIKE '%auth%'"
```

#### Step 3: Enforce Decisions as Constraints

**Decisions are CONSTRAINTS, not documentation.** If your proposed change would conflict with a recorded decision, you must:

1. **State the conflict explicitly**: "This would violate DEC-xxx which says..."
2. **STOP and ask for approval** before proceeding
3. **Do NOT rationalize compatibility** — if there's tension, surface it

You cannot supersede a decision on your own. Only a human can approve reversing a decision.

**Example conflict:**
```
I need to add a 6th user to the free plan, but DEC-a1b2 says
"Free plan = 5 users". This would violate that decision.

Should I:
1. Proceed and supersede DEC-a1b2?
2. Find another approach that respects the limit?
```

---

### Recording Decisions

**When you make a decision, record it immediately.** Don't wait until end of session.

#### Basic Decision

```bash
keel decide \
  --type product \
  --problem "Need to choose authentication method" \
  --choice "JWT with refresh tokens" \
  --rationale "Stateless, works with microservices, team has experience"
```

#### Decision with File Links

```bash
keel decide \
  --type constraint \
  --problem "API rate limits required for stability" \
  --choice "100 requests per minute per user" \
  --rationale "Based on load testing, prevents cascade failures" \
  --files "src/middleware/rateLimit.ts,src/config/limits.ts"
```

#### Decision with External References

Link decisions to external systems (Beads, Jira, GitHub Issues, etc.):

```bash
keel decide \
  --type product \
  --problem "Tried approach X for caching" \
  --choice "Abandoned in favor of Redis" \
  --rationale "In-memory cache caused OOM under load" \
  --refs "bd-auth-123,JIRA-456" \
  --agent
```

Query decisions by reference:

```bash
keel context --ref bd-auth-123
```

#### Agent-Made Decision

```bash
keel decide \
  --type process \
  --problem "Need consistent error handling" \
  --choice "Use Result type pattern, not exceptions" \
  --agent
```

#### Decision with Commit Reference (for Rollback)

Include the current commit hash to enable rollback to this decision point:

```bash
keel decide \
  --type product \
  --problem "Need to choose database" \
  --choice "PostgreSQL with Prisma" \
  --refs "commit:$(git rev-parse HEAD)" \
  --agent
```

This captures the exact code state when the decision was made.

**When to include commit refs:**
- Architectural decisions
- Significant code changes
- Decisions you might want to revisit

**When to skip:**
- Forward-looking decisions (before implementation)
- Non-code decisions

---

### Decision Types

| Type | When to Use | Example |
|------|-------------|---------|
| `product` | Business logic, features, architecture | "Free plan = 5 users" |
| `process` | How we work, patterns, style | "Use functional style, not OOP" |
| `constraint` | Hard limits, requirements, rules | "Must support IE11", "Max 100 RPS" |

**Choosing the Right Type**:
- If it affects what users see/do or system architecture → `product`
- If it affects how code is written or team works → `process`
- If it's a hard limit that can't be violated → `constraint`

---

### Superseding Decisions

When a decision is no longer valid, supersede it (don't delete):

```bash
keel supersede DEC-a1b2 \
  --problem "5 user limit causing churn" \
  --choice "Free plan = 10 users" \
  --rationale "Analytics show retention improves significantly"
```

This:
1. Marks old decision as `superseded`
2. Creates new decision with `supersedes: DEC-a1b2`
3. Preserves full history

---

### When Files Are Renamed or Moved

File paths in decisions are historical - they show where code was when the decision was made. When renaming or moving files:

1. **Check for affected decisions:**
```bash
keel sql "SELECT * FROM decisions WHERE raw_json LIKE '%<old-filename>%'"
```

2. **If critical decisions exist**, supersede with updated paths:
```bash
keel supersede DEC-xxx \
  --problem "Same problem" \
  --choice "Same choice" \
  --files "new/path/to/file.go"
```

3. **If just historical context**, leave as-is - the decision content is what matters.

**Note:** File refs are optional context, not the primary way to find decisions. Use `keel sql` to find decisions by content.

---

### Rolling Back to a Decision Point

To rollback to the code state when a decision was made:

**1. Query the decision:**
```bash
keel why DEC-xxx --json
```

**2. Extract commit ref from refs array:**
Look for ref starting with `commit:`, e.g., `"refs": ["commit:a1b2c3d4..."]`

**3. Check if decision is superseded:**
If `status: "superseded"`, warn the user before proceeding.

**4. Verify clean working tree:**
```bash
git status --porcelain
```
Stop if there are uncommitted changes.

**5. Checkout the commit:**
```bash
git checkout <commit-hash>
```

**6. Inform user:**
```
You are now in detached HEAD state at decision DEC-xxx.
To return to your branch: git checkout <branch-name>
```

**Note:** Rollback only works for decisions that included a commit ref. See "Decision with Commit Reference" above.

---

### Session End Protocol

Before ending session:

1. **Review what you decided**: Did you choose approaches, set limits, pick patterns?
2. **Record any unrecorded decisions**: Use `keel decide` for each

---

## Command Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `keel init` | Initialize keel (humans only) | `keel init` |
| `keel decide` | Record a new decision | `keel decide --type product --problem "..." --choice "..."` |
| `keel context <path>` | Get decisions for a file | `keel context src/auth/oauth.ts` |
| `keel context --ref <id>` | Get decisions for a reference | `keel context --ref bd-auth-123` |
| `keel why <id>` | Show full decision details | `keel why DEC-a1b2` |
| `keel sql <query>` | Execute SQL query | `keel sql "SELECT * FROM decisions WHERE status = 'active'"` |
| `keel supersede <id>` | Replace a decision | `keel supersede DEC-a1b2 --problem "..." --choice "..."` |
| `keel curate` | Get decisions for summarization | `keel curate --older-than 30` |
| `keel graph` | Output decision graph as Mermaid | `keel graph` |

---

## SQL Schema

The `keel sql` command queries a SQLite index. Schema:

```sql
-- Main decisions table
decisions (
  id TEXT PRIMARY KEY,       -- e.g., "DEC-3957"
  type TEXT,                  -- 'product', 'process', 'constraint'
  status TEXT,                -- 'active' = in effect, 'superseded' = replaced by newer decision
  problem TEXT,
  choice TEXT,
  rationale TEXT,
  created_at TEXT,
  supersedes TEXT,            -- ID of decision this supersedes
  superseded_by TEXT,         -- ID of decision that superseded this
  raw_json TEXT               -- Full decision as JSON
)

-- File associations
decision_files (decision_id, file_path)

-- Reference associations (Beads, Jira, commits, etc.)
decision_refs (decision_id, ref_id)

-- Symbol associations
decision_symbols (decision_id, symbol)
```

**Common Queries:**

```bash
# All active decisions
keel sql "SELECT raw_json FROM decisions WHERE status = 'active'"

# All constraints
keel sql "SELECT raw_json FROM decisions WHERE type = 'constraint' AND status = 'active'"

# Decisions mentioning a topic
keel sql "SELECT raw_json FROM decisions WHERE problem LIKE '%billing%' OR choice LIKE '%billing%'"

# Decisions for a file pattern
keel sql "SELECT d.raw_json FROM decisions d JOIN decision_files df ON d.id = df.decision_id WHERE df.file_path LIKE '%auth%'"
```

---

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| "Keel not initialized" | keel init not run | Run `keel init` (humans do this, not agents) |
| "Decision not found" | ID doesn't exist or typo | Use `keel sql` to find correct ID |
| "Not a git repository" | keel needs git context | Run from within a git repo |
| "No decisions found" | New repo or no matches | This is fine - start recording decisions |

---

## Worked Examples

### Example 1: Starting Work on a File

```bash
# Before editing src/billing/checkout.ts
$ keel context src/billing/checkout.ts

Decisions affecting this file:

DEC-7f2a [product] active
  Problem: Need to handle failed payments
  Choice: Retry 3 times with exponential backoff

DEC-9b1c [constraint] active
  Problem: PCI compliance requirements
  Choice: Never log full card numbers

Active constraints:
  DEC-3957 Append-only ledger
```

Now you know: retry logic exists (don't duplicate), never log card numbers.

### Example 2: Making an Architectural Choice

```bash
# You chose WebSockets over polling
$ keel decide \
  --type product \
  --problem "Need real-time updates for dashboard" \
  --choice "WebSockets with Socket.io" \
  --rationale "Lower latency than polling, better UX, handles reconnection" \
  --files "src/realtime/socket.ts,src/client/hooks/useSocket.ts"

Created DEC-4e8f
```

### Example 3: Recording an Architectural Choice

```bash
# Chose SSE over WebSockets
$ keel decide \
  --type product \
  --problem "Need real-time notifications" \
  --choice "Server-Sent Events over WebSockets" \
  --rationale "Simpler, unidirectional sufficient, better browser support"

Created DEC-2a9c
```

---

## Codebase Onboarding

When joining an existing codebase, use the onboarding protocol to extract implicit decisions through interviews.

**How it works:**
1. Analyze codebase structure
2. Check existing decisions: `keel sql "SELECT * FROM decisions WHERE type = 'product'"`
3. Interview user about the "why" behind choices
4. Group related answers into coherent decisions
5. Batch record with `keel decide`

**Resumability:** Decisions are the state. On resume, query what's documented and ask user what else to cover.

**See:** `references/ONBOARD.md` for the full protocol.

---

## Philosophy

You are stateless. Every time you wake up, you've lost context. Keel is your memory - the durable record of decisions that lets you operate with confidence instead of guessing.

Without recorded decisions, you'll:
- Repeat mistakes that were already learned
- Violate constraints you didn't know existed
- Redo work that was already decided against
- Make choices that conflict with previous decisions

**Record decisions as you make them. Your future self depends on it.**
