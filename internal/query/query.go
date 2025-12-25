package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tyroneavnit/keel/internal/index"
	"github.com/tyroneavnit/keel/internal/types"
)

// Options for querying decisions
type Options struct {
	Type   string
	Status string
	Limit  int
}

// ContextResult contains decisions and constraints for a given context
type ContextResult struct {
	Decisions   []*types.Decision
	Constraints []*types.Decision
}

func rowToDecision(rawJSON string) (*types.Decision, error) {
	var d types.Decision
	if err := json.Unmarshal([]byte(rawJSON), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// ByID queries a decision by its ID
func ByID(db *index.DB, id string) (*types.Decision, error) {
	var rawJSON string
	err := db.QueryRow("SELECT raw_json FROM decisions WHERE id = ?", id).Scan(&rawJSON)
	if err != nil {
		return nil, nil // Not found
	}
	return rowToDecision(rawJSON)
}

// ByFile queries decisions affecting a file path
func ByFile(db *index.DB, filePath string) ([]*types.Decision, error) {
	// Support glob patterns with LIKE
	pattern := filePath
	if strings.Contains(filePath, "*") {
		pattern = strings.ReplaceAll(filePath, "*", "%")
	}

	rows, err := db.Query(`
		SELECT d.raw_json FROM decisions d
		INNER JOIN decision_files df ON d.id = df.decision_id
		WHERE df.file_path LIKE ?
		AND d.status = 'active'
		ORDER BY d.created_at DESC
	`, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// BySymbol queries decisions for a symbol
func BySymbol(db *index.DB, symbol string) ([]*types.Decision, error) {
	rows, err := db.Query(`
		SELECT d.raw_json FROM decisions d
		INNER JOIN decision_symbols ds ON d.id = ds.decision_id
		WHERE ds.symbol = ?
		AND d.status = 'active'
		ORDER BY d.created_at DESC
	`, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// ByRef queries decisions linked to a reference ID
func ByRef(db *index.DB, refID string) ([]*types.Decision, error) {
	rows, err := db.Query(`
		SELECT d.raw_json FROM decisions d
		INNER JOIN decision_refs dr ON d.id = dr.decision_id
		WHERE dr.ref_id = ?
		AND d.status = 'active'
		ORDER BY d.created_at DESC
	`, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// All queries all decisions with optional filters
func All(db *index.DB, opts Options) ([]*types.Decision, error) {
	sql := "SELECT raw_json FROM decisions WHERE 1=1"
	var args []interface{}

	if opts.Type != "" {
		sql += " AND type = ?"
		args = append(args, opts.Type)
	}

	if opts.Status != "" {
		sql += " AND status = ?"
		args = append(args, opts.Status)
	}

	sql += " ORDER BY created_at DESC"

	if opts.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// SearchFullText performs a full-text search using FTS5
func SearchFullText(db *index.DB, query string, opts Options) ([]*types.Decision, error) {
	sql := `
		SELECT d.raw_json FROM decisions d
		INNER JOIN decisions_fts fts ON d.id = fts.id
		WHERE decisions_fts MATCH ?
	`
	args := []interface{}{query}

	if opts.Type != "" {
		sql += " AND d.type = ?"
		args = append(args, opts.Type)
	}

	if opts.Status != "" {
		sql += " AND d.status = ?"
		args = append(args, opts.Status)
	}

	sql += " ORDER BY rank"

	if opts.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		// Fall back to LIKE search on FTS error
		return searchLike(db, query, opts)
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

func searchLike(db *index.DB, query string, opts Options) ([]*types.Decision, error) {
	pattern := "%" + query + "%"

	sql := `
		SELECT raw_json FROM decisions
		WHERE (problem LIKE ? OR choice LIKE ? OR rationale LIKE ?)
	`
	args := []interface{}{pattern, pattern, pattern}

	if opts.Type != "" {
		sql += " AND type = ?"
		args = append(args, opts.Type)
	}

	if opts.Status != "" {
		sql += " AND status = ?"
		args = append(args, opts.Status)
	}

	sql += " ORDER BY created_at DESC"

	if opts.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// ActiveConstraints returns all active constraint decisions
func ActiveConstraints(db *index.DB) ([]*types.Decision, error) {
	rows, err := db.Query(`
		SELECT raw_json FROM decisions
		WHERE type = 'constraint' AND status = 'active'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*types.Decision
	for rows.Next() {
		var rawJSON string
		if err := rows.Scan(&rawJSON); err != nil {
			continue
		}
		if d, err := rowToDecision(rawJSON); err == nil {
			decisions = append(decisions, d)
		}
	}

	return decisions, nil
}

// ForContext returns decisions and constraints for a given file path
func ForContext(db *index.DB, path string) (*ContextResult, error) {
	decisions, err := ByFile(db, path)
	if err != nil {
		return nil, err
	}

	constraints, err := ActiveConstraints(db)
	if err != nil {
		return nil, err
	}

	return &ContextResult{
		Decisions:   decisions,
		Constraints: constraints,
	}, nil
}
