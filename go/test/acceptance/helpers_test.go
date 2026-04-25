package acceptance

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runTfenv executes the tfenv binary under test with the given arguments.
// It returns stdout, stderr, and the exit code. The test is NOT failed on
// non-zero exit — callers decide whether that is expected.
func runTfenv(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(tfenvBinary, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Provide a clean environment with only essentials.
	cmd.Env = cleanEnv(t)

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run %s: %v", tfenvBinary, err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// runTfenvWithEnv executes the tfenv binary with additional environment
// variables on top of the clean environment.
func runTfenvWithEnv(t *testing.T, env map[string]string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(tfenvBinary, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	baseEnv := cleanEnv(t)
	for k, v := range env {
		baseEnv = append(baseEnv, k+"="+v)
	}
	cmd.Env = baseEnv

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run %s: %v", tfenvBinary, err)
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// cleanEnv returns a minimal environment slice suitable for running the
// binary under test in isolation.
func cleanEnv(t *testing.T) []string {
	t.Helper()
	return []string{
		"PATH=" + filepath.Dir(tfenvBinary) + ":" + os.Getenv("PATH"),
		"HOME=" + t.TempDir(),
	}
}

// setupTempHome creates an isolated directory tree suitable for use as
// the TFENV_CONFIG_DIR. It returns the path to the root.
func setupTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Pre-create the versions directory so install tests have a target.
	versionsDir := filepath.Join(dir, "versions")
	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		t.Fatalf("failed to create versions dir: %v", err)
	}

	return dir
}

// assertExitCode fails the test if got != want.
func assertExitCode(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("exit code: got %d, want %d", got, want)
	}
}

// assertOutputContains fails the test if output does not contain substring.
func assertOutputContains(t *testing.T, output, substring string) {
	t.Helper()
	if !strings.Contains(output, substring) {
		t.Errorf("expected output to contain %q, got:\n%s", substring, output)
	}
}

// assertOutputNotContains fails the test if output contains substring.
func assertOutputNotContains(t *testing.T, output, substring string) {
	t.Helper()
	if strings.Contains(output, substring) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", substring, output)
	}
}

// assertFileExists fails the test if the path does not exist.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}
