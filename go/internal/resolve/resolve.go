// Package resolve implements .terraform-version file discovery and version
// constraint resolution. It walks the directory tree upward to find version
// files, parses their content, and applies the tfenv version precedence chain.
package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/logging"
)

// VersionResult holds the resolved version string and the source that
// provided it (for diagnostics and logging).
type VersionResult struct {
	Version string // The version string or keyword (e.g. "1.5.0", "latest", "latest:^1.5")
	Source  string // Where it came from: "TFENV_TERRAFORM_VERSION", "/path/to/.terraform-version", etc.
}

// versionFileName is the name of the per-directory version file.
const versionFileName = ".terraform-version"

// ResolveVersionFile determines the Terraform version to use by searching
// the following sources in order of decreasing precedence:
//
//  1. TFENV_TERRAFORM_VERSION environment variable (Config.TerraformVersion)
//  2. .terraform-version in Config.Dir (or working directory)
//  3. .terraform-version walking up parent directories toward filesystem root
//  4. .terraform-version walking up from $HOME toward filesystem root
//  5. ${Config.ConfigDir}/version (default version file)
//
// Returns an error if no version is found anywhere.
func ResolveVersionFile(cfg *config.Config) (*VersionResult, error) {
	// 1. Environment variable takes highest precedence.
	if cfg.TerraformVersion != "" {
		version := cleanVersion(cfg.TerraformVersion)
		logging.Debug("version set by TFENV_TERRAFORM_VERSION", "version", version)
		return &VersionResult{
			Version: version,
			Source:  "TFENV_TERRAFORM_VERSION",
		}, nil
	}

	logging.Debug("TFENV_TERRAFORM_VERSION not set, searching for version file")

	// 2–3. Walk up from TFENV_DIR (or working directory).
	startDir := cfg.Dir
	if startDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determining working directory: %w", err)
		}
		startDir = wd
	}

	if result, err := findVersionFileWalk(startDir); err != nil {
		return nil, err
	} else if result != nil {
		return result, nil
	}

	// 4. Walk up from $HOME.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}

	if result, err := findVersionFileWalk(homeDir); err != nil {
		return nil, err
	} else if result != nil {
		return result, nil
	}

	// 5. Fall back to ConfigDir/version.
	defaultFile := filepath.Join(cfg.ConfigDir, "version")
	logging.Debug("no version file found in search paths, checking default", "path", defaultFile)

	result, err := readVersionFile(defaultFile)
	if err != nil {
		return nil, fmt.Errorf("no version file found: searched from %s and %s, "+
			"and no default version in %s", startDir, homeDir, defaultFile)
	}

	return result, nil
}

// findVersionFileWalk walks from dir upward through parent directories
// looking for a .terraform-version file. Returns nil, nil if no file is
// found (not an error — the caller should try the next search path).
func findVersionFileWalk(dir string) (*VersionResult, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving absolute path for %q: %w", dir, err)
	}

	logging.Debug("walking directory tree for version file", "start", dir)

	for {
		candidate := filepath.Join(dir, versionFileName)
		logging.Debug("checking for version file", "path", candidate)

		if _, err := os.Stat(candidate); err == nil {
			result, err := readVersionFile(candidate)
			if err != nil {
				logging.Debug("version file found but unreadable or empty", "path", candidate, "err", err)
			} else {
				logging.Debug("version file found", "path", candidate, "version", result.Version)
				return result, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			break
		}
		dir = parent
	}

	return nil, nil
}

// readVersionFile reads and parses a version file, returning the first
// non-empty line after stripping comments, whitespace, carriage returns,
// and a leading "v" prefix.
func readVersionFile(path string) (*VersionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading version file %q: %w", path, err)
	}

	version := parseVersionContent(string(data))
	if version == "" {
		return nil, fmt.Errorf("version file %q is empty or contains only comments", path)
	}

	return &VersionResult{
		Version: version,
		Source:  path,
	}, nil
}

// parseVersionContent extracts the version string from file content.
// For each line it:
//  1. Strips everything after # (comment removal)
//  2. Strips carriage returns
//  3. Trims whitespace
//  4. Skips blank lines
//  5. Returns the first non-empty line with any leading "v" prefix removed
func parseVersionContent(content string) string {
	for _, line := range strings.Split(content, "\n") {
		// Strip comments.
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}

		// Strip carriage returns.
		line = strings.ReplaceAll(line, "\r", "")

		// Trim whitespace.
		line = strings.TrimSpace(line)

		// Skip blank lines.
		if line == "" {
			continue
		}

		// Strip leading "v" prefix.
		line = strings.TrimPrefix(line, "v")

		return line
	}

	return ""
}

// cleanVersion trims whitespace, strips carriage returns, and removes a
// leading "v" prefix from an environment variable value.
func cleanVersion(v string) string {
	v = strings.ReplaceAll(v, "\r", "")
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}
