package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/query"
	"github.com/tyroneavnit/keel/internal/types"
)

var curateCmd = &cobra.Command{
	Use:   "curate",
	Short: "Get decisions ready for summarization by an agent",
	Long:  `List decisions that are candidates for summarization, filtered by age, type, or file pattern.`,
	RunE:  runCurate,
}

var (
	curateOlderThan   int
	curateType        string
	curateFilePattern string
	curateJSON        bool
)

func init() {
	curateCmd.Flags().IntVar(&curateOlderThan, "older-than", 0, "Only include decisions older than N days")
	curateCmd.Flags().StringVarP(&curateType, "type", "t", "", "Filter by type: product, process, constraint")
	curateCmd.Flags().StringVarP(&curateFilePattern, "file-pattern", "f", "", "Filter by file pattern (e.g., 'src/auth/*')")
	curateCmd.Flags().BoolVar(&curateJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(curateCmd)
}

type CurationCandidate struct {
	Decision     *types.Decision `json:"decision"`
	Age          int             `json:"age_days"`
	RelatedCount int             `json:"related_count"`
}

func runCurate(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	// Get all active decisions
	opts := query.Options{Status: "active"}
	if curateType != "" {
		opts.Type = curateType
	}

	decisions, err := query.All(db, opts)
	if err != nil {
		return err
	}

	// Filter by age
	var candidates []CurationCandidate
	now := time.Now()
	cutoff := now.AddDate(0, 0, -curateOlderThan)

	for _, d := range decisions {
		createdAt, err := time.Parse(time.RFC3339Nano, d.CreatedAt)
		if err != nil {
			createdAt, _ = time.Parse(time.RFC3339, d.CreatedAt)
		}

		if curateOlderThan > 0 && createdAt.After(cutoff) {
			continue
		}

		// Filter by file pattern if specified
		if curateFilePattern != "" {
			matched := false
			for _, f := range d.Files {
				if matchPattern(curateFilePattern, f) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		age := int(now.Sub(createdAt).Hours() / 24)
		candidates = append(candidates, CurationCandidate{
			Decision:     d,
			Age:          age,
			RelatedCount: len(d.Files) + len(d.Refs),
		})
	}

	if len(candidates) == 0 {
		fmt.Println("No decisions found matching criteria.")
		return nil
	}

	if curateJSON {
		data, _ := json.MarshalIndent(candidates, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Found %d decisions for potential summarization:\n\n", len(candidates))
		for _, c := range candidates {
			fmt.Printf("\033[1m%s\033[0m [%s] (%d days old)\n", c.Decision.ID, c.Decision.Type, c.Age)
			fmt.Printf("  Problem: %s\n", c.Decision.Problem)
			fmt.Printf("  Choice: %s\n\n", c.Decision.Choice)
		}
	}

	return nil
}

func matchPattern(pattern, path string) bool {
	// Simple glob matching - just check if path starts with pattern prefix
	// Replace * with empty for prefix matching
	prefix := pattern
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			prefix = pattern[:i]
			break
		}
	}
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}
