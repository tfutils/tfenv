//go:build !windows

package shim

import (
	"fmt"
	"os"
	"syscall"
)

// execTerraform replaces the current process with the real terraform binary
// using syscall.Exec. On success this function never returns — the process
// image is replaced entirely. This avoids child-process overhead and ensures
// terraform receives signals (SIGINT, SIGTERM) directly.
func execTerraform(binaryPath string, args []string) int {
	argv := append([]string{binaryPath}, args...)
	env := os.Environ()

	if err := syscall.Exec(binaryPath, argv, env); err != nil {
		fmt.Fprintf(os.Stderr, "tfenv: exec %s: %v\n", binaryPath, err)
		return 1
	}

	// Unreachable on success — process is replaced.
	return 0
}
