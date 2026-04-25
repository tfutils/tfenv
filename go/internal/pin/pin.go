// Package pin implements the tfenv pin command — writing the currently
// resolved version to .terraform-version in the current directory.
package pin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/list"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

func init() {
	cli.Register("pin", "Write the current active version to .terraform-version", Run)
}

// Run implements the tfenv pin command.
// It accepts no arguments. Resolves the current version from the version
// file chain, resolves keywords to a concrete installed version, and writes
// it to .terraform-version in the current directory.
//
// Returns 0 on success, 1 on error.
func Run(args []string) int {
	if len(args) > 0 {
		logging.Error("usage: tfenv pin")
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	// Verify versions directory exists and is accessible.
	versionsDir := filepath.Join(cfg.ConfigDir, "versions")
	if _, err := os.Stat(versionsDir); err != nil {
		if os.IsNotExist(err) {
			logging.Error("No versions available. Please install one with: tfenv install")
			return 1
		}
		logging.Error("failed to access versions directory", "err", err)
		return 1
	}

	// Resolve current version from the version file chain.
	result, err := resolve.ResolveVersionFile(cfg)
	if err != nil {
		logging.Error("failed to resolve current version", "err", err)
		return 1
	}

	logging.Debug("version file resolved", "version", result.Version, "source", result.Source)

	// Resolve keywords (latest, latest:<regex>, etc.) to a concrete
	// installed version.
	version, err := resolveConcreteVersion(result.Version, cfg)
	if err != nil {
		logging.Error("failed to resolve concrete version", "err", err)
		return 1
	}

	// Write .terraform-version in the current directory.
	cwd, err := os.Getwd()
	if err != nil {
		logging.Error("failed to determine working directory", "err", err)
		return 1
	}

	path := filepath.Join(cwd, ".terraform-version")
	if err := os.WriteFile(path, []byte(version+"\n"), 0o644); err != nil {
		logging.Error("failed to write .terraform-version", "err", err)
		return 1
	}

	logging.Info(fmt.Sprintf("Pinned version by writing %q to %s", version, path))
	return 0
}

// resolveConcreteVersion resolves a version specifier to a concrete version
// string using locally installed versions. If the specifier is already an
// exact version, it is returned as-is.
func resolveConcreteVersion(specifier string, cfg *config.Config) (string, error) {
	// Exact versions pass through directly.
	if isExactVersion(specifier) {
		return specifier, nil
	}

	// Keyword specifiers need resolution against local versions.
	localVersions, err := list.ListLocal(cfg)
	if err != nil {
		return "", fmt.Errorf("listing local versions: %w", err)
	}

	if len(localVersions) == 0 {
		return "", fmt.Errorf(
			"no installed versions available to resolve %q. Please install one with: tfenv install",
			specifier)
	}

	resolved, err := resolve.ResolveVersion(specifier, localVersions, "local", cfg)
	if err != nil {
		return "", fmt.Errorf("resolving %q against local versions: %w", specifier, err)
	}

	return resolved.Version, nil
}

// isExactVersion returns true if the specifier looks like a concrete version
// number (starts with a digit).
func isExactVersion(specifier string) bool {
	if len(specifier) == 0 {
		return false
	}
	return specifier[0] >= '0' && specifier[0] <= '9'
}
