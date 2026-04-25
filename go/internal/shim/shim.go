// Package shim implements the Terraform shim, intercepting calls to the
// terraform binary, resolving the correct installed version, optionally
// auto-installing missing versions, and exec'ing the real binary.
//
// All shim output goes to stderr — stdout is reserved exclusively for
// terraform's own output (critical for terraform output, JSON mode, etc.).
package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/install"
	"github.com/tfutils/tfenv/go/internal/list"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

// Run is the entry point for the terraform shim.
// It is invoked when the binary is called as "terraform" via hardlink
// or copy (multi-call binary pattern).
func Run(args []string) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfenv: failed to load config: %v\n", err)
		return 1
	}

	// Detect -chdir=<dir> and set cfg.Dir so version file resolution
	// starts from the terraform working directory, not the shell cwd.
	if dir := extractChdir(args); dir != "" {
		cfg.Dir = dir
	}

	// Resolve the version specifier from the version file chain.
	result, err := resolve.ResolveVersionFile(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfenv: %v\n", err)
		return 1
	}

	logging.Debug("version file resolved",
		"version", result.Version, "source", result.Source)

	version := result.Version

	// If the specifier is a keyword (latest, latest:<regex>, etc.),
	// resolve it against locally installed versions first.
	if !isExactVersion(version) {
		version, err = resolveKeyword(version, result.Source, cfg)
		if err != nil {
			// No local match — auto-install or error.
			if !cfg.AutoInstall {
				fmt.Fprintf(os.Stderr,
					"tfenv: no installed version matches %q "+
						"and auto-install is disabled\n",
					result.Version)
				return 1
			}

			logging.Info("No local version matches, auto-installing",
				"specifier", result.Version)

			if install.Run([]string{result.Version}) != 0 {
				return 1
			}

			// Re-resolve against local versions after install.
			version, err = resolveKeyword(
				result.Version, result.Source, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"tfenv: failed to resolve version "+
						"after install: %v\n", err)
				return 1
			}
		}
	}

	// Determine the path to the real terraform binary.
	binaryPath := terraformBinaryPath(cfg.ConfigDir, version)

	// If the binary doesn't exist, auto-install or error.
	if _, statErr := os.Stat(binaryPath); os.IsNotExist(statErr) {
		if cfg.AutoInstall {
			logging.Info("Auto-installing Terraform",
				"version", version)
			if install.Run([]string{version}) != 0 {
				return 1
			}
		} else {
			fmt.Fprintf(os.Stderr,
				"Terraform v%s is not installed. "+
					"Install it with: tfenv install %s\n",
				version, version)
			return 1
		}
	}

	// Exec the real terraform binary, replacing this process on Unix.
	logging.Debug("exec terraform", "path", binaryPath, "args", args)
	return execTerraform(binaryPath, args)
}

// extractChdir scans terraform arguments for -chdir=<dir> and returns
// the directory path. Returns "" if -chdir is not present.
func extractChdir(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-chdir=") {
			return strings.TrimPrefix(arg, "-chdir=")
		}
	}
	return ""
}

// isExactVersion returns true if the version string looks like a concrete
// semver version (starts with a digit) rather than a keyword.
func isExactVersion(version string) bool {
	return len(version) > 0 && version[0] >= '0' && version[0] <= '9'
}

// resolveKeyword resolves a keyword specifier (latest, latest:<regex>,
// latest-allowed, min-required) against locally installed versions.
func resolveKeyword(
	specifier string,
	source string,
	cfg *config.Config,
) (string, error) {
	locals, err := list.ListLocal(cfg)
	if err != nil {
		return "", fmt.Errorf("listing local versions: %w", err)
	}

	if len(locals) == 0 {
		return "", fmt.Errorf(
			"no local versions installed for specifier %q",
			specifier)
	}

	resolved, err := resolve.ResolveVersion(
		specifier, locals, source, cfg)
	if err != nil {
		return "", err
	}

	return resolved.Version, nil
}

// terraformBinaryPath returns the full path to the terraform binary for
// a given version within the config directory.
func terraformBinaryPath(configDir, version string) string {
	name := "terraform"
	if runtime.GOOS == "windows" {
		name = "terraform.exe"
	}
	return filepath.Join(configDir, "versions", version, name)
}
