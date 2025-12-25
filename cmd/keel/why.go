package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tyroneavnit/keel/internal/id"
	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/query"
	"github.com/tyroneavnit/keel/internal/types"
)

var whyCmd = &cobra.Command{
	Use:   "why <id>",
	Short: "Show full decision details",
	Long:  `Display the complete details of a decision by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runWhy,
}

var whyJSON bool

func init() {
	whyCmd.Flags().BoolVar(&whyJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(whyCmd)
}

func runWhy(cmd *cobra.Command, args []string) error {
	rawID := args[0]
	normalizedID, err := id.Normalize(rawID)
	if err != nil {
		return err
	}

	repoRoot, _ := os.Getwd()
	db, err := index.Open(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open index: %w", err)
	}
	defer db.Close()

	decision, err := query.ByID(db, normalizedID)
	if err != nil {
		return err
	}

	if decision == nil {
		return fmt.Errorf("decision %s not found", normalizedID)
	}

	if whyJSON {
		output, _ := json.MarshalIndent(decision, "", "  ")
		fmt.Println(string(output))
	} else {
		printDecisionFull(decision)
	}

	return nil
}

func printDecisionFull(d *types.Decision) {
	fmt.Printf("\033[1mDecision %s\033[0m\n\n", d.ID)
	fmt.Printf("\033[2mType:\033[0m     %s\n", colorType(string(d.Type)))
	fmt.Printf("\033[2mStatus:\033[0m   %s\n", colorStatus(string(d.Status)))
	fmt.Printf("\033[2mCreated:\033[0m  %s\n\n", d.CreatedAt)

	fmt.Printf("\033[1mProblem\033[0m\n%s\n\n", d.Problem)
	fmt.Printf("\033[1mChoice\033[0m\n%s\n", d.Choice)

	if d.Rationale != nil && *d.Rationale != "" {
		fmt.Printf("\n\033[1mRationale\033[0m\n%s\n", *d.Rationale)
	}

	if len(d.Tradeoffs) > 0 {
		fmt.Printf("\n\033[1mTradeoffs\033[0m\n")
		for _, t := range d.Tradeoffs {
			fmt.Printf("  - %s\n", t)
		}
	}

	if len(d.Files) > 0 {
		fmt.Printf("\n\033[1mFiles\033[0m\n")
		for _, f := range d.Files {
			fmt.Printf("  %s\n", f)
		}
	}

	if len(d.Symbols) > 0 {
		fmt.Printf("\n\033[1mSymbols\033[0m\n")
		for _, s := range d.Symbols {
			fmt.Printf("  %s\n", s)
		}
	}

	if len(d.Refs) > 0 {
		fmt.Printf("\n\033[1mRefs\033[0m\n")
		for _, r := range d.Refs {
			fmt.Printf("  %s\n", r)
		}
	}

	fmt.Println()
	identifier := ""
	if d.DecidedBy.Identifier != nil {
		identifier = fmt.Sprintf(" (%s)", *d.DecidedBy.Identifier)
	}
	fmt.Printf("\033[2mDecided by:\033[0m %s%s\n", d.DecidedBy.Role, identifier)

	if d.Supersedes != nil {
		fmt.Printf("\033[2mSupersedes:\033[0m %s\n", *d.Supersedes)
	}
	if d.SupersededBy != nil {
		fmt.Printf("\033[2mSuperseded by:\033[0m %s\n", *d.SupersededBy)
	}
}

func colorType(t string) string {
	colors := map[string]string{
		"product":    "\033[34m",
		"process":    "\033[35m",
		"constraint": "\033[31m",
		"learning":   "\033[36m",
	}
	color := colors[t]
	if color == "" {
		return t
	}
	return color + t + "\033[0m"
}

func colorStatus(s string) string {
	colors := map[string]string{
		"active":     "\033[32m",
		"superseded": "\033[2m",
	}
	color := colors[s]
	if color == "" {
		return s
	}
	return color + s + "\033[0m"
}
