package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade keel to the latest version",
	Long:  `Download and install the latest version of keel from GitHub releases.`,
	RunE:  runUpgrade,
}

var upgradeCheck bool

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeCheck, "check", false, "Only check for updates, don't install")
	rootCmd.AddCommand(upgradeCmd)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	// Get latest version from GitHub
	resp, err := http.Get("https://api.github.com/repos/TYRONEMICHAEL/keel/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")

	if latest == current {
		fmt.Printf("Already at latest version (%s)\n", version)
		return nil
	}

	fmt.Printf("Current: %s\n", version)
	fmt.Printf("Latest:  %s\n", release.TagName)

	if upgradeCheck {
		fmt.Println("\nRun 'keel upgrade' to install the update.")
		return nil
	}

	// Download and install
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	url := fmt.Sprintf("https://github.com/TYRONEMICHAEL/keel/releases/download/%s/keel-%s-%s",
		release.TagName, goos, goarch)

	fmt.Printf("\nDownloading %s...\n", url)

	resp, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "keel-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to chmod: %w", err)
	}

	// Replace current binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		// Try with sudo on permission error
		if os.IsPermission(err) {
			fmt.Println("Permission denied. Trying with sudo...")
			cmd := exec.Command("sudo", "mv", tmpPath, execPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to replace binary: %w", err)
			}
		} else {
			return fmt.Errorf("failed to replace binary: %w", err)
		}
	}

	fmt.Printf("\nUpgraded to %s\n", release.TagName)
	return nil
}
