package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/query"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check that file references still exist",
	Long:  `Validate that all files referenced by decisions still exist in the repository.`,
	RunE:  runValidate,
}

var validateJSON bool

func init() {
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(validateCmd)
}

type ValidationIssue struct {
	DecisionID string `json:"decision_id"`
	FilePath   string `json:"file_path"`
	Issue      string `json:"issue"`
}

func runValidate(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	// Get all active decisions
	decisions, err := query.All(db, query.Options{Status: "active"})
	if err != nil {
		return err
	}

	var issues []ValidationIssue

	for _, d := range decisions {
		for _, file := range d.Files {
			fullPath := filepath.Join(repoRoot, file)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				issues = append(issues, ValidationIssue{
					DecisionID: d.ID,
					FilePath:   file,
					Issue:      "file not found",
				})
			}
		}
	}

	if validateJSON {
		data, _ := json.MarshalIndent(issues, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(issues) == 0 {
			fmt.Println("\033[32m✓ All file references are valid\033[0m")
		} else {
			fmt.Printf("\033[31m✗ Found %d validation issues:\033[0m\n\n", len(issues))
			for _, issue := range issues {
				fmt.Printf("  \033[1m%s\033[0m: %s - %s\n", issue.DecisionID, issue.FilePath, issue.Issue)
			}
		}
	}

	if len(issues) > 0 {
		os.Exit(1)
	}

	return nil
}
