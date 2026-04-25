package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/list"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/platform"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

func init() {
	cli.Register("install", "Install a specific version of Terraform", Run)
}

// Run implements the tfenv install command. It accepts zero or one argument:
//   - No args: resolve version from the version file chain
//   - One arg: use it as the version specifier
//
// Returns 0 on success, 1 on error.
func Run(args []string) int {
	if len(args) > 1 {
		logging.Error("usage: tfenv install [<version>]")
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	// Determine version specifier.
	specifier, source, err := resolveSpecifier(args, cfg)
	if err != nil {
		logging.Error("failed to determine version", "err", err)
		return 1
	}

	logging.Info("Resolving version", "specifier", specifier, "source", source)

	// Resolve specifier to a concrete version.
	version, err := resolveConcreteVersion(specifier, source, cfg)
	if err != nil {
		logging.Error("failed to resolve version", "err", err)
		return 1
	}

	logging.Info("Resolved version", "version", version)

	// Check if already installed.
	if isInstalled(cfg.ConfigDir, version) {
		logging.Info("Terraform already installed", "version", version)
		return 0
	}

	// Acquire install lock.
	lock, err := acquireLock(cfg.ConfigDir, version)
	if err != nil {
		logging.Error("failed to acquire install lock", "err", err)
		return 1
	}
	defer releaseLock(lock)

	// Double-check after acquiring lock — another process may have finished.
	if isInstalled(cfg.ConfigDir, version) {
		logging.Info("Terraform already installed (after lock)", "version", version)
		return 0
	}

	// Detect platform.
	plat := platform.Detect(cfg.Arch)
	arch := plat.DownloadArch(version)
	logging.Info("Detected platform",
		"os", plat.OS, "arch", arch, "native_arch", runtime.GOARCH)

	// Download, verify, and extract.
	ctx := context.Background()
	result, err := Download(ctx, version, plat, cfg)
	if err != nil {
		logging.Error("download failed", "err", err)
		return 1
	}

	installPath, err := Install(result, cfg.ConfigDir)
	if err != nil {
		logging.Error("installation failed", "err", err)
		return 1
	}

	logging.Info("Installation complete",
		"version", version, "path", installPath)
	return 0
}

// resolveSpecifier determines the version specifier from args or version file.
func resolveSpecifier(args []string, cfg *config.Config) (string, string, error) {
	if len(args) == 1 {
		return args[0], "command-line argument", nil
	}

	result, err := resolve.ResolveVersionFile(cfg)
	if err != nil {
		return "", "", fmt.Errorf("no version specified and no version file found: %w", err)
	}

	return result.Version, result.Source, nil
}

// resolveConcreteVersion resolves a specifier (which may be "latest",
// "latest:<regex>", "latest-allowed", "min-required", or an exact version)
// to a concrete version string.
func resolveConcreteVersion(specifier, source string, cfg *config.Config) (string, error) {
	// For exact versions, skip the remote listing unless we need to verify
	// the version exists remotely.
	if isExactVersion(specifier) && !cfg.SkipRemoteCheck {
		// If skip-remote-check is not set, we still resolve through the
		// normal path to validate the version exists, but we could also
		// just trust the user. For now, return it directly — the download
		// step will fail with a clear error if the version doesn't exist.
		return specifier, nil
	}

	if isExactVersion(specifier) {
		return specifier, nil
	}

	// Non-exact specifiers need the remote version list.
	versions, err := list.ListRemote(cfg)
	if err != nil {
		return "", fmt.Errorf("fetching remote versions: %w", err)
	}

	resolved, err := resolve.ResolveVersion(specifier, versions, source, cfg)
	if err != nil {
		return "", err
	}

	return resolved.Version, nil
}

// isExactVersion returns true if the specifier looks like a concrete version
// (digits.digits.digits with optional pre-release suffix) rather than a
// keyword like "latest".
func isExactVersion(specifier string) bool {
	if len(specifier) == 0 {
		return false
	}
	return specifier[0] >= '0' && specifier[0] <= '9'
}

// isInstalled checks whether a specific Terraform version is already installed
// by looking for the terraform binary in the versions directory.
func isInstalled(configDir, version string) bool {
	binaryName := "terraform"
	if runtime.GOOS == "windows" {
		binaryName = "terraform.exe"
	}
	binaryPath := filepath.Join(configDir, "versions", version, binaryName)
	_, err := os.Stat(binaryPath)
	return err == nil
}
