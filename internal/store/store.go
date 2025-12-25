package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tyroneavnit/keel/internal/types"
)

const (
	KeelDir       = ".keel"
	DecisionsFile = "decisions.jsonl"
)

// GetKeelDir returns the path to the .keel directory
func GetKeelDir(repoRoot string) string {
	if repoRoot == "" {
		repoRoot, _ = os.Getwd()
	}
	return filepath.Join(repoRoot, KeelDir)
}

// GetDecisionsPath returns the path to the decisions.jsonl file
func GetDecisionsPath(repoRoot string) string {
	return filepath.Join(GetKeelDir(repoRoot), DecisionsFile)
}

// EnsureKeelDir creates the .keel directory if it doesn't exist
func EnsureKeelDir(repoRoot string) error {
	keelDir := GetKeelDir(repoRoot)
	return os.MkdirAll(keelDir, 0755)
}

// AppendDecision appends a decision to the JSONL file
func AppendDecision(decision *types.Decision, repoRoot string) error {
	if err := EnsureKeelDir(repoRoot); err != nil {
		return fmt.Errorf("failed to create keel directory: %w", err)
	}

	path := GetDecisionsPath(repoRoot)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open decisions file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(decision)
	if err != nil {
		return fmt.Errorf("failed to marshal decision: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write decision: %w", err)
	}

	return nil
}

// ReadAllDecisions reads all decisions from the JSONL file
func ReadAllDecisions(repoRoot string) ([]*types.Decision, error) {
	path := GetDecisionsPath(repoRoot)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []*types.Decision{}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open decisions file: %w", err)
	}
	defer f.Close()

	var decisions []*types.Decision
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var d types.Decision
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to parse line %d: %s\n", lineNum, line)
			continue
		}
		decisions = append(decisions, &d)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading decisions file: %w", err)
	}

	return decisions, nil
}

// GetLatestState returns the latest state of all decisions
// (later lines override earlier ones for the same ID)
func GetLatestState(repoRoot string) (map[string]*types.Decision, error) {
	decisions, err := ReadAllDecisions(repoRoot)
	if err != nil {
		return nil, err
	}

	state := make(map[string]*types.Decision)
	for _, d := range decisions {
		if existing, ok := state[d.ID]; ok {
			// Merge: update existing with new values
			merged := mergeDecisions(existing, d)
			state[d.ID] = merged
		} else {
			state[d.ID] = d
		}
	}

	return state, nil
}

// GetDecisionByID returns a decision by its ID
func GetDecisionByID(id string, repoRoot string) (*types.Decision, error) {
	state, err := GetLatestState(repoRoot)
	if err != nil {
		return nil, err
	}
	return state[id], nil
}

// GetActiveDecisions returns all active decisions
func GetActiveDecisions(repoRoot string) ([]*types.Decision, error) {
	state, err := GetLatestState(repoRoot)
	if err != nil {
		return nil, err
	}

	var active []*types.Decision
	for _, d := range state {
		if d.Status == types.StatusActive {
			active = append(active, d)
		}
	}
	return active, nil
}

// mergeDecisions merges two decisions, with newer values overriding older ones
func mergeDecisions(existing, newer *types.Decision) *types.Decision {
	merged := *existing

	// Override with newer values if set
	if newer.Status != "" {
		merged.Status = newer.Status
	}
	if newer.SupersededBy != nil {
		merged.SupersededBy = newer.SupersededBy
	}
	if newer.Supersedes != nil {
		merged.Supersedes = newer.Supersedes
	}

	return &merged
}
