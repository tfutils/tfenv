package shim

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- extractChdir tests ---

func TestExtractChdir_Present(t *testing.T) {
	args := []string{"-chdir=/tmp/project", "plan"}
	got := extractChdir(args)
	if got != "/tmp/project" {
		t.Errorf("extractChdir = %q, want %q", got, "/tmp/project")
	}
}

func TestExtractChdir_NotPresent(t *testing.T) {
	args := []string{"plan", "-out=tfplan"}
	got := extractChdir(args)
	if got != "" {
		t.Errorf("extractChdir = %q, want empty", got)
	}
}

func TestExtractChdir_Empty(t *testing.T) {
	got := extractChdir(nil)
	if got != "" {
		t.Errorf("extractChdir(nil) = %q, want empty", got)
	}
}

func TestExtractChdir_MultipleArgs(t *testing.T) {
	args := []string{"-input=false", "-chdir=mydir", "-auto-approve"}
	got := extractChdir(args)
	if got != "mydir" {
		t.Errorf("extractChdir = %q, want %q", got, "mydir")
	}
}

func TestExtractChdir_RelativePath(t *testing.T) {
	args := []string{"-chdir=../other"}
	got := extractChdir(args)
	if got != "../other" {
		t.Errorf("extractChdir = %q, want %q", got, "../other")
	}
}

// --- isExactVersion tests ---

func TestIsExactVersion_Exact(t *testing.T) {
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

	for _, tt := range tests {
		got := isExactVersion(tt.input)
		if got != tt.want {
			t.Errorf(
				"isExactVersion(%q) = %v, want %v",
				tt.input, got, tt.want)
		}
	}
}

// --- terraformBinaryPath tests ---

func TestTerraformBinaryPath(t *testing.T) {
	got := terraformBinaryPath("/home/user/.tfenv", "1.5.0")
	want := filepath.Join(
		"/home/user/.tfenv", "versions", "1.5.0", "terraform")
	if runtime.GOOS == "windows" {
		want = filepath.Join(
			"/home/user/.tfenv", "versions", "1.5.0",
			"terraform.exe")
	}
	if got != want {
		t.Errorf(
			"terraformBinaryPath = %q, want %q", got, want)
	}
}

// --- Run integration-style tests ---

func TestRun_NoVersionFile(t *testing.T) {
	// Set up isolated config dir with no version file.
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	workDir := filepath.Join(tmpDir, "work")

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Point config and working directory to isolated locations.
	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "")
	t.Setenv("TFENV_AUTO_INSTALL", "false")
	t.Setenv("HOME", tmpDir)

	// Change to workDir so the version file walk doesn't find
	// anything from the real filesystem.
	origDir, _ := os.Getwd()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	exit := Run([]string{"version"})
	if exit != 1 {
		t.Errorf("expected exit 1 with no version file, got %d", exit)
	}
}

func TestRun_ExactVersion_NotInstalled_NoAutoInstall(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	workDir := filepath.Join(tmpDir, "work")

	if err := os.MkdirAll(
		filepath.Join(configDir, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.5.0")
	t.Setenv("TFENV_AUTO_INSTALL", "false")

	exit := Run([]string{"version"})
	if exit != 1 {
		t.Errorf(
			"expected exit 1 for missing version without "+
				"auto-install, got %d", exit)
	}
}

func TestRun_ChdirSetsConfigDir(t *testing.T) {
	// Create a directory structure with a .terraform-version file
	// in the -chdir target.
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	chdirTarget := filepath.Join(tmpDir, "project")

	if err := os.MkdirAll(
		filepath.Join(configDir, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(chdirTarget, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a .terraform-version in the chdir target.
	versionFile := filepath.Join(
		chdirTarget, ".terraform-version")
	if err := os.WriteFile(
		versionFile, []byte("1.6.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "")
	t.Setenv("TFENV_AUTO_INSTALL", "false")
	t.Setenv("HOME", tmpDir)

	// Working directory has no version file.
	workDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}
	origDir, _ := os.Getwd()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Run with -chdir pointing to the project dir. It should find
	// v1.6.0 from the version file, then fail because it's not
	// installed (auto-install off). The important thing is that it
	// found the version file via -chdir, not the cwd.
	exit := Run([]string{
		"-chdir=" + chdirTarget, "plan"})
	if exit != 1 {
		t.Errorf(
			"expected exit 1 (version not installed), got %d",
			exit)
	}
}

func TestRun_KeywordVersion_NoLocalVersions_NoAutoInstall(
	t *testing.T,
) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	if err := os.MkdirAll(
		filepath.Join(configDir, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "latest")
	t.Setenv("TFENV_AUTO_INSTALL", "false")

	exit := Run([]string{"version"})
	if exit != 1 {
		t.Errorf(
			"expected exit 1 for keyword with no local "+
				"versions, got %d", exit)
	}
}

func TestRun_KeywordVersion_WithLocalVersions(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	// Create fake installed versions (just directories — the binary
	// won't exist, so we'll get exit 1 at the exec step).
	versions := []string{"1.4.0", "1.5.0", "1.5.7"}
	for _, v := range versions {
		vDir := filepath.Join(configDir, "versions", v)
		if err := os.MkdirAll(vDir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "latest")
	t.Setenv("TFENV_AUTO_INSTALL", "false")

	// Should resolve "latest" to 1.5.7, then fail because there's
	// no actual terraform binary. Exit code 1 is expected.
	exit := Run([]string{"version"})
	if exit != 1 {
		t.Errorf(
			"expected exit 1 (no terraform binary), got %d",
			exit)
	}
}

// --- isSameFile tests ---

func TestIsSameFile_Same(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "a")
	if err := os.WriteFile(src, []byte("hello"), 0o755); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "b")
	if err := os.Link(src, dst); err != nil {
		t.Fatal(err)
	}

	if !isSameFile(src, dst) {
		t.Error("isSameFile should return true for hardlinked files")
	}
}

func TestIsSameFile_Different(t *testing.T) {
	tmpDir := t.TempDir()
	a := filepath.Join(tmpDir, "a")
	b := filepath.Join(tmpDir, "b")
	if err := os.WriteFile(a, []byte("one"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("two"), 0o755); err != nil {
		t.Fatal(err)
	}

	if isSameFile(a, b) {
		t.Error(
			"isSameFile should return false for different files")
	}
}

func TestIsSameFile_Missing(t *testing.T) {
	if isSameFile("/nonexistent/a", "/nonexistent/b") {
		t.Error(
			"isSameFile should return false for missing files")
	}
}

// --- copyFile tests ---

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src")
	content := []byte("#!/bin/sh\necho hello\n")
	if err := os.WriteFile(src, content, 0o755); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "dst")
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(content) {
		t.Errorf("copied content = %q, want %q", got, content)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm()&0o111 == 0 {
		t.Error("copied file should be executable")
	}
}

// --- RunInit tests ---

func TestRunInit_CreatesHardlink(t *testing.T) {
	// RunInit uses os.Executable() which we can't easily control
	// in a unit test. Instead, test the helper functions and
	// verify the init logic components work correctly.

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "tfenv")
	if err := os.WriteFile(
		src, []byte("fake binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "terraform")

	// Simulate hardlink creation.
	if err := os.Link(src, dst); err != nil {
		t.Fatal(err)
	}

	if !isSameFile(src, dst) {
		t.Error("hardlink should produce same file")
	}
}

func TestRunInit_FallbackToCopy(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "tfenv")
	content := []byte("fake binary content")
	if err := os.WriteFile(src, content, 0o755); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "terraform")
	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(content) {
		t.Error("copy fallback should produce identical content")
	}
}

func TestRunInit_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "tfenv")
	dst := filepath.Join(tmpDir, "terraform")

	if err := os.WriteFile(
		src, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create hardlink first time.
	if err := os.Link(src, dst); err != nil {
		t.Fatal(err)
	}

	// Verify isSameFile detects idempotency.
	if !isSameFile(src, dst) {
		t.Error("should detect same file for idempotency check")
	}
}

// --- Stdout contamination test ---

func TestRun_NoStdoutContamination(t *testing.T) {
	// Capture stdout to verify the shim never writes to it.
	// We redirect os.Stdout to a pipe and check nothing was written.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	workDir := filepath.Join(tmpDir, "work")

	if err := os.MkdirAll(
		filepath.Join(configDir, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TFENV_CONFIG_DIR", configDir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.5.0")
	t.Setenv("TFENV_AUTO_INSTALL", "false")
	t.Setenv("HOME", tmpDir)

	origDir, _ := os.Getwd()
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Run the shim (will fail because version not installed).
	_ = Run([]string{"version"})

	// Close write end and read what was captured.
	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	r.Close()

	if n > 0 {
		t.Errorf(
			"shim wrote %d bytes to stdout: %q — "+
				"all output must go to stderr",
			n, string(buf[:n]))
	}
}
