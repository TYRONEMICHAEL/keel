package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// DecisionType represents the category of a decision
type DecisionType string

const (
	TypeProduct    DecisionType = "product"
	TypeProcess    DecisionType = "process"
	TypeConstraint DecisionType = "constraint"
	TypeLearning   DecisionType = "learning"
)

// DecisionStatus represents the current state of a decision
type DecisionStatus string

const (
	StatusActive     DecisionStatus = "active"
	StatusSuperseded DecisionStatus = "superseded"
)

// DecidedBy represents who made the decision
type DecidedBy struct {
	Role       string  `json:"role"`                 // "human" or "agent"
	Identifier *string `json:"identifier,omitempty"` // email, agent name, etc.
}

// Decision represents a recorded decision in the ledger
type Decision struct {
	ID              string         `json:"id"`
	CreatedAt       string         `json:"created_at"`
	Type            DecisionType   `json:"type"`
	Problem         string         `json:"problem"`
	Choice          string         `json:"choice"`
	Rationale       *string        `json:"rationale,omitempty"`
	Tradeoffs       []string       `json:"tradeoffs,omitempty"`
	DecidedBy       DecidedBy      `json:"decided_by"`
	Files           []string       `json:"files,omitempty"`
	Symbols         []string       `json:"symbols,omitempty"`
	Refs            []string       `json:"refs,omitempty"`
	Status          DecisionStatus `json:"status"`
	SupersededBy    *string        `json:"superseded_by,omitempty"`
	Supersedes      *string        `json:"supersedes,omitempty"`
	Hypothesis      *string        `json:"hypothesis,omitempty"`
	SuccessCriteria *string        `json:"success_criteria,omitempty"`
}

// DecisionInput represents the input for creating a new decision
type DecisionInput struct {
	Type            DecisionType `json:"type"`
	Problem         string       `json:"problem"`
	Choice          string       `json:"choice"`
	Rationale       *string      `json:"rationale,omitempty"`
	Tradeoffs       []string     `json:"tradeoffs,omitempty"`
	DecidedBy       *DecidedBy   `json:"decided_by,omitempty"`
	Files           []string     `json:"files,omitempty"`
	Symbols         []string     `json:"symbols,omitempty"`
	Refs            []string     `json:"refs,omitempty"`
	Hypothesis      *string      `json:"hypothesis,omitempty"`
	SuccessCriteria *string      `json:"success_criteria,omitempty"`
	Supersedes      *string      `json:"supersedes,omitempty"`
}

// ValidDecisionTypes returns all valid decision types
func ValidDecisionTypes() []DecisionType {
	return []DecisionType{TypeProduct, TypeProcess, TypeConstraint, TypeLearning}
}

// IsValidType checks if a string is a valid decision type
func IsValidType(t string) bool {
	switch DecisionType(t) {
	case TypeProduct, TypeProcess, TypeConstraint, TypeLearning:
		return true
	}
	return false
}

// ParseDecision parses a JSON line into a Decision
func ParseDecision(data []byte) (*Decision, error) {
	var d Decision
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to parse decision: %w", err)
	}
	return &d, nil
}

// ToJSON converts a Decision to JSON bytes
func (d *Decision) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// NewDecision creates a new Decision from input
func NewDecision(id string, input DecisionInput) *Decision {
	decidedBy := DecidedBy{Role: "human"}
	if input.DecidedBy != nil {
		decidedBy = *input.DecidedBy
	}

	return &Decision{
		ID:              id,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Type:            input.Type,
		Problem:         input.Problem,
		Choice:          input.Choice,
		Rationale:       input.Rationale,
		Tradeoffs:       input.Tradeoffs,
		DecidedBy:       decidedBy,
		Files:           input.Files,
		Symbols:         input.Symbols,
		Refs:            input.Refs,
		Status:          StatusActive,
		Supersedes:      input.Supersedes,
		Hypothesis:      input.Hypothesis,
		SuccessCriteria: input.SuccessCriteria,
	}
}
