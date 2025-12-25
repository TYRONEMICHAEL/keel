package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/store"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Keel in the current repository",
	Long: `Initialize Keel decision tracking in the current git repository.

This creates the .keel/ directory and sets up the decision ledger.
This command should be run once by a human, not by agents.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	repoRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	// Check if already initialized
	keelDir := store.GetKeelDir(repoRoot)
	if _, err := os.Stat(keelDir); err == nil {
		fmt.Println("Keel is already initialized in this repository.")
		return nil
	}

	// Check if this is a git repo
	gitDir := filepath.Join(repoRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository. Run 'git init' first")
	}

	// Create .keel directory
	if err := store.EnsureKeelDir(repoRoot); err != nil {
		return fmt.Errorf("failed to create .keel directory: %w", err)
	}

	// Create empty decisions.jsonl
	decisionsPath := store.GetDecisionsPath(repoRoot)
	f, err := os.Create(decisionsPath)
	if err != nil {
		return fmt.Errorf("failed to create decisions file: %w", err)
	}
	f.Close()

	fmt.Println("\033[32m✓ Keel initialized\033[0m")
	fmt.Println()
	fmt.Println("Created .keel/ directory with empty decision ledger.")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add '.keel/index.sqlite' to .gitignore")
	fmt.Println("  2. Commit '.keel/decisions.jsonl' to git")
	fmt.Println("  3. Record your first decision: keel decide --type product ...")

	return nil
}
