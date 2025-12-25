package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/id"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/store"
	"github.com/tyroneavnit/keel/internal/types"
)

var decideCmd = &cobra.Command{
	Use:   "decide",
	Short: "Record a new decision",
	Long: `Record a new decision in the ledger.

Decision types:
  product    - Business logic decisions (e.g., "Free plan = 5 users")
  process    - How-to-work decisions (e.g., "Use functional style")
  constraint - Hard limits and requirements (e.g., "Must support IE11")
  learning   - What we discovered (e.g., "Approach X failed because Y")`,
	RunE: runDecide,
}

var (
	decideType       string
	decideProblem    string
	decideChoice     string
	decideRationale  string
	decideFiles      string
	decideSymbols    string
	decideRefs       string
	decideAgent      bool
	decideSupersedes string
)

func init() {
	decideCmd.Flags().StringVarP(&decideType, "type", "t", "", "Decision type: product, process, constraint, learning (required)")
	decideCmd.Flags().StringVar(&decideProblem, "problem", "", "What problem this addresses (required)")
	decideCmd.Flags().StringVar(&decideChoice, "choice", "", "What was decided (required)")
	decideCmd.Flags().StringVar(&decideRationale, "rationale", "", "Why this choice was made")
	decideCmd.Flags().StringVar(&decideFiles, "files", "", "Comma-separated list of affected files")
	decideCmd.Flags().StringVar(&decideSymbols, "symbols", "", "Comma-separated list of affected symbols")
	decideCmd.Flags().StringVar(&decideRefs, "refs", "", "Comma-separated list of external references (issues, epics, etc.)")
	decideCmd.Flags().BoolVar(&decideAgent, "agent", false, "Mark as an agent decision")
	decideCmd.Flags().StringVar(&decideSupersedes, "supersedes", "", "ID of decision this supersedes")

	decideCmd.MarkFlagRequired("type")
	decideCmd.MarkFlagRequired("problem")
	decideCmd.MarkFlagRequired("choice")

	rootCmd.AddCommand(decideCmd)
}

func runDecide(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()

	// Check initialization
	if err := store.RequireInit(repoRoot); err != nil {
		return err
	}

	// Validate type
	if !types.IsValidType(decideType) {
		return fmt.Errorf("invalid type: %s. Must be one of: product, process, constraint, learning", decideType)
	}

	// Build input
	input := types.DecisionInput{
		Type:    types.DecisionType(decideType),
		Problem: decideProblem,
		Choice:  decideChoice,
	}

	if decideRationale != "" {
		input.Rationale = &decideRationale
	}

	if decideFiles != "" {
		input.Files = splitAndTrim(decideFiles)
	}

	if decideSymbols != "" {
		input.Symbols = splitAndTrim(decideSymbols)
	}

	if decideRefs != "" {
		input.Refs = splitAndTrim(decideRefs)
	}

	if decideSupersedes != "" {
		normalized, err := id.Normalize(decideSupersedes)
		if err != nil {
			return err
		}
		input.Supersedes = &normalized
	}

	// Set decided_by
	role := "human"
	if decideAgent {
		role = "agent"
	}
	input.DecidedBy = &types.DecidedBy{Role: role}

	// Generate ID
	decisionID := id.Generate(input.Problem, input.Choice)

	// Create decision
	decision := types.NewDecision(decisionID, input)

	// Append to JSONL
	if err := store.AppendDecision(decision, repoRoot); err != nil {
		return fmt.Errorf("failed to save decision: %w", err)
	}

	// Update index
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	if err := db.IndexDecision(decision); err != nil {
		return fmt.Errorf("failed to index decision: %w", err)
	}

	// Handle supersedes
	if input.Supersedes != nil {
		// Mark old decision as superseded
		oldDecision, err := store.GetDecisionByID(*input.Supersedes, repoRoot)
		if err != nil {
			return fmt.Errorf("failed to get superseded decision: %w", err)
		}
		if oldDecision != nil {
			oldDecision.Status = types.StatusSuperseded
			oldDecision.SupersededBy = &decisionID
			if err := store.AppendDecision(oldDecision, repoRoot); err != nil {
				return fmt.Errorf("failed to update superseded decision: %w", err)
			}
			if err := db.IndexDecision(oldDecision); err != nil {
				return fmt.Errorf("failed to index superseded decision: %w", err)
			}
		}
	}

	fmt.Printf("Created \033[1m%s\033[0m\n", decisionID)
	return nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
