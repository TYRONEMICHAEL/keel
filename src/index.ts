// Types
export type {
  Decision,
  DecisionInput,
  DecisionType,
  DecisionStatus,
  DecidedBy,
} from "./core/types";

export {
  DecisionSchema,
  DecisionInputSchema,
  validateDecision,
  validateDecisionInput,
} from "./core/types";

// Store operations
export {
  appendDecision,
  readAllDecisions,
  getLatestState,
  getDecisionById,
  getActiveDecisions,
  getKeelDir,
  getDecisionsPath,
  ensureKeelDir,
} from "./core/store";

// Database operations
export { openIndex, indexDecision, closeIndex } from "./core/db";

// ID utilities
export {
  generateDecisionId,
  isValidDecisionId,
  normalizeDecisionId,
} from "./core/id";

// Query operations
export {
  queryById,
  queryByFile,
  queryBySymbol,
  queryByBead,
  queryAll,
  searchFullText,
  getActiveConstraints,
  getDecisionsForContext,
} from "./core/query";

// Formatting utilities
export {
  formatDecisionSummary,
  formatDecisionFull,
  formatDecisionList,
  formatContextResult,
  formatJson,
} from "./utils/format";

// Git utilities
export {
  findRepoRoot,
  getGitUser,
  getGitIdentifier,
  normalizeFilePath,
} from "./utils/git";

// Validation
export { validateCommand } from "./commands/validate";

// Stubs for future features
export async function curate(): Promise<void> {
  throw new Error("curate: Not yet implemented. Will allow bulk decision management.");
}

export async function generateBeads(_decisionId: string): Promise<string[]> {
  throw new Error("generateBeads: Not yet implemented. Will extract key concepts from decisions.");
}
