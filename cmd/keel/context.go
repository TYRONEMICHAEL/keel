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

var contextCmd = &cobra.Command{
	Use:   "context [path]",
	Short: "Get decisions affecting a file, symbol, or reference",
	Long:  `Display all decisions that affect a given file path, symbol, or external reference.`,
	RunE:  runContext,
}

var (
	contextJSON bool
	contextRef  string
)

func init() {
	contextCmd.Flags().BoolVar(&contextJSON, "json", false, "Output as JSON")
	contextCmd.Flags().StringVar(&contextRef, "ref", "", "Get decisions linked to an external reference (issue, epic, etc.)")
	rootCmd.AddCommand(contextCmd)
}

func runContext(cmd *cobra.Command, args []string) error {
	repoRoot, _ := os.Getwd()
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	var decisions []*types.Decision
	var constraints []*types.Decision
	var path string

	if contextRef != "" {
		// Query by ref
		path = fmt.Sprintf("ref:%s", contextRef)
		decisions, err = query.ByRef(db, contextRef)
		if err != nil {
			return err
		}
		constraints, err = query.ActiveConstraints(db)
		if err != nil {
			return err
		}
	} else if len(args) > 0 {
		// Query by file path
		path = args[0]
		result, err := query.ForContext(db, path)
		if err != nil {
			return err
		}
		decisions = result.Decisions
		constraints = result.Constraints

		// If no file decisions, try symbol lookup
		if len(decisions) == 0 {
			decisions, err = query.BySymbol(db, path)
			if err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("must provide a path or --ref option")
	}

	if contextJSON {
		output := map[string]interface{}{
			"path":        path,
			"decisions":   decisions,
			"constraints": constraints,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		printContextResult(decisions, constraints)
	}

	return nil
}

func printContextResult(decisions, constraints []*types.Decision) {
	if len(decisions) > 0 {
		fmt.Println("\033[1mDecisions affecting this file:\033[0m\n")
		for _, d := range decisions {
			printDecisionSummary(d)
			fmt.Println()
		}
	} else {
		fmt.Println("\033[2mNo decisions directly affect this file.\033[0m")
	}

	if len(constraints) > 0 {
		fmt.Println("\n\033[1mActive constraints:\033[0m\n")
		for _, c := range constraints {
			fmt.Printf("  \033[1m%s\033[0m %s\n", c.ID, c.Choice)
		}
	}
}

func printDecisionSummary(d *types.Decision) {
	fmt.Printf("\033[1m%s\033[0m [%s] %s\n", d.ID, colorType(string(d.Type)), colorStatus(string(d.Status)))
	fmt.Printf("  \033[2mProblem:\033[0m %s\n", d.Problem)
	fmt.Printf("  \033[2mChoice:\033[0m %s\n", d.Choice)
}
