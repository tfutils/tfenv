package cli

import (
	"testing"
)

var testInfo = BuildInfo{Version: "1.2.3", Commit: "abc123", Date: "2025-01-01"}

func TestRunVersion(t *testing.T) {
	exit := Run(testInfo, []string{"--version"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunVersionSubcommand(t *testing.T) {
	exit := Run(testInfo, []string{"version"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunHelp(t *testing.T) {
	exit := Run(testInfo, []string{"help"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunNoArgs(t *testing.T) {
	exit := Run(testInfo, []string{})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	exit := Run(testInfo, []string{"unknown-command"})
	if exit != 1 {
		t.Errorf("expected exit code 1, got %d", exit)
	}
}

func TestRegisterAndRun(t *testing.T) {
	Register("test-cmd", "A test command", func(args []string) int {
		return 0
	})
	defer func() {
		delete(registry, "test-cmd")
	}()

	exit := Run(testInfo, []string{"test-cmd"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}
