package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tyroneavnit/keel/internal/store"
	"github.com/tyroneavnit/keel/internal/types"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

const IndexFile = "index.sqlite"

// DB wraps a SQLite database connection
type DB struct {
	*sql.DB
	repoRoot string
}

// GetIndexPath returns the path to the SQLite index file
func GetIndexPath(repoRoot string) string {
	return filepath.Join(store.GetKeelDir(repoRoot), IndexFile)
}

// Open opens or creates the SQLite index
func Open(repoRoot string) (*DB, error) {
	if repoRoot == "" {
		var err error
		repoRoot, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Ensure .keel directory exists
	if err := store.EnsureKeelDir(repoRoot); err != nil {
		return nil, err
	}

	indexPath := GetIndexPath(repoRoot)
	db, err := sql.Open("sqlite3", indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	idx := &DB{DB: db, repoRoot: repoRoot}

	if err := idx.createSchema(); err != nil {
		db.Close()
		return nil, err
	}

	if idx.needsRebuild() {
		if err := idx.rebuild(); err != nil {
			db.Close()
			return nil, err
		}
	}

	return idx, nil
}

func (db *DB) createSchema() error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS decisions (
			id TEXT PRIMARY KEY,
			created_at TEXT NOT NULL,
			type TEXT NOT NULL,
			problem TEXT NOT NULL,
			choice TEXT NOT NULL,
			rationale TEXT,
			decided_by_role TEXT NOT NULL,
			decided_by_identifier TEXT,
			status TEXT NOT NULL,
			supersedes TEXT,
			superseded_by TEXT,
			raw_json TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS decision_files (
			decision_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			PRIMARY KEY (decision_id, file_path)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_files_path ON decision_files(file_path)`,
		`CREATE TABLE IF NOT EXISTS decision_symbols (
			decision_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			PRIMARY KEY (decision_id, symbol)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_symbols_name ON decision_symbols(symbol)`,
		`CREATE TABLE IF NOT EXISTS decision_refs (
			decision_id TEXT NOT NULL,
			ref_id TEXT NOT NULL,
			PRIMARY KEY (decision_id, ref_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refs_id ON decision_refs(ref_id)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS decisions_fts USING fts5(
			id,
			problem,
			choice,
			rationale,
			content='decisions',
			content_rowid='rowid'
		)`,
		`CREATE TRIGGER IF NOT EXISTS decisions_ai AFTER INSERT ON decisions BEGIN
			INSERT INTO decisions_fts(rowid, id, problem, choice, rationale)
			VALUES (NEW.rowid, NEW.id, NEW.problem, NEW.choice, NEW.rationale);
		END`,
		`CREATE TRIGGER IF NOT EXISTS decisions_ad AFTER DELETE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, id, problem, choice, rationale)
			VALUES('delete', OLD.rowid, OLD.id, OLD.problem, OLD.choice, OLD.rationale);
		END`,
		`CREATE TRIGGER IF NOT EXISTS decisions_au AFTER UPDATE ON decisions BEGIN
			INSERT INTO decisions_fts(decisions_fts, rowid, id, problem, choice, rationale)
			VALUES('delete', OLD.rowid, OLD.id, OLD.problem, OLD.choice, OLD.rationale);
			INSERT INTO decisions_fts(rowid, id, problem, choice, rationale)
			VALUES (NEW.rowid, NEW.id, NEW.problem, NEW.choice, NEW.rationale);
		END`,
		`CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return nil
}

func (db *DB) insertDecision(d *types.Decision) error {
	rawJSON, err := json.Marshal(d)
	if err != nil {
		return err
	}

	var rationale, identifier, supersedes, supersededBy interface{}
	if d.Rationale != nil {
		rationale = *d.Rationale
	}
	if d.DecidedBy.Identifier != nil {
		identifier = *d.DecidedBy.Identifier
	}
	if d.Supersedes != nil {
		supersedes = *d.Supersedes
	}
	if d.SupersededBy != nil {
		supersededBy = *d.SupersededBy
	}

	_, err = db.Exec(`
		INSERT OR REPLACE INTO decisions (
			id, created_at, type, problem, choice, rationale,
			decided_by_role, decided_by_identifier, status,
			supersedes, superseded_by, raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.CreatedAt, d.Type, d.Problem, d.Choice, rationale,
		d.DecidedBy.Role, identifier, d.Status,
		supersedes, supersededBy, string(rawJSON),
	)
	if err != nil {
		return err
	}

	// Insert file associations
	for _, file := range d.Files {
		_, err = db.Exec(`INSERT OR IGNORE INTO decision_files (decision_id, file_path) VALUES (?, ?)`,
			d.ID, file)
		if err != nil {
			return err
		}
	}

	// Insert symbol associations
	for _, symbol := range d.Symbols {
		_, err = db.Exec(`INSERT OR IGNORE INTO decision_symbols (decision_id, symbol) VALUES (?, ?)`,
			d.ID, symbol)
		if err != nil {
			return err
		}
	}

	// Insert ref associations
	for _, ref := range d.Refs {
		_, err = db.Exec(`INSERT OR IGNORE INTO decision_refs (decision_id, ref_id) VALUES (?, ?)`,
			d.ID, ref)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) needsRebuild() bool {
	decisionsPath := store.GetDecisionsPath(db.repoRoot)

	info, err := os.Stat(decisionsPath)
	if os.IsNotExist(err) {
		return false // No source file
	}

	var storedMtime string
	err = db.QueryRow("SELECT value FROM metadata WHERE key = ?", "jsonl_mtime").Scan(&storedMtime)
	if err != nil {
		return true // No mtime recorded
	}

	currentMtime := fmt.Sprintf("%d", info.ModTime().UnixNano())
	return currentMtime != storedMtime
}

func (db *DB) rebuild() error {
	// Clear existing data
	tables := []string{"decision_files", "decision_symbols", "decision_refs", "decisions"}
	for _, table := range tables {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			return err
		}
	}

	// Read all decisions from JSONL
	decisions, err := store.ReadAllDecisions(db.repoRoot)
	if err != nil {
		return err
	}

	// Build latest state
	state := make(map[string]*types.Decision)
	for _, d := range decisions {
		if existing, ok := state[d.ID]; ok {
			// Merge
			if d.Status != "" {
				existing.Status = d.Status
			}
			if d.SupersededBy != nil {
				existing.SupersededBy = d.SupersededBy
			}
		} else {
			state[d.ID] = d
		}
	}

	// Insert all decisions
	for _, d := range state {
		if err := db.insertDecision(d); err != nil {
			return err
		}
	}

	// Store mtime
	decisionsPath := store.GetDecisionsPath(db.repoRoot)
	if info, err := os.Stat(decisionsPath); err == nil {
		mtime := fmt.Sprintf("%d", info.ModTime().UnixNano())
		_, err = db.Exec("INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?)",
			"jsonl_mtime", mtime)
		if err != nil {
			return err
		}
	}

	return nil
}

// IndexDecision adds a decision to the index
func (db *DB) IndexDecision(d *types.Decision) error {
	if err := db.insertDecision(d); err != nil {
		return err
	}

	// Update mtime
	decisionsPath := store.GetDecisionsPath(db.repoRoot)
	if info, err := os.Stat(decisionsPath); err == nil {
		mtime := fmt.Sprintf("%d", info.ModTime().UnixNano())
		_, err = db.Exec("INSERT OR REPLACE INTO metadata (key, value) VALUES (?, ?)",
			"jsonl_mtime", mtime)
		if err != nil {
			return err
		}
	}

	return nil
}
