import { Command } from "commander";
import { openIndex, closeIndex } from "../core/db";
import {
  getCurationCandidates,
  groupByFilePattern,
  formatForAgent,
} from "../core/curate";
import type { DecisionType } from "../core/types";

interface CurateOptions {
  olderThan?: string;
  type?: DecisionType;
  filePattern?: string;
  json?: boolean;
  grouped?: boolean;
}

export const curateCommand = new Command("curate")
  .description("Get decisions ready for summarization by an agent")
  .option("--older-than <days>", "Only include decisions older than N days")
  .option("-t, --type <type>", "Filter by type: product, process, constraint, learning")
  .option("-f, --file-pattern <pattern>", "Filter by file pattern (e.g., 'src/auth/*')")
  .option("--grouped", "Group candidates by file directory")
  .option("--json", "Output as JSON")
  .action(async (options: CurateOptions) => {
    try {
      const db = openIndex();

      const olderThan = options.olderThan
        ? new Date(Date.now() - parseInt(options.olderThan) * 24 * 60 * 60 * 1000)
        : undefined;

      const candidates = getCurationCandidates(db, {
        olderThan,
        type: options.type,
        filePattern: options.filePattern,
        excludeCurated: true,
      });

      closeIndex(db);

      if (candidates.length === 0) {
        console.log("No decisions found matching criteria.");
        return;
      }

      if (options.json) {
        if (options.grouped) {
          const groups = groupByFilePattern(candidates);
          const output: Record<string, any[]> = {};
          for (const [key, items] of groups) {
            output[key] = items.map((c) => ({
              ...c.decision,
              _age: c.age,
              _relatedCount: c.relatedCount,
            }));
          }
          console.log(JSON.stringify(output, null, 2));
        } else {
          console.log(
            JSON.stringify(
              candidates.map((c) => ({
                ...c.decision,
                _age: c.age,
                _relatedCount: c.relatedCount,
              })),
              null,
              2
            )
          );
        }
      } else {
        // Output in agent-friendly format
        console.log(formatForAgent(candidates));
      }
    } catch (error) {
      if (error instanceof Error) {
        console.error(`Error: ${error.message}`);
      } else {
        console.error("An unexpected error occurred");
      }
      process.exit(1);
    }
  });
