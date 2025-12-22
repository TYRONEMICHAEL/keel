import { Command } from "commander";
import { existsSync } from "node:fs";
import { join } from "node:path";
import { openIndex, closeIndex } from "../core/db";
import { queryAll } from "../core/query";
import { findRepoRoot } from "../utils/git";

interface ValidationIssue {
  decisionId: string;
  type: "missing_file" | "missing_symbol";
  reference: string;
  message: string;
}

interface ValidationResult {
  valid: boolean;
  issues: ValidationIssue[];
  stats: {
    decisionsChecked: number;
    filesChecked: number;
    missingFiles: number;
  };
}

interface ValidateOptions {
  json?: boolean;
  fix?: boolean;
}

export const validateCommand = new Command("validate")
  .description("Check that file/symbol references in decisions still exist")
  .option("--json", "Output as JSON")
  .option("--fix", "Suggest fixes for broken references (not implemented)")
  .action(async (options: ValidateOptions) => {
    try {
      const repoRoot = findRepoRoot() ?? process.cwd();
      const db = openIndex();
      const decisions = queryAll(db, { status: "active" });
      closeIndex(db);

      const issues: ValidationIssue[] = [];
      const filesChecked = new Set<string>();

      for (const decision of decisions) {
        if (decision.files?.length) {
          for (const file of decision.files) {
            filesChecked.add(file);
            const fullPath = join(repoRoot, file);
            if (!existsSync(fullPath)) {
              issues.push({
                decisionId: decision.id,
                type: "missing_file",
                reference: file,
                message: `File not found: ${file}`,
              });
            }
          }
        }
      }

      const result: ValidationResult = {
        valid: issues.length === 0,
        issues,
        stats: {
          decisionsChecked: decisions.length,
          filesChecked: filesChecked.size,
          missingFiles: issues.filter((i) => i.type === "missing_file").length,
        },
      };

      if (options.json) {
        console.log(JSON.stringify(result, null, 2));
      } else {
        console.log(`Checked ${result.stats.decisionsChecked} decisions, ${result.stats.filesChecked} file references`);
        console.log("");

        if (result.valid) {
          console.log("\x1b[32m✓ All references valid\x1b[0m");
        } else {
          console.log(`\x1b[31m✖ Found ${issues.length} issue(s):\x1b[0m`);
          console.log("");

          for (const issue of issues) {
            console.log(`  ${issue.decisionId}: ${issue.message}`);
          }

          if (options.fix) {
            console.log("");
            console.log("\x1b[33m--fix is not yet implemented\x1b[0m");
          }
        }
      }

      process.exit(result.valid ? 0 : 1);
    } catch (error) {
      if (error instanceof Error) {
        console.error(`Error: ${error.message}`);
      } else {
        console.error("An unexpected error occurred");
      }
      process.exit(1);
    }
  });
