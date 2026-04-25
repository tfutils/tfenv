//go:build windows

package shim

import (
	"fmt"
	"os"
	"os/exec"
)

// execTerraform runs the real terraform binary as a child process on Windows,
// piping stdin/stdout/stderr through. Windows does not support syscall.Exec
// (process image replacement), so we use os/exec and forward the exit code.
func execTerraform(binaryPath string, args []string) int {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr,
			"tfenv: exec %s: %v\n", binaryPath, err)
		return 1
	}

	return 0
}
