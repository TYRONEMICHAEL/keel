package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/id"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/query"
	"github.com/tyroneavnit/keel/internal/store"
	"github.com/tyroneavnit/keel/internal/types"
)

var supersedeCmd = &cobra.Command{
	Use:   "supersede <id>",
	Short: "Replace a decision with a new one",
	Long:  `Create a new decision that supersedes an existing one. The old decision is marked as superseded.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSupersede,
}

var (
	supersedeProblem   string
	supersedeChoice    string
	supersedeRationale string
	supersedeFiles     string
	supersedeRefs      string
	supersedeAgent     bool
)

func init() {
	supersedeCmd.Flags().StringVar(&supersedeProblem, "problem", "", "New problem statement (defaults to original)")
	supersedeCmd.Flags().StringVar(&supersedeChoice, "choice", "", "New choice (required)")
	supersedeCmd.Flags().StringVar(&supersedeRationale, "rationale", "", "Why this supersedes the original")
	supersedeCmd.Flags().StringVar(&supersedeFiles, "files", "", "Comma-separated list of affected files")
	supersedeCmd.Flags().StringVar(&supersedeRefs, "refs", "", "Comma-separated list of external references (issues, epics, etc.)")
	supersedeCmd.Flags().BoolVar(&supersedeAgent, "agent", false, "Mark as an agent decision")

	supersedeCmd.MarkFlagRequired("choice")

	rootCmd.AddCommand(supersedeCmd)
}

func runSupersede(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()

	// Check initialization
	if err := store.RequireInit(repoRoot); err != nil {
		return err
	}

	rawID := args[0]
	normalizedID, err := id.Normalize(rawID)
	if err != nil {
		return err
	}
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	// Get original decision
	original, err := query.ByID(db, normalizedID)
	if err != nil {
		return err
	}
	if original == nil {
		return fmt.Errorf("decision %s not found", normalizedID)
	}

	// Build new decision input
	problem := original.Problem
	if supersedeProblem != "" {
		problem = supersedeProblem
	}

	input := types.DecisionInput{
		Type:       original.Type,
		Problem:    problem,
		Choice:     supersedeChoice,
		Supersedes: &normalizedID,
	}

	if supersedeRationale != "" {
		input.Rationale = &supersedeRationale
	}

	if supersedeFiles != "" {
		input.Files = splitAndTrim(supersedeFiles)
	} else {
		input.Files = original.Files
	}

	if supersedeRefs != "" {
		input.Refs = splitAndTrim(supersedeRefs)
	} else {
		input.Refs = original.Refs
	}

	role := "human"
	if supersedeAgent {
		role = "agent"
	}
	input.DecidedBy = &types.DecidedBy{Role: role}

	// Generate new ID
	newID := id.Generate(input.Problem, input.Choice)

	// Create new decision
	newDecision := types.NewDecision(newID, input)

	// Save new decision
	if err := store.AppendDecision(newDecision, repoRoot); err != nil {
		return fmt.Errorf("failed to save decision: %w", err)
	}
	if err := db.IndexDecision(newDecision); err != nil {
		return fmt.Errorf("failed to index decision: %w", err)
	}

	// Mark original as superseded
	original.Status = types.StatusSuperseded
	original.SupersededBy = &newID
	if err := store.AppendDecision(original, repoRoot); err != nil {
		return fmt.Errorf("failed to update original decision: %w", err)
	}
	if err := db.IndexDecision(original); err != nil {
		return fmt.Errorf("failed to index original decision: %w", err)
	}

	fmt.Printf("Created \033[1m%s\033[0m (supersedes %s)\n", newID, normalizedID)
	return nil
}
