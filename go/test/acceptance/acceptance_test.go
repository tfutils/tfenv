// Package acceptance provides shared acceptance test infrastructure for tfenv.
//
// Tests invoke whichever binary TFENV_TEST_BINARY points to, enabling parity
// testing between the Bash and Go editions.
package acceptance

import (
	"fmt"
	"os"
	"testing"
)

// tfenvBinary holds the path to the tfenv binary under test.
// It is set once in TestMain from the TFENV_TEST_BINARY env var.
var tfenvBinary string

func TestMain(m *testing.M) {
	tfenvBinary = os.Getenv("TFENV_TEST_BINARY")
	if tfenvBinary == "" {
		fmt.Println("TFENV_TEST_BINARY not set — skipping acceptance tests")
		fmt.Println("To run: TFENV_TEST_BINARY=/path/to/binary go test ./test/acceptance/... -v")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
