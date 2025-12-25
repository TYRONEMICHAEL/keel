---
name: keel
description: >
  Records decisions with rationale so agents can understand why code is the way it is. Use when making
  architectural choices, setting limits, choosing approaches, or learning from failures. Trigger with
  phrases like "we decided", "the reason is", "why is this", "what constraints", or "record decision".
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
- ⚡ **Fast Search**: SQLite index with FTS5 full-text search

**When to Record a Decision**:
- ❓ "Would a future agent wonder why this is like this?" → **YES** = record
- ❓ "Did I choose between multiple valid approaches?" → **YES** = record
- ❓ "Is there a limit, constraint, or requirement here?" → **YES** = record
- ❓ "Did something fail that we shouldn't try again?" → **YES** = record
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

#### Step 2: Search for Related Decisions

```bash
keel search "authentication"
keel search --type constraint
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

#### Agent-Made Decision

```bash
keel decide \
  --type process \
  --problem "Need consistent error handling" \
  --choice "Use Result type pattern, not exceptions" \
  --agent
```

---

### Decision Types

| Type | When to Use | Example |
|------|-------------|---------|
| `product` | Business logic, features, limits | "Free plan = 5 users" |
| `process` | How we work, patterns, style | "Use functional style, not OOP" |
| `constraint` | Hard limits, requirements, rules | "Must support IE11", "Max 100 RPS" |
| `learning` | Failed approaches, discoveries | "Approach X failed because Y" |

**Choosing the Right Type**:
- If it affects what users see/do → `product`
- If it affects how code is written → `process`
- If it's a hard limit that can't be violated → `constraint`
- If it's knowledge gained from trying something → `learning`

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
| `keel why <id>` | Show full decision details | `keel why DEC-a1b2` |
| `keel search <query>` | Full-text search | `keel search "authentication"` |
| `keel search --type <type>` | Filter by type | `keel search --type constraint` |
| `keel supersede <id>` | Replace a decision | `keel supersede DEC-a1b2 --problem "..." --choice "..."` |
| `keel validate` | Check file references exist | `keel validate` |
| `keel curate` | Get decisions for summarization | `keel curate --older-than 30` |

---

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| "Keel not initialized" | keel init not run | Run `keel init` (humans do this, not agents) |
| "Decision not found" | ID doesn't exist or typo | Use `keel search` to find correct ID |
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

### Example 3: Recording a Failed Approach

```bash
# GraphQL subscriptions didn't work out
$ keel decide \
  --type learning \
  --problem "Tried GraphQL subscriptions for real-time" \
  --choice "Abandoned - too complex for our needs" \
  --rationale "Required Apollo Server, added 50KB to bundle, team unfamiliar"

Created DEC-2a9c
```

Now future agents won't waste time trying GraphQL subscriptions.

---

## Philosophy

You are stateless. Every time you wake up, you've lost context. Keel is your memory - the durable record of decisions that lets you operate with confidence instead of guessing.

Without recorded decisions, you'll:
- Repeat mistakes that were already learned
- Violate constraints you didn't know existed
- Redo work that was already decided against
- Make choices that conflict with previous decisions

**Record decisions as you make them. Your future self depends on it.**
