// Package use implements the tfenv use command — switching the default
// Terraform version.
package use

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/install"
	"github.com/tfutils/tfenv/go/internal/list"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

func init() {
	cli.Register("use", "Switch the default Terraform version", Run)
}

// Run implements the tfenv use command. It accepts zero or one argument:
//   - No args: resolve version from the version file chain
//   - One arg: use it as the version specifier
//   - "-": switch to the previous version
//
// Returns 0 on success, 1 on error.
func Run(args []string) int {
	if len(args) > 1 {
		logging.Error("usage: tfenv use [<version>]")
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	specifier, versionFileSource, err := resolveSpecifier(args, cfg)
	if err != nil {
		logging.Error("failed to determine version", "err", err)
		return 1
	}

	logging.Debug("resolved specifier", "specifier", specifier, "source", versionFileSource)

	// Resolve specifier to a concrete installed version.
	version, err := resolveInstalledVersion(specifier, cfg)
	if err != nil {
		logging.Error("failed to resolve version", "err", err)
		return 1
	}

	// Verify the binary exists and is executable.
	if err := verifyBinary(cfg.ConfigDir, version); err != nil {
		logging.Error("version verification failed", "err", err)
		return 1
	}

	// Save previous version before switching.
	versionFile := filepath.Join(cfg.ConfigDir, "version")
	if err := savePreviousVersion(versionFile, version); err != nil {
		logging.Warn("failed to save previous version", "err", err)
	}

	// Write new version atomically.
	if err := atomicWriteFile(versionFile, version); err != nil {
		logging.Error("failed to write version file", "err", err)
		return 1
	}

	logging.Info(fmt.Sprintf("Switching default version to v%s", version))

	// Warn if a .terraform-version file overrides the default.
	defaultVersionFile := versionFile
	if versionFileSource != "" && versionFileSource != defaultVersionFile {
		logging.Warn(fmt.Sprintf(
			"Default version file overridden by %s, changing the default version has no effect",
			versionFileSource))
	}

	logging.Info(fmt.Sprintf(
		"Default version (when not overridden by .terraform-version or TFENV_TERRAFORM_VERSION) is now: %s",
		version))

	return 0
}

// resolveSpecifier determines the version specifier from args or the
// version file chain. Returns the specifier, the source of the version
// file (for override detection), and any error.
func resolveSpecifier(args []string, cfg *config.Config) (string, string, error) {
	// Explicit argument provided.
	if len(args) == 1 {
		arg := args[0]

		// Handle "tfenv use -" — switch to previous version.
		if arg == "-" {
			prev, err := readPreviousVersion(cfg.ConfigDir)
			if err != nil {
				return "", "", err
			}
			logging.Info("Switching to previous version", "version", prev)
			return prev, "", nil
		}

		return arg, "", nil
	}

	// No argument — resolve from version file chain.
	result, err := resolve.ResolveVersionFile(cfg)
	if err != nil {
		return "", "", fmt.Errorf("no version specified and no version file found: %w", err)
	}

	return result.Version, result.Source, nil
}

// readPreviousVersion reads the previous version from version.prev.
func readPreviousVersion(configDir string) (string, error) {
	prevFile := filepath.Join(configDir, "version.prev")

	data, err := os.ReadFile(prevFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no previous version to switch to. Use tfenv use <version> first")
		}
		return "", fmt.Errorf("reading previous version file: %w", err)
	}

	version := strings.TrimSpace(strings.ReplaceAll(string(data), "\r", ""))
	if version == "" {
		return "", fmt.Errorf("previous version file is empty")
	}

	return version, nil
}

// resolveInstalledVersion resolves a specifier to a concrete version that
// is installed locally. For keywords like "latest", it searches local
// versions first and only falls back to remote + auto-install if enabled.
func resolveInstalledVersion(specifier string, cfg *config.Config) (string, error) {
	// Exact versions go straight to install check.
	if isExactVersion(specifier) {
		return ensureInstalled(specifier, specifier, cfg)
	}

	// For keywords, search locally installed versions first.
	localVersions, err := list.ListLocal(cfg)
	if err != nil {
		return "", fmt.Errorf("listing local versions: %w", err)
	}

	if len(localVersions) > 0 {
		resolved, err := resolve.ResolveVersion(specifier, localVersions, "local", cfg)
		if err == nil {
			logging.Debug("resolved from local versions", "version", resolved.Version)
			return resolved.Version, nil
		}
		logging.Debug("no local match for specifier", "specifier", specifier, "err", err)
	}

	// No local match — try remote + auto-install if enabled.
	if !cfg.AutoInstall {
		return "", fmt.Errorf(
			"no installed version matches %q and TFENV_AUTO_INSTALL is not enabled",
			specifier)
	}

	logging.Info("No installed versions match, attempting auto-install",
		"specifier", specifier)

	// Resolve via remote listing.
	remoteVersions, err := list.ListRemote(cfg)
	if err != nil {
		return "", fmt.Errorf("fetching remote versions: %w", err)
	}

	resolved, err := resolve.ResolveVersion(specifier, remoteVersions, "remote", cfg)
	if err != nil {
		return "", fmt.Errorf("resolving version from remote: %w", err)
	}

	return ensureInstalled(resolved.Version, specifier, cfg)
}

// ensureInstalled checks if the version is installed, triggering
// auto-install if configured. Returns the version string on success.
func ensureInstalled(version, specifier string, cfg *config.Config) (string, error) {
	if isVersionInstalled(cfg.ConfigDir, version) {
		return version, nil
	}

	if !cfg.AutoInstall {
		return "", fmt.Errorf(
			"version %s is not installed. Install it with: tfenv install %s",
			version, version)
	}

	logging.Info("Auto-installing missing version", "version", version)
	exitCode := install.Run([]string{version})
	if exitCode != 0 {
		return "", fmt.Errorf("auto-install of version %s failed", version)
	}

	// Verify install succeeded.
	if !isVersionInstalled(cfg.ConfigDir, version) {
		return "", fmt.Errorf(
			"version %s was installed but cannot be found — this should be impossible",
			version)
	}

	return version, nil
}

// isExactVersion returns true if the specifier looks like a concrete version
// (starts with a digit) rather than a keyword like "latest".
func isExactVersion(specifier string) bool {
	if len(specifier) == 0 {
		return false
	}
	return specifier[0] >= '0' && specifier[0] <= '9'
}

// isVersionInstalled checks whether a specific Terraform version is installed
// by looking for the terraform binary in the versions directory.
func isVersionInstalled(configDir, version string) bool {
	binaryName := "terraform"
	if runtime.GOOS == "windows" {
		binaryName = "terraform.exe"
	}
	binaryPath := filepath.Join(configDir, "versions", version, binaryName)
	info, err := os.Stat(binaryPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// verifyBinary checks that the terraform binary exists, is a regular file,
// and is executable.
func verifyBinary(configDir, version string) error {
	binaryName := "terraform"
	if runtime.GOOS == "windows" {
		binaryName = "terraform.exe"
	}
	binaryPath := filepath.Join(configDir, "versions", version, binaryName)

	info, err := os.Stat(binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"version directory for %s is present, but the terraform binary is not",
				version)
		}
		return fmt.Errorf("checking terraform binary: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("terraform path is a directory, not a binary: %s", binaryPath)
	}

	// On Unix, check the executable bit.
	if runtime.GOOS != "windows" {
		if info.Mode()&0o111 == 0 {
			return fmt.Errorf(
				"version directory for %s is present, but the terraform binary is not executable",
				version)
		}
	}

	return nil
}

// savePreviousVersion reads the current version file and saves it to
// version.prev if it differs from the new version being set.
func savePreviousVersion(versionFile, newVersion string) error {
	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No current version to save.
		}
		return fmt.Errorf("reading current version file: %w", err)
	}

	current := strings.TrimSpace(strings.ReplaceAll(string(data), "\r", ""))
	if current == "" || current == newVersion {
		return nil
	}

	logging.Debug("saving previous version", "version", current)
	prevFile := versionFile + ".prev"
	return atomicWriteFile(prevFile, current)
}

// atomicWriteFile writes content to path atomically by writing to a
// temporary file and renaming.
func atomicWriteFile(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing temporary file %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		// Clean up temp file on rename failure.
		os.Remove(tmp)
		return fmt.Errorf("renaming %s to %s: %w", tmp, path, err)
	}
	return nil
}
