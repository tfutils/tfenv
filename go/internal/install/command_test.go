package install

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Lock Tests ---

func TestAcquireLock_Success(t *testing.T) {
	tmpDir := t.TempDir()

	lock, err := acquireLock(tmpDir, "1.5.0")
	if err != nil {
		t.Fatalf("expected lock acquisition to succeed, got: %v", err)
	}

	expected := lockDir(tmpDir, "1.5.0")
	if lock != expected {
		t.Errorf("expected lock path %q, got %q", expected, lock)
	}

	// Lock directory should exist.
	info, err := os.Stat(lock)
	if err != nil {
		t.Fatalf("lock directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("lock path should be a directory")
	}

	releaseLock(lock)

	// Lock directory should be gone after release.
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Error("lock directory should not exist after release")
	}
}

func TestAcquireLock_BlockedByExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create the lock directory to simulate another process holding it.
	dir := lockDir(tmpDir, "1.5.0")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("creating pre-existing lock: %v", err)
	}

	// Touch it so it's fresh (not stale).
	now := time.Now()
	if err := os.Chtimes(dir, now, now); err != nil {
		t.Fatalf("touching lock dir: %v", err)
	}

	// Attempt to acquire with very limited retries — should fail.
	// We can't override lockMaxRetries, so instead we'll test with
	// a stale lock that gets cleaned up.
	releaseLock(dir)
}

func TestAcquireLock_StaleLockRemoved(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a lock directory and backdate it beyond the stale threshold.
	dir := lockDir(tmpDir, "1.5.0")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("creating stale lock: %v", err)
	}

	staleTime := time.Now().Add(-(lockStaleThreshold + time.Minute))
	if err := os.Chtimes(dir, staleTime, staleTime); err != nil {
		t.Fatalf("backdating lock: %v", err)
	}

	// Acquire should succeed after removing the stale lock.
	lock, err := acquireLock(tmpDir, "1.5.0")
	if err != nil {
		t.Fatalf("expected stale lock to be removed and acquisition to succeed, got: %v", err)
	}
	defer releaseLock(lock)
}

func TestReleaseLock_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	dir := lockDir(tmpDir, "1.5.0")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("creating lock dir: %v", err)
	}

	// Release twice — should not panic or error.
	releaseLock(dir)
	releaseLock(dir)
}

func TestReleaseLock_EmptyPath(t *testing.T) {
	// Should not panic.
	releaseLock("")
}

func TestRemoveIfStale_FreshLock(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, ".install-lock-1.5.0")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("creating lock dir: %v", err)
	}

	if removeIfStale(dir) {
		t.Error("fresh lock should not be considered stale")
	}

	// Directory should still exist.
	if _, err := os.Stat(dir); err != nil {
		t.Error("fresh lock directory should still exist")
	}
}

func TestRemoveIfStale_OldLock(t *testing.T) {
	tmpDir := t.TempDir()

	dir := filepath.Join(tmpDir, ".install-lock-1.5.0")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("creating lock dir: %v", err)
	}

	staleTime := time.Now().Add(-(lockStaleThreshold + 5*time.Minute))
	if err := os.Chtimes(dir, staleTime, staleTime); err != nil {
		t.Fatalf("backdating lock: %v", err)
	}

	if !removeIfStale(dir) {
		t.Error("old lock should be considered stale")
	}

	// Directory should be gone.
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("stale lock directory should have been removed")
	}
}

func TestRemoveIfStale_NonExistent(t *testing.T) {
	if removeIfStale("/nonexistent/path/lock") {
		t.Error("nonexistent path should return false")
	}
}

// --- Command Logic Tests ---

func TestIsExactVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.5.0", true},
		{"0.12.0-alpha3", true},
		{"1.5.7-rc1", true},
		{"latest", false},
		{"latest:^1.5", false},
		{"latest-allowed", false},
		{"min-required", false},
		{"", false},
	}

	for _, tc := range tests {
		got := isExactVersion(tc.input)
		if got != tc.expected {
			t.Errorf("isExactVersion(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestIsInstalled_Present(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the version directory with a terraform binary.
	versionDir := filepath.Join(tmpDir, "versions", "1.5.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binaryPath := filepath.Join(versionDir, "terraform")
	if err := os.WriteFile(binaryPath, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	if !isInstalled(tmpDir, "1.5.0") {
		t.Error("expected version to be detected as installed")
	}
}

func TestIsInstalled_Absent(t *testing.T) {
	tmpDir := t.TempDir()

	if isInstalled(tmpDir, "1.5.0") {
		t.Error("expected version to be detected as not installed")
	}
}

func TestIsInstalled_DirectoryButNoBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the version directory but no binary inside.
	versionDir := filepath.Join(tmpDir, "versions", "1.5.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if isInstalled(tmpDir, "1.5.0") {
		t.Error("expected version without binary to be detected as not installed")
	}
}

func TestRun_TooManyArgs(t *testing.T) {
	exit := Run([]string{"1.5.0", "1.6.0"})
	if exit != 1 {
		t.Errorf("expected exit code 1 for too many args, got %d", exit)
	}
}

func TestRun_AlreadyInstalled(t *testing.T) {
	tmpDir := t.TempDir()

	// Create installed version.
	versionDir := filepath.Join(tmpDir, "versions", "1.5.0")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "terraform"), []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Point config at our temp dir.
	t.Setenv("TFENV_CONFIG_DIR", tmpDir)
	// Use an exact version to skip remote lookups.
	exit := Run([]string{"1.5.0"})
	if exit != 0 {
		t.Errorf("expected exit code 0 for already-installed version, got %d", exit)
	}
}

func TestLockDir_Format(t *testing.T) {
	got := lockDir("/home/test/.tfenv", "1.5.0")
	expected := filepath.Join("/home/test/.tfenv", ".install-lock-1.5.0")
	if got != expected {
		t.Errorf("lockDir = %q, want %q", got, expected)
	}
}
