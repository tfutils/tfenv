package cli

import (
	"testing"
)

func TestRunVersion(t *testing.T) {
	exit := Run("1.2.3", []string{"--version"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunVersionSubcommand(t *testing.T) {
	exit := Run("1.2.3", []string{"version"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunHelp(t *testing.T) {
	exit := Run("1.2.3", []string{"help"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunNoArgs(t *testing.T) {
	exit := Run("1.2.3", []string{})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	exit := Run("1.2.3", []string{"unknown-command"})
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

	exit := Run("1.2.3", []string{"test-cmd"})
	if exit != 0 {
		t.Errorf("expected exit code 0, got %d", exit)
	}
}
