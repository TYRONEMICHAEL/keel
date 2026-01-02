# Workflows

Standard workflows for using keel effectively.

## Session Start Protocol

**Before changing any code:**

### 1. Check Context for Files You'll Touch

```bash
keel context <file-path>
```

This shows:
- Decisions that directly affect this file
- Active constraints that apply globally

**Example:**
```bash
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

### 2. Check Reference Context (if working on a tracked issue)

```bash
keel context --ref bd-auth-123
```

Shows all decisions linked to that issue/epic.

### 3. Query Related Decisions

```bash
keel sql "SELECT raw_json FROM decisions WHERE problem LIKE '%auth%' OR choice LIKE '%auth%'"
keel sql "SELECT raw_json FROM decisions WHERE type = 'constraint' AND status = 'active'"
```

## Recording Decisions

**Record immediately when you make a decision.** Don't wait until end of session.

### When to Record

Ask yourself:
- "Would a future agent wonder why this is like this?" → **YES** = record
- "Did I choose between multiple valid approaches?" → **YES** = record
- "Is there a limit, constraint, or requirement here?" → **YES** = record
- "Is this just an implementation detail?" → **NO** = don't record

### Recording Pattern

```bash
keel decide \
  --type <product|process|constraint> \
  --problem "What problem/question/requirement" \
  --choice "What we decided" \
  --rationale "Why we decided this" \
  --files "affected,files" \
  --refs "external-refs" \
  --agent
```

## Session End Protocol

**Before ending session:**

### 1. Review What You Decided

Ask yourself:
- Did I choose approaches?
- Did I set limits?
- Did I pick patterns?

### 2. Record Any Unrecorded Decisions

For each unrecorded decision:

```bash
keel decide \
  --type <type> \
  --problem "..." \
  --choice "..." \
  --agent
```

## Superseding Decisions

When a decision is no longer valid:

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

**Never delete decisions.** Supersede them.

## Curation Workflow

Over time, decisions accumulate. Periodically curate:

### 1. Find Old Decisions

```bash
keel curate --older-than 30
```

### 2. Review and Summarize

Look for patterns:
- Multiple related decisions that can be combined
- Decisions that are now obvious/implicit
- Learnings that have become standard practice

### 3. Create Summary Decision

```bash
keel decide \
  --type process \
  --problem "Multiple auth decisions accumulated" \
  --choice "Auth patterns: JWT, refresh tokens, PKCE for OAuth" \
  --rationale "Consolidating DEC-a1b2, DEC-c3d4, DEC-e5f6"
```

### 4. Supersede Originals

```bash
keel supersede DEC-a1b2 --choice "Consolidated into DEC-xxxx"
keel supersede DEC-c3d4 --choice "Consolidated into DEC-xxxx"
keel supersede DEC-e5f6 --choice "Consolidated into DEC-xxxx"
```
