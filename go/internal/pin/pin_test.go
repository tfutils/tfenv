package pin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/tfutils/tfenv/go/internal/config"
)

// setupTestDir creates a temporary config directory with the given installed
// versions. Each version gets a directory with a fake terraform binary.
func setupTestDir(t *testing.T, versions ...string) string {
	t.Helper()
	dir := t.TempDir()

	versionsDir := filepath.Join(dir, "versions")
	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		t.Fatalf("creating versions dir: %v", err)
	}

	for _, v := range versions {
		vDir := filepath.Join(versionsDir, v)
		if err := os.MkdirAll(vDir, 0o755); err != nil {
			t.Fatalf("creating version dir %s: %v", v, err)
		}

		binaryName := "terraform"
		if runtime.GOOS == "windows" {
			binaryName = "terraform.exe"
		}
		binaryPath := filepath.Join(vDir, binaryName)
		if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho fake"), 0o755); err != nil {
			t.Fatalf("writing fake binary %s: %v", binaryPath, err)
		}
	}

	return dir
}

// testConfig returns a Config pointing at the test directory.
func testConfig(configDir string) *config.Config {
	return &config.Config{
		ConfigDir:   configDir,
		AutoInstall: false,
	}
}

func TestPinRejectsArguments(t *testing.T) {
	exitCode := Run([]string{"1.5.0"})
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for extra arguments, got %d", exitCode)
	}
}

func TestPinErrorIfNoVersionsDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create a regular file that blocks config.Load from creating the
	// config directory tree (MkdirAll fails when a path component is a file).
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("block"), 0o644); err != nil {
		t.Fatalf("creating blocker file: %v", err)
	}
	t.Setenv("TFENV_CONFIG_DIR", filepath.Join(blocker, "config"))

	exitCode := Run(nil)
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when config dir cannot be created, got %d", exitCode)
	}
}

func TestPinErrorIfNoVersionResolvable(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")

	// No version file, no env var — resolution should fail.
	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "")

	// Change to a temp directory with no .terraform-version.
	workDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to %s: %v", workDir, err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Remove the default version file if it exists.
	os.Remove(filepath.Join(dir, "version"))

	exitCode := Run(nil)
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when no version resolvable, got %d", exitCode)
	}
}

func TestPinWritesCorrectContent(t *testing.T) {
	dir := setupTestDir(t, "1.5.0", "1.6.0")

	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.5.0")

	// Work in a temporary directory.
	workDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to %s: %v", workDir, err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	exitCode := Run(nil)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	versionFile := filepath.Join(workDir, ".terraform-version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("reading .terraform-version: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "1.5.0" {
		t.Errorf("expected .terraform-version to contain %q, got %q", "1.5.0", got)
	}
}

func TestPinOverwritesExistingFile(t *testing.T) {
	dir := setupTestDir(t, "1.5.0", "1.6.0")

	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.6.0")

	// Work in a temporary directory with an existing .terraform-version.
	workDir := t.TempDir()
	existingFile := filepath.Join(workDir, ".terraform-version")
	if err := os.WriteFile(existingFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing existing file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to %s: %v", workDir, err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	exitCode := Run(nil)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	data, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("reading .terraform-version: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "1.6.0" {
		t.Errorf("expected .terraform-version to contain %q, got %q", "1.6.0", got)
	}
}

func TestPinResolvesLatestKeyword(t *testing.T) {
	dir := setupTestDir(t, "1.4.0", "1.5.0", "1.6.0")

	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "latest")

	workDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("chdir to %s: %v", workDir, err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	exitCode := Run(nil)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	data, err := os.ReadFile(filepath.Join(workDir, ".terraform-version"))
	if err != nil {
		t.Fatalf("reading .terraform-version: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "1.6.0" {
		t.Errorf("expected .terraform-version to contain %q, got %q", "1.6.0", got)
	}
}

func TestResolveConcreteVersionExact(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")
	cfg := testConfig(dir)

	got, err := resolveConcreteVersion("1.5.0", cfg)
	if err != nil {
		t.Fatalf("resolveConcreteVersion(1.5.0): %v", err)
	}
	if got != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", got)
	}
}

func TestResolveConcreteVersionLatest(t *testing.T) {
	dir := setupTestDir(t, "1.4.0", "1.5.0", "1.6.0")
	cfg := testConfig(dir)

	got, err := resolveConcreteVersion("latest", cfg)
	if err != nil {
		t.Fatalf("resolveConcreteVersion(latest): %v", err)
	}
	if got != "1.6.0" {
		t.Errorf("expected 1.6.0, got %s", got)
	}
}

func TestResolveConcreteVersionLatestRegex(t *testing.T) {
	dir := setupTestDir(t, "1.4.0", "1.5.1", "1.5.3", "1.6.0")
	cfg := testConfig(dir)

	got, err := resolveConcreteVersion("latest:^1\\.5", cfg)
	if err != nil {
		t.Fatalf("resolveConcreteVersion(latest:^1.5): %v", err)
	}
	if got != "1.5.3" {
		t.Errorf("expected 1.5.3, got %s", got)
	}
}

func TestResolveConcreteVersionNoLocalVersions(t *testing.T) {
	dir := setupTestDir(t) // No versions installed.
	cfg := testConfig(dir)

	_, err := resolveConcreteVersion("latest", cfg)
	if err == nil {
		t.Fatal("expected error for latest with no installed versions")
	}
}
