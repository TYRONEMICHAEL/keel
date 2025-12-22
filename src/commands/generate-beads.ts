import { Command } from "commander";
import { openIndex, closeIndex } from "../core/db";
import { queryById } from "../core/query";
import { normalizeDecisionId } from "../core/id";

interface GenerateBeadsOptions {
  json?: boolean;
  dryRun?: boolean;
}

export const generateBeadsCommand = new Command("generate-beads")
  .description("Create Beads issues from a decision")
  .argument("<id>", "Decision ID to generate beads from")
  .option("--dry-run", "Show what would be created without creating")
  .option("--json", "Output as JSON")
  .action(async (rawId: string, options: GenerateBeadsOptions) => {
    try {
      const id = normalizeDecisionId(rawId);
      const db = openIndex();
      const decision = queryById(db, id);
      closeIndex(db);

      if (!decision) {
        console.error(`Decision ${id} not found`);
        process.exit(1);
      }

      // Generate issue title and description from decision
      const title = `[${decision.type}] ${decision.choice}`;
      const description = [
        `## Problem`,
        decision.problem,
        "",
        `## Decision`,
        decision.choice,
        decision.rationale ? `\n## Rationale\n${decision.rationale}` : "",
        "",
        `---`,
        `Generated from Keel decision ${decision.id}`,
      ].join("\n");

      if (options.dryRun) {
        console.log("Would create bead:");
        console.log(`  Title: ${title}`);
        console.log(`  Description:`);
        console.log(description.split("\n").map(l => `    ${l}`).join("\n"));
        return;
      }

      // Shell out to bd create
      const result = Bun.spawnSync([
        "bd", "create", title,
        "--description", description,
        "--label", decision.type,
      ]);

      if (result.exitCode !== 0) {
        console.error("Failed to create bead:");
        console.error(result.stderr.toString());
        process.exit(1);
      }

      const output = result.stdout.toString();

      if (options.json) {
        // Parse the bead ID from output
        const match = output.match(/Created issue: (\S+)/);
        const beadId = match ? match[1] : null;
        console.log(JSON.stringify({
          decisionId: decision.id,
          beadId,
          title,
        }, null, 2));
      } else {
        console.log(output);
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
