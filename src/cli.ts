#!/usr/bin/env bun

import { Command } from "commander";
import { decideCommand } from "./commands/decide";
import { whyCommand } from "./commands/why";
import { supersedeCommand } from "./commands/supersede";
import { contextCommand } from "./commands/context";
import { searchCommand } from "./commands/search";
import { validateCommand } from "./commands/validate";

const program = new Command();

program
  .name("keel")
  .description("Git-native decision ledger CLI")
  .version("0.1.0");

program.addCommand(decideCommand);
program.addCommand(whyCommand);
program.addCommand(supersedeCommand);
program.addCommand(contextCommand);
program.addCommand(searchCommand);
program.addCommand(validateCommand);

program.parse();
