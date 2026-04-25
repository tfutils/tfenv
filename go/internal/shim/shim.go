// Package shim implements the Terraform shim, intercepting calls to the
// terraform binary and delegating to the correct installed version.
package shim

import (
	"fmt"
	"os"
)

// Run is the entry point for the terraform shim.
// It is invoked when the binary is called as "terraform" via symlink.
func Run(args []string) int {
	fmt.Fprintf(os.Stderr, "terraform shim not yet implemented\n")
	return 1
}
