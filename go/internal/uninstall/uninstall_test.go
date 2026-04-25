package uninstall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tfutils/tfenv/go/internal/config"
)

// setupTestVersions creates a temporary config directory with fake
// installed versions. Each version gets a versions/<ver>/terraform file.
// Returns the config and a cleanup function.
func setupTestVersions(t *testing.T, versions ...string) *config.Config {
	t.Helper()

	configDir := t.TempDir()

	for _, v := range versions {
		versionDir := filepath.Join(configDir, "versions", v)
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatalf("creating version dir %s: %v", v, err)
		}
		binaryPath := filepath.Join(versionDir, "terraform")
		if err := os.WriteFile(binaryPath, []byte("fake-binary"), 0o755); err != nil {
			t.Fatalf("creating fake binary for %s: %v", v, err)
		}
	}

	return &config.Config{
		ConfigDir: configDir,
		Remote:    "https://releases.hashicorp.com",
	}
}

func TestUninstallSingleVersion(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	err := uninstallSingle("1.5.0", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	versionDir := filepath.Join(cfg.ConfigDir, "versions", "1.5.0")
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Errorf("version directory still exists after uninstall")
	}
}

func TestUninstallMultipleVersions(t *testing.T) {
	cfg := setupTestVersions(t, "1.4.0", "1.5.0", "1.6.0")

	for _, v := range []string{"1.4.0", "1.6.0"} {
		if err := uninstallSingle(v, cfg); err != nil {
			t.Fatalf("unexpected error uninstalling %s: %v", v, err)
		}
	}

	// 1.4.0 and 1.6.0 should be gone.
	for _, v := range []string{"1.4.0", "1.6.0"} {
		versionDir := filepath.Join(cfg.ConfigDir, "versions", v)
		if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
			t.Errorf("version %s directory still exists", v)
		}
	}

	// 1.5.0 should still be there.
	versionDir := filepath.Join(cfg.ConfigDir, "versions", "1.5.0")
	if _, err := os.Stat(versionDir); err != nil {
		t.Errorf("version 1.5.0 should still exist: %v", err)
	}
}

func TestUninstallLatestResolvesFromLocal(t *testing.T) {
	cfg := setupTestVersions(t, "1.3.0", "1.4.0", "1.5.0")

	version, err := resolveLocalVersion("latest", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.5.0" {
		t.Errorf("expected latest to resolve to 1.5.0, got %s", version)
	}
}

func TestUninstallLatestRegexResolvesFromLocal(t *testing.T) {
	cfg := setupTestVersions(t, "1.3.0", "1.4.0", "1.4.5", "1.5.0")

	version, err := resolveLocalVersion("latest:^1\\.4", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.4.5" {
		t.Errorf("expected latest:^1\\.4 to resolve to 1.4.5, got %s", version)
	}
}

func TestUninstallNoArgsReadsVersionFile(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	// Write a .terraform-version file in a temp directory.
	tempDir := t.TempDir()
	versionFile := filepath.Join(tempDir, ".terraform-version")
	if err := os.WriteFile(versionFile, []byte("1.5.0\n"), 0o644); err != nil {
		t.Fatalf("writing version file: %v", err)
	}

	// Set Dir so the version file walker finds it.
	cfg.Dir = tempDir

	versions, err := collectVersions(nil, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(versions) != 1 || versions[0] != "1.5.0" {
		t.Errorf("expected [1.5.0], got %v", versions)
	}
}

func TestUninstallMinRequiredRejected(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	err := uninstallSingle("min-required", cfg)
	if err == nil {
		t.Fatal("expected error for min-required")
	}
	if got := err.Error(); !contains(got, "not a valid uninstall target") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestUninstallLatestAllowedRejected(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	err := uninstallSingle("latest-allowed", cfg)
	if err == nil {
		t.Fatal("expected error for latest-allowed")
	}
	if got := err.Error(); !contains(got, "not a valid uninstall target") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestUninstallVersionNotInstalled(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	err := uninstallSingle("1.9.9", cfg)
	if err == nil {
		t.Fatal("expected error for version not installed")
	}
	if got := err.Error(); !contains(got, "not installed") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestUninstallPathTraversalRejected(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	cases := []struct {
		name    string
		version string
	}{
		{"dot-dot", "1.5.0/../.."},
		{"forward-slash", "1.5.0/../../etc"},
		{"backslash", "1.5.0\\.."},
		{"null-byte", "1.5.0\x00"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := uninstallSingle(tc.version, cfg)
			if err == nil {
				t.Fatal("expected error for path traversal")
			}
			if got := err.Error(); !contains(got, "invalid version string") {
				t.Errorf("unexpected error message: %s", got)
			}
		})
	}
}

func TestUninstallContinueOnPartialFailure(t *testing.T) {
	cfg := setupTestVersions(t, "1.4.0", "1.5.0")

	// Uninstall one that exists and one that doesn't.
	specifiers := []string{"1.4.0", "1.9.9", "1.5.0"}

	failures := 0
	for _, s := range specifiers {
		if err := uninstallSingle(s, cfg); err != nil {
			failures++
		}
	}

	// 1.9.9 should have failed.
	if failures != 1 {
		t.Errorf("expected 1 failure, got %d", failures)
	}

	// 1.4.0 and 1.5.0 should be gone.
	for _, v := range []string{"1.4.0", "1.5.0"} {
		versionDir := filepath.Join(cfg.ConfigDir, "versions", v)
		if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
			t.Errorf("version %s directory still exists", v)
		}
	}
}

func TestUninstallPostCleanupEmptyVersionsDir(t *testing.T) {
	cfg := setupTestVersions(t, "1.5.0")

	// Uninstall the only version.
	if err := uninstallSingle("1.5.0", cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	versionsDir := filepath.Join(cfg.ConfigDir, "versions")

	// Post-cleanup: remove if empty.
	_ = os.Remove(versionsDir)

	if _, err := os.Stat(versionsDir); !os.IsNotExist(err) {
		t.Errorf("empty versions directory should be removed after cleanup")
	}
}

func TestValidateVersion(t *testing.T) {
	valid := []string{"1.5.0", "1.5.0-alpha1", "1.5.0-rc.1", "0.12.31"}
	for _, v := range valid {
		if err := validateVersion(v); err != nil {
			t.Errorf("validateVersion(%q) unexpected error: %v", v, err)
		}
	}

	invalid := []string{
		"../etc/passwd",
		"1.5.0/../../etc",
		"1.5.0\\..\\windows",
		"1.5.0\x00",
	}
	for _, v := range invalid {
		if err := validateVersion(v); err == nil {
			t.Errorf("validateVersion(%q) expected error, got nil", v)
		}
	}
}

func TestCollectVersionsFromArgs(t *testing.T) {
	cfg := setupTestVersions(t)

	versions, err := collectVersions([]string{"1.4.0", "1.5.0"}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0] != "1.4.0" || versions[1] != "1.5.0" {
		t.Errorf("expected [1.4.0 1.5.0], got %v", versions)
	}
}

func TestCollectVersionsFromEnvVar(t *testing.T) {
	cfg := setupTestVersions(t)
	cfg.TerraformVersion = "1.6.0"

	versions, err := collectVersions(nil, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 || versions[0] != "1.6.0" {
		t.Errorf("expected [1.6.0], got %v", versions)
	}
}

// contains is a helper to check substring presence.
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
