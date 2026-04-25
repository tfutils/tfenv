// Package config handles environment variable loading and state directory resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config aggregates all resolved configuration from TFENV_* environment variables.
// Loaded once at startup and passed to commands.
type Config struct {
	ConfigDir          string // TFENV_CONFIG_DIR — state directory (default: ~/.tfenv/)
	Root               string // TFENV_ROOT — Bash edition install dir (read-only, for compat)
	Remote             string // TFENV_REMOTE — mirror URL (default: https://releases.hashicorp.com)
	Arch               string // TFENV_ARCH — architecture override
	AutoInstall        bool   // TFENV_AUTO_INSTALL — auto-install on exec (default: true)
	CurlOutput         int    // TFENV_CURL_OUTPUT — download progress verbosity (0=quiet, 1=progress, 2=verbose)
	TerraformVersion   string // TFENV_TERRAFORM_VERSION — version override (highest precedence)
	NetrcPath          string // TFENV_NETRC_PATH — custom .netrc path
	ReverseRemote      bool   // TFENV_REVERSE_REMOTE — reverse remote listing order
	SortVersionsRemote bool   // TFENV_SORT_VERSIONS_REMOTE — sort remote by semver
	SkipRemoteCheck    bool   // TFENV_SKIP_REMOTE_CHECK — skip remote during install
	Dir                string // TFENV_DIR — override start dir for .terraform-version walk
	PGPKeyPath         string // TFENV_PGP_KEY_PATH — custom PGP key for private mirrors
	LogLevel           string // TFENV_LOG_LEVEL — logging level
	LogFormat          string // TFENV_LOG_FORMAT — logging format (text/json)
}

// Load reads all TFENV_* environment variables, applies defaults, and returns
// a fully resolved Config. It auto-creates ConfigDir and ConfigDir/versions/
// if they don't exist. Returns an error only for genuinely invalid config
// (e.g. unparseable int for TFENV_CURL_OUTPUT).
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home directory: %w", err)
	}

	configDir := os.Getenv("TFENV_CONFIG_DIR")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".tfenv")
	}

	autoInstall, err := parseBool(os.Getenv("TFENV_AUTO_INSTALL"), true)
	if err != nil {
		return nil, fmt.Errorf("parsing TFENV_AUTO_INSTALL: %w", err)
	}

	curlOutput, err := parseInt(os.Getenv("TFENV_CURL_OUTPUT"), 0)
	if err != nil {
		return nil, fmt.Errorf("parsing TFENV_CURL_OUTPUT: %w", err)
	}

	remote := os.Getenv("TFENV_REMOTE")
	if remote == "" {
		remote = "https://releases.hashicorp.com"
	}

	reverseRemote, err := parseBool(os.Getenv("TFENV_REVERSE_REMOTE"), false)
	if err != nil {
		return nil, fmt.Errorf("parsing TFENV_REVERSE_REMOTE: %w", err)
	}

	sortVersionsRemote, err := parseBool(os.Getenv("TFENV_SORT_VERSIONS_REMOTE"), false)
	if err != nil {
		return nil, fmt.Errorf("parsing TFENV_SORT_VERSIONS_REMOTE: %w", err)
	}

	skipRemoteCheck, err := parseBool(os.Getenv("TFENV_SKIP_REMOTE_CHECK"), false)
	if err != nil {
		return nil, fmt.Errorf("parsing TFENV_SKIP_REMOTE_CHECK: %w", err)
	}

	cfg := &Config{
		ConfigDir:          configDir,
		Root:               os.Getenv("TFENV_ROOT"),
		Remote:             remote,
		Arch:               os.Getenv("TFENV_ARCH"),
		AutoInstall:        autoInstall,
		CurlOutput:         curlOutput,
		TerraformVersion:   os.Getenv("TFENV_TERRAFORM_VERSION"),
		NetrcPath:          os.Getenv("TFENV_NETRC_PATH"),
		ReverseRemote:      reverseRemote,
		SortVersionsRemote: sortVersionsRemote,
		SkipRemoteCheck:    skipRemoteCheck,
		Dir:                os.Getenv("TFENV_DIR"),
		PGPKeyPath:         os.Getenv("TFENV_PGP_KEY_PATH"),
		LogLevel:           os.Getenv("TFENV_LOG_LEVEL"),
		LogFormat:          os.Getenv("TFENV_LOG_FORMAT"),
	}

	if err := ensureDir(cfg.ConfigDir); err != nil {
		return nil, fmt.Errorf("creating config directory %q: %w", cfg.ConfigDir, err)
	}

	versionsDir := filepath.Join(cfg.ConfigDir, "versions")
	if err := ensureDir(versionsDir); err != nil {
		return nil, fmt.Errorf("creating versions directory %q: %w", versionsDir, err)
	}

	return cfg, nil
}

// parseBool parses a string as a boolean, accepting true/false/1/0 (case-insensitive).
// Returns the defaultVal if the input is empty.
func parseBool(s string, defaultVal bool) (bool, error) {
	if s == "" {
		return defaultVal, nil
	}
	switch strings.ToLower(s) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %q (expected true/false/1/0)", s)
	}
}

// parseInt parses a string as an integer. Returns the defaultVal if the input is empty.
func parseInt(s string, defaultVal int) (int, error) {
	if s == "" {
		return defaultVal, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %q", s)
	}
	return v, nil
}

// ensureDir creates a directory (and parents) if it doesn't already exist.
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
