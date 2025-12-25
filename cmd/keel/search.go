package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/query"
	"github.com/tyroneavnit/keel/internal/types"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Full-text search across decisions",
	Long:  `Search decisions by problem, choice, or rationale text.`,
	RunE:  runSearch,
}

var (
	searchJSON   bool
	searchType   string
	searchStatus string
	searchLimit  int
)

func init() {
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output as JSON")
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "Filter by type: product, process, constraint, learning")
	searchCmd.Flags().StringVar(&searchStatus, "status", "", "Filter by status: active, superseded")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 0, "Maximum number of results")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	opts := query.Options{
		Type:   searchType,
		Status: searchStatus,
		Limit:  searchLimit,
	}

	var decisions []*types.Decision

	if len(args) > 0 {
		// Full-text search
		decisions, err = query.SearchFullText(db, args[0], opts)
	} else {
		// List all with filters
		decisions, err = query.All(db, opts)
	}

	if err != nil {
		return err
	}

	if searchJSON {
		data, _ := json.MarshalIndent(decisions, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(decisions) == 0 {
			fmt.Println("\033[2mNo decisions found.\033[0m")
		} else {
			for i, d := range decisions {
				printDecisionSummary(d)
				if i < len(decisions)-1 {
					fmt.Println()
				}
			}
		}
	}

	return nil
}
