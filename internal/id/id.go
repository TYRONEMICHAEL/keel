package id

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const IDPrefix = "DEC"

var validIDPattern = regexp.MustCompile(`^DEC-[a-f0-9]{4}$`)
var hexSuffixPattern = regexp.MustCompile(`^[a-fA-F0-9]{4}$`)

// Generate creates a hash-based decision ID.
// Uses content hashing to prevent collisions in multi-agent workflows.
// Format: DEC-xxxx (4 hex characters from content hash + entropy)
func Generate(problem, choice string) string {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 36)
	random := randomString(4)
	content := fmt.Sprintf("%s:%s:%s:%s", problem, choice, timestamp, random)

	// djb2 hash
	hash := uint32(5381)
	for i := 0; i < len(content); i++ {
		hash = ((hash << 5) + hash) ^ uint32(content[i])
	}

	// Convert to 4-character hex suffix
	suffix := fmt.Sprintf("%04x", hash&0xFFFF)
	return fmt.Sprintf("%s-%s", IDPrefix, suffix)
}

// IsValid checks if a string is a valid decision ID format.
func IsValid(id string) bool {
	return validIDPattern.MatchString(strings.ToLower(id))
}

// Normalize normalizes a decision ID input.
// Accepts: "DEC-a1b2", "dec-a1b2", "a1b2"
// Always returns lowercase suffix for consistency.
func Normalize(input string) (string, error) {
	trimmed := strings.TrimSpace(input)

	if strings.HasPrefix(strings.ToUpper(trimmed), "DEC-") {
		suffix := strings.ToLower(trimmed[4:])
		if !hexSuffixPattern.MatchString(suffix) {
			return "", fmt.Errorf("invalid decision ID: %s. Expected format: DEC-xxxx (4 hex chars)", input)
		}
		return fmt.Sprintf("DEC-%s", suffix), nil
	}

	// Assume it's just the suffix
	if hexSuffixPattern.MatchString(trimmed) {
		return fmt.Sprintf("DEC-%s", strings.ToLower(trimmed)), nil
	}

	return "", fmt.Errorf("invalid decision ID: %s. Expected format: DEC-xxxx (4 hex chars)", input)
}

func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
