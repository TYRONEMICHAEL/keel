import type { Database } from "bun:sqlite";
import type { Decision, DecisionType } from "./types";
import { queryAll } from "./query";
import { appendDecision } from "./store";
import { generateDecisionId } from "./id";
import { openIndex, indexDecision, closeIndex } from "./db";

export interface CurationOptions {
  olderThan?: Date;
  type?: DecisionType;
  filePattern?: string;
  excludeCurated?: boolean;
}

export interface CurationCandidate {
  decision: Decision;
  age: number; // days
  relatedCount: number; // other decisions affecting same files
}

/**
 * Get decisions that are candidates for curation/summarization.
 * The calling agent can then summarize these and call createSummary.
 */
export function getCurationCandidates(
  db: Database,
  options: CurationOptions = {}
): CurationCandidate[] {
  const allDecisions = queryAll(db, { status: "active" });
  const now = Date.now();

  // Filter by options
  let candidates = allDecisions.filter((d) => {
    // Exclude summaries (decisions that summarize others)
    if ((d as any).summarizes?.length) {
      return false;
    }

    // Exclude already curated
    if (options.excludeCurated && (d as any).curated_into) {
      return false;
    }

    // Filter by age
    if (options.olderThan) {
      const createdAt = new Date(d.created_at).getTime();
      if (createdAt > options.olderThan.getTime()) {
        return false;
      }
    }

    // Filter by type
    if (options.type && d.type !== options.type) {
      return false;
    }

    // Filter by file pattern
    if (options.filePattern && d.files?.length) {
      const pattern = options.filePattern.replace(/\*/g, ".*");
      const regex = new RegExp(pattern);
      if (!d.files.some((f) => regex.test(f))) {
        return false;
      }
    }

    return true;
  });

  // Build file -> decision count map for related count
  const fileDecisionCount = new Map<string, number>();
  for (const d of allDecisions) {
    for (const file of d.files ?? []) {
      fileDecisionCount.set(file, (fileDecisionCount.get(file) ?? 0) + 1);
    }
  }

  // Map to candidates with metadata
  return candidates.map((decision) => {
    const createdAt = new Date(decision.created_at).getTime();
    const age = Math.floor((now - createdAt) / (1000 * 60 * 60 * 24));

    let relatedCount = 0;
    for (const file of decision.files ?? []) {
      relatedCount += (fileDecisionCount.get(file) ?? 1) - 1;
    }

    return { decision, age, relatedCount };
  });
}

/**
 * Group curation candidates by file patterns for easier summarization.
 */
export function groupByFilePattern(
  candidates: CurationCandidate[]
): Map<string, CurationCandidate[]> {
  const groups = new Map<string, CurationCandidate[]>();

  for (const candidate of candidates) {
    const files = candidate.decision.files ?? [];
    if (files.length === 0) {
      const key = "(no files)";
      if (!groups.has(key)) groups.set(key, []);
      groups.get(key)!.push(candidate);
    } else {
      // Use first directory as grouping key
      for (const file of files) {
        const dir = file.split("/").slice(0, 2).join("/") || file;
        if (!groups.has(dir)) groups.set(dir, []);
        groups.get(dir)!.push(candidate);
      }
    }
  }

  return groups;
}

export interface SummaryInput {
  summarizes: string[]; // IDs of decisions being summarized
  summary: string; // Agent-generated summary text
  title?: string; // Optional short title
}

/**
 * Create a summary decision that references the original decisions.
 * This is called by the agent after it has summarized the candidates.
 */
export async function createSummary(
  input: SummaryInput,
  repoRoot: string = process.cwd()
): Promise<Decision> {
  const id = generateDecisionId(input.summary, input.summarizes.join(","));

  const decision: Decision = {
    id,
    created_at: new Date().toISOString(),
    type: "learning", // Summaries are learnings
    problem: input.title ?? `Summary of ${input.summarizes.length} decisions`,
    choice: input.summary,
    decided_by: { role: "agent", identifier: "curate" },
    status: "active",
    // Store the IDs being summarized (extends Decision type)
    ...(({ summarizes: input.summarizes } as any)),
  };

  await appendDecision(decision, repoRoot);

  const db = openIndex(repoRoot);
  indexDecision(db, decision);
  closeIndex(db);

  return decision;
}

/**
 * Mark decisions as curated into a summary.
 * This allows excluding them from future context loads.
 */
export async function markCurated(
  ids: string[],
  summaryId: string,
  repoRoot: string = process.cwd()
): Promise<void> {
  const db = openIndex(repoRoot);

  for (const id of ids) {
    // Append an update marking the decision as curated
    const update = {
      id,
      curated_into: summaryId,
    };
    await appendDecision(update as Decision, repoRoot);
  }

  closeIndex(db);
}

/**
 * Format candidates for agent consumption.
 * Returns a string that's easy for an LLM to process.
 */
export function formatForAgent(candidates: CurationCandidate[]): string {
  const lines: string[] = [];

  lines.push(`# Decisions to Summarize (${candidates.length})`);
  lines.push("");

  for (const { decision, age } of candidates) {
    lines.push(`## ${decision.id} [${decision.type}] (${age} days old)`);
    lines.push(`**Problem:** ${decision.problem}`);
    lines.push(`**Choice:** ${decision.choice}`);
    if (decision.rationale) {
      lines.push(`**Rationale:** ${decision.rationale}`);
    }
    if (decision.files?.length) {
      lines.push(`**Files:** ${decision.files.join(", ")}`);
    }
    lines.push("");
  }

  lines.push("---");
  lines.push("Summarize these decisions into a concise playbook.");
  lines.push("Group by theme. Preserve key constraints and learnings.");

  return lines.join("\n");
}
