package shim

import (
	"testing"
)

func TestRunReturnsOne(t *testing.T) {
	exit := Run([]string{"version"})
	if exit != 1 {
		t.Errorf("expected exit code 1 (stub not implemented), got %d", exit)
	}
}
