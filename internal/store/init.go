package store

import (
	"fmt"
	"os"
)

// IsInitialized checks if Keel has been initialized in the repo
func IsInitialized(repoRoot string) bool {
	decisionsPath := GetDecisionsPath(repoRoot)
	_, err := os.Stat(decisionsPath)
	return err == nil
}

// RequireInit returns an error if Keel is not initialized
func RequireInit(repoRoot string) error {
	if !IsInitialized(repoRoot) {
		return fmt.Errorf("Keel not initialized. Run 'keel init' first (humans do this, not agents)")
	}
	return nil
}
