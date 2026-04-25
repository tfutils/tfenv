package use

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

// testConfig returns a Config pointing at the test directory with
// auto-install disabled by default.
func testConfig(configDir string) *config.Config {
	return &config.Config{
		ConfigDir:   configDir,
		AutoInstall: false,
	}
}

func TestUseExactVersion(t *testing.T) {
	dir := setupTestDir(t, "1.5.0", "1.6.0")
	cfg := testConfig(dir)

	specifier := "1.5.0"
	version, err := resolveInstalledVersion(specifier, cfg)
	if err != nil {
		t.Fatalf("resolveInstalledVersion(%q): %v", specifier, err)
	}
	if version != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", version)
	}

	// Write the version file.
	versionFile := filepath.Join(dir, "version")
	if err := atomicWriteFile(versionFile, version); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("reading version file: %v", err)
	}
	got := strings.TrimSpace(string(data))
	if got != "1.5.0" {
		t.Errorf("version file contains %q, want %q", got, "1.5.0")
	}
}

func TestUseLatestResolvesFromLocal(t *testing.T) {
	dir := setupTestDir(t, "1.4.0", "1.5.0", "1.6.0")
	cfg := testConfig(dir)

	version, err := resolveInstalledVersion("latest", cfg)
	if err != nil {
		t.Fatalf("resolveInstalledVersion(latest): %v", err)
	}
	if version != "1.6.0" {
		t.Errorf("expected 1.6.0, got %s", version)
	}
}

func TestUseLatestRegexResolvesFromLocal(t *testing.T) {
	dir := setupTestDir(t, "1.4.0", "1.5.1", "1.5.3", "1.6.0")
	cfg := testConfig(dir)

	version, err := resolveInstalledVersion("latest:^1\\.5", cfg)
	if err != nil {
		t.Fatalf("resolveInstalledVersion(latest:^1.5): %v", err)
	}
	if version != "1.5.3" {
		t.Errorf("expected 1.5.3, got %s", version)
	}
}

func TestUseDash(t *testing.T) {
	dir := setupTestDir(t, "1.5.0", "1.6.0")

	// Write version.prev.
	prevFile := filepath.Join(dir, "version.prev")
	if err := os.WriteFile(prevFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing version.prev: %v", err)
	}

	prev, err := readPreviousVersion(dir)
	if err != nil {
		t.Fatalf("readPreviousVersion: %v", err)
	}
	if prev != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", prev)
	}
}

func TestUseDashWithCarriageReturn(t *testing.T) {
	dir := setupTestDir(t)

	prevFile := filepath.Join(dir, "version.prev")
	if err := os.WriteFile(prevFile, []byte("1.5.0\r\n"), 0o644); err != nil {
		t.Fatalf("writing version.prev: %v", err)
	}

	prev, err := readPreviousVersion(dir)
	if err != nil {
		t.Fatalf("readPreviousVersion: %v", err)
	}
	if prev != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", prev)
	}
}

func TestUseDashErrorNoPrev(t *testing.T) {
	dir := setupTestDir(t)

	_, err := readPreviousVersion(dir)
	if err == nil {
		t.Fatal("expected error when no version.prev exists")
	}
	if !strings.Contains(err.Error(), "no previous version") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUseDashErrorEmptyPrev(t *testing.T) {
	dir := setupTestDir(t)

	prevFile := filepath.Join(dir, "version.prev")
	if err := os.WriteFile(prevFile, []byte("\n"), 0o644); err != nil {
		t.Fatalf("writing version.prev: %v", err)
	}

	_, err := readPreviousVersion(dir)
	if err == nil {
		t.Fatal("expected error when version.prev is empty")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNoArgsReadsVersionFile(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")
	cfg := testConfig(dir)

	// Set TFENV_TERRAFORM_VERSION in config to simulate env var.
	cfg.TerraformVersion = "1.5.0"

	specifier, source, err := resolveSpecifier(nil, cfg)
	if err != nil {
		t.Fatalf("resolveSpecifier: %v", err)
	}
	if specifier != "1.5.0" {
		t.Errorf("expected specifier 1.5.0, got %s", specifier)
	}
	if source != "TFENV_TERRAFORM_VERSION" {
		t.Errorf("expected source TFENV_TERRAFORM_VERSION, got %s", source)
	}
}

func TestTooManyArgsError(t *testing.T) {
	exitCode := Run([]string{"1.5.0", "1.6.0"})
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestNotInstalledAutoInstallOff(t *testing.T) {
	dir := setupTestDir(t) // No versions installed.
	cfg := testConfig(dir)
	cfg.AutoInstall = false

	_, err := resolveInstalledVersion("1.5.0", cfg)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "tfenv install") {
		t.Errorf("error should suggest tfenv install: %v", err)
	}
}

func TestVersionPrevUpdatedOnSwitch(t *testing.T) {
	dir := setupTestDir(t, "1.5.0", "1.6.0")

	// Write current version file.
	versionFile := filepath.Join(dir, "version")
	if err := os.WriteFile(versionFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing version file: %v", err)
	}

	// Simulate switch to 1.6.0 — save previous.
	if err := savePreviousVersion(versionFile, "1.6.0"); err != nil {
		t.Fatalf("savePreviousVersion: %v", err)
	}

	// Check version.prev was written.
	prevData, err := os.ReadFile(versionFile + ".prev")
	if err != nil {
		t.Fatalf("reading version.prev: %v", err)
	}
	got := strings.TrimSpace(string(prevData))
	if got != "1.5.0" {
		t.Errorf("version.prev contains %q, want %q", got, "1.5.0")
	}
}

func TestVersionPrevNotUpdatedWhenSameVersion(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")

	versionFile := filepath.Join(dir, "version")
	if err := os.WriteFile(versionFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing version file: %v", err)
	}

	// Switching to the same version should not create version.prev.
	if err := savePreviousVersion(versionFile, "1.5.0"); err != nil {
		t.Fatalf("savePreviousVersion: %v", err)
	}

	prevFile := versionFile + ".prev"
	if _, err := os.Stat(prevFile); !os.IsNotExist(err) {
		t.Errorf("version.prev should not exist when switching to same version")
	}
}

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	if err := atomicWriteFile(path, "hello"); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	if string(data) != "hello\n" {
		t.Errorf("file contains %q, want %q", string(data), "hello\n")
	}

	// Verify temp file was cleaned up.
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Errorf("temp file %s should not exist after successful write", tmpPath)
	}
}

func TestVerifyBinary(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")

	// Should succeed for an installed version with executable binary.
	if err := verifyBinary(dir, "1.5.0"); err != nil {
		t.Errorf("verifyBinary should succeed: %v", err)
	}
}

func TestVerifyBinaryMissing(t *testing.T) {
	dir := setupTestDir(t)

	// Create version dir without binary.
	vDir := filepath.Join(dir, "versions", "1.5.0")
	if err := os.MkdirAll(vDir, 0o755); err != nil {
		t.Fatalf("creating version dir: %v", err)
	}

	err := verifyBinary(dir, "1.5.0")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	if !strings.Contains(err.Error(), "binary is not") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyBinaryNotExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable bit not meaningful on Windows")
	}

	dir := setupTestDir(t)

	vDir := filepath.Join(dir, "versions", "1.5.0")
	if err := os.MkdirAll(vDir, 0o755); err != nil {
		t.Fatalf("creating version dir: %v", err)
	}

	// Write binary without executable permission.
	binaryPath := filepath.Join(vDir, "terraform")
	if err := os.WriteFile(binaryPath, []byte("fake"), 0o644); err != nil {
		t.Fatalf("writing binary: %v", err)
	}

	err := verifyBinary(dir, "1.5.0")
	if err == nil {
		t.Fatal("expected error for non-executable binary")
	}
	if !strings.Contains(err.Error(), "not executable") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsExactVersion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1.5.0", true},
		{"0.12.31", true},
		{"1.5.0-rc1", true},
		{"latest", false},
		{"latest:^1.5", false},
		{"latest-allowed", false},
		{"min-required", false},
		{"", false},
	}
	for _, tc := range tests {
		got := isExactVersion(tc.input)
		if got != tc.want {
			t.Errorf("isExactVersion(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestResolveSpecifierDash(t *testing.T) {
	dir := setupTestDir(t, "1.5.0")

	// Write version.prev.
	prevFile := filepath.Join(dir, "version.prev")
	if err := os.WriteFile(prevFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing version.prev: %v", err)
	}

	cfg := testConfig(dir)
	specifier, _, err := resolveSpecifier([]string{"-"}, cfg)
	if err != nil {
		t.Fatalf("resolveSpecifier(-): %v", err)
	}
	if specifier != "1.5.0" {
		t.Errorf("expected 1.5.0, got %s", specifier)
	}
}

func TestSavePreviousVersionNoExistingFile(t *testing.T) {
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "version")

	// No existing version file — should not error, should not create .prev.
	if err := savePreviousVersion(versionFile, "1.5.0"); err != nil {
		t.Fatalf("savePreviousVersion: %v", err)
	}

	prevFile := versionFile + ".prev"
	if _, err := os.Stat(prevFile); !os.IsNotExist(err) {
		t.Errorf("version.prev should not exist when no current version file")
	}
}

func TestLatestNoLocalMatchAutoInstallOff(t *testing.T) {
	dir := setupTestDir(t) // No versions installed.
	cfg := testConfig(dir)
	cfg.AutoInstall = false

	_, err := resolveInstalledVersion("latest", cfg)
	if err == nil {
		t.Fatal("expected error when no local versions match and auto-install is off")
	}
	if !strings.Contains(err.Error(), "TFENV_AUTO_INSTALL") {
		t.Errorf("error should mention TFENV_AUTO_INSTALL: %v", err)
	}
}
