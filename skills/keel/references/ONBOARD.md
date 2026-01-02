# Codebase Onboarding

Extract architectural decisions from an existing codebase through structured interviews.

## Overview

When joining an existing codebase, decisions exist implicitly in the code but aren't documented. This protocol helps extract and record them through conversation with the team.

## How It Works

1. **Analyze** the codebase structure
2. **Interview** the user deeply about the "why" behind choices
3. **Group** related answers into coherent decisions
4. **Record** decisions in batch with proper relationships

## Resumability

Keel doesn't track interview progress—just decisions. On resume:

1. Query existing decisions: `keel sql "SELECT * FROM decisions WHERE status = 'active'"`
2. Show user what's documented
3. Ask: "What else should we cover?"
4. User guides the remaining interview

This is simpler than tracking coverage and puts the human in control.

---

## The Onboarding Prompt

Use this prompt with any agent. Adapt based on whether structured question tools are available.

```markdown
# Codebase Onboarding Interview

Extract architectural decisions from this codebase through conversation.

## Phase 1: Codebase Analysis

Explore the codebase to understand its structure:
- Major directories and their purposes
- Key configuration (package.json, go.mod, Dockerfile, etc.)
- Apparent patterns (monorepo, microservices, etc.)
- Core technologies and frameworks

Build a mental map of areas that likely have decisions:
- Architecture & structure
- Data model & storage
- Authentication & authorization
- API design
- Error handling
- Testing approach
- Deployment & infrastructure
- Performance considerations
- Security measures
- Third-party integrations

## Phase 2: Check Existing Decisions

Before interviewing, see what's already documented:

```bash
keel sql "SELECT * FROM decisions WHERE type = 'product' AND status = 'active'"
keel sql "SELECT * FROM decisions WHERE type = 'constraint' AND status = 'active'"
keel sql "SELECT * FROM decisions WHERE type = 'process' AND status = 'active'"
```

Show the user:
- "I found X decisions already documented"
- "They cover: [areas]"
- "What areas should we discuss?"

## Phase 3: Interview

Conduct a deep interview about undocumented areas.

### Interview Guidelines

**Ask about WHY, not WHAT** (code shows what, you need why):
- "I see you're using PostgreSQL. What alternatives did you consider?"
- "The auth uses custom session handling instead of [framework default]. Why?"
- "There's no caching layer. Deliberate choice or not needed yet?"

**Dig for tradeoffs**:
- "What did you give up by choosing X?"
- "What would break if someone changed this?"

**Probe rejected alternatives**:
- "What else did you consider?"
- "Why not [obvious alternative]?"

**Find constraints**:
- "Are there hard limits we can't violate?"
- "What would compliance/security/legal prohibit?"

**Don't ask obvious questions** the code already answers.

### If you have AskUserQuestionTool

Use it for structured choices:
- "Which auth provider do you use?" [Google, GitHub, Auth0, Other]
- "Is this limit negotiable?" [Hard constraint, Soft preference]

### If no structured tool

Interview conversationally. Ask open questions, follow up on interesting answers.

### Keep Going Until Complete

Interview iteratively:
- Cover one area at a time
- Ask follow-up questions on interesting answers
- Move to next area when current one is exhausted
- Stop when user says they're done or all areas covered

## Phase 4: Group & Record

After completing the interview (or when user needs to stop):

### 1. Group Related Answers

Don't create one decision per answer. Group related items:
- Multiple auth answers → one auth decision
- Several API design points → one API conventions decision

### 2. Identify Relationships

Note which decisions:
- Depend on each other
- Conflict with each other (resolve with user)
- Are constraints vs preferences

### 3. Determine Types

- `product` — Choices about system architecture, features, behavior
- `process` — How the team works, conventions, patterns
- `constraint` — Hard limits that cannot be violated

### 4. Batch Record

Record all decisions together:

```bash
# Core architecture decision
keel decide --type product \
  --problem "Needed to handle 10k concurrent connections" \
  --choice "Event-driven architecture with message queue" \
  --rationale "Request-response couldn't scale; team had RabbitMQ experience"

# Related constraint
keel decide --type constraint \
  --problem "Message ordering requirements" \
  --choice "Single consumer per queue partition" \
  --rationale "Business logic requires ordered processing"

# Team convention
keel decide --type process \
  --problem "Need consistent API response format" \
  --choice "All endpoints return {data, error, meta} envelope" \
  --rationale "Makes client parsing uniform, supports pagination metadata"
```

## Phase 5: Session Summary

After recording, summarize for the user:

```
Recorded X decisions:
- Architecture: Event-driven with RabbitMQ
- Auth: OAuth2 with PKCE, Google/GitHub providers
- API: RESTful with envelope pattern
- Constraint: Single consumer per queue partition

Areas not yet covered:
- Deployment & infrastructure
- Testing strategy

To continue later, just run onboarding again.
```

---

## Handling Large Codebases

### Prioritization

For large codebases, don't try to cover everything:
1. Start with core business logic
2. Focus on areas with most risk/complexity
3. Cover constraints first (they block other decisions)
4. Accept that full coverage may take multiple sessions

### Session Boundaries

When user needs to stop:
1. Record all decisions gathered so far
2. Summarize what was covered
3. Note what's remaining
4. On resume, user tells you what to continue with

---

## Anti-Patterns

❌ Recording every answer as a separate decision
❌ Asking questions the code clearly answers
❌ Recording obvious/standard choices
❌ Skipping the "why" (just recording "what")
❌ Asking all questions before recording anything (lose context if interrupted)

## Good Patterns

✓ Group related decisions before recording
✓ Always capture rationale and tradeoffs
✓ Record after each major area (not all at end)
✓ Ask follow-up questions to get to the real "why"
✓ Note when user is uncertain (that's valuable info)
✓ Let user guide priority and session boundaries
