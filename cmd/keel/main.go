package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "keel",
	Short: "Git-native decision ledger CLI",
	Long: `Keel captures the "why" behind changes so agents can act with confidence instead of guessing.

Built to be called by LLM-based coding agents (Claude, GPT, Codex, etc).
Keel provides the data and storage - your agent does the thinking.`,
	Version: version,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
