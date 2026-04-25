// Package uninstall implements the tfenv uninstall command — removing
// installed Terraform versions from the local filesystem.
package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/list"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

func init() {
	cli.Register("uninstall", "Uninstall a specific version of Terraform", Run)
}

// Run implements the tfenv uninstall command. It accepts zero or more
// arguments:
//   - No args: resolve version from env var or version file chain
//   - One or more args: each is a version specifier to uninstall
//
// Returns 0 on full success, 1 if any uninstall failed.
func Run(args []string) int {
	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	versions, err := collectVersions(args, cfg)
	if err != nil {
		logging.Error("failed to determine version to uninstall", "err", err)
		return 1
	}

	failures := 0
	for _, specifier := range versions {
		if err := uninstallSingle(specifier, cfg); err != nil {
			logging.Error("failed to uninstall version",
				"specifier", specifier, "err", err)
			failures++
		}
	}

	// Post-cleanup: remove versions/ dir if empty (best-effort).
	versionsDir := filepath.Join(cfg.ConfigDir, "versions")
	_ = os.Remove(versionsDir)

	if failures > 0 {
		return 1
	}
	return 0
}

// collectVersions determines the list of version specifiers to uninstall.
// Priority: command-line args > TFENV_TERRAFORM_VERSION > version file chain.
func collectVersions(args []string, cfg *config.Config) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	// No args — try env var.
	if cfg.TerraformVersion != "" {
		version := strings.TrimSpace(strings.TrimPrefix(cfg.TerraformVersion, "v"))
		return []string{version}, nil
	}

	// No env var — try version file chain.
	result, err := resolve.ResolveVersionFile(cfg)
	if err != nil {
		return nil, fmt.Errorf(
			"no version specified on command line, "+
				"no TFENV_TERRAFORM_VERSION set, and no version file found: %w", err)
	}

	return []string{result.Version}, nil
}

// uninstallSingle handles the uninstall of one version specifier.
// It resolves keywords against locally installed versions, validates
// the version string, and removes the version directory.
func uninstallSingle(specifier string, cfg *config.Config) error {
	specifier = strings.TrimSpace(strings.TrimPrefix(specifier, "v"))

	// Reject unsupported keywords.
	if specifier == "min-required" {
		return fmt.Errorf("%q is not a valid uninstall target — "+
			"resolve it first with tfenv resolve-version", specifier)
	}
	if specifier == "latest-allowed" {
		return fmt.Errorf("%q is not a valid uninstall target — "+
			"resolve it first with tfenv resolve-version", specifier)
	}

	// Resolve keywords against locally installed versions.
	version, err := resolveLocalVersion(specifier, cfg)
	if err != nil {
		return err
	}

	// Path traversal protection.
	if err := validateVersion(version); err != nil {
		return err
	}

	// Verify the version is installed.
	versionDir := filepath.Join(cfg.ConfigDir, "versions", version)
	binaryPath := filepath.Join(versionDir, "terraform")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("Terraform v%s is not installed", version)
	} else if err != nil {
		return fmt.Errorf("checking installation of v%s: %w", version, err)
	}

	// Remove the version directory.
	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("removing version directory %s: %w", versionDir, err)
	}

	fmt.Fprintf(os.Stdout, "Terraform v%s is successfully uninstalled\n", version)
	return nil
}

// resolveLocalVersion resolves a version specifier against locally
// installed versions. For "latest" and "latest:<regex>" it searches
// the local installation directory. For exact versions, it returns
// the specifier as-is.
func resolveLocalVersion(specifier string, cfg *config.Config) (string, error) {
	switch {
	case specifier == "latest":
		return resolveLatestLocal("", cfg)

	case strings.HasPrefix(specifier, "latest:"):
		pattern := strings.TrimPrefix(specifier, "latest:")
		return resolveLatestLocal(pattern, cfg)

	default:
		return specifier, nil
	}
}

// resolveLatestLocal finds the newest locally installed version,
// optionally filtered by a regex pattern. An empty pattern matches all
// stable versions.
func resolveLatestLocal(pattern string, cfg *config.Config) (string, error) {
	localVersions, err := list.ListLocal(cfg)
	if err != nil {
		return "", fmt.Errorf("listing local versions: %w", err)
	}

	if len(localVersions) == 0 {
		return "", fmt.Errorf("no versions installed")
	}

	if pattern == "" {
		// "latest" — use resolve engine for consistency.
		resolved, err := resolve.ResolveVersion("latest", localVersions, "local", cfg)
		if err != nil {
			return "", fmt.Errorf("resolving latest local version: %w", err)
		}
		return resolved.Version, nil
	}

	// "latest:<regex>" — filter locally and pick newest.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex %q: %w", pattern, err)
	}

	var matched []string
	for _, v := range localVersions {
		if re.MatchString(v) {
			matched = append(matched, v)
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("no installed version matches regex %q", pattern)
	}

	resolved, err := resolve.ResolveVersion("latest", matched, "local", cfg)
	if err != nil {
		return "", fmt.Errorf("resolving latest from matched versions: %w", err)
	}
	return resolved.Version, nil
}

// validateVersion rejects version strings that could cause path traversal.
func validateVersion(version string) error {
	if strings.Contains(version, "/") ||
		strings.Contains(version, "\\") ||
		strings.Contains(version, "..") ||
		strings.ContainsRune(version, 0) {
		return fmt.Errorf("invalid version string: %q", version)
	}
	return nil
}
