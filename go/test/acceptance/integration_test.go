//go:build integration

// Package acceptance_integration contains integration tests that hit the real
// releases.hashicorp.com. These are excluded from the default test run and
// must be opted into via: go test -tags integration ./test/acceptance/...
//
// These tests require network access and are slow — they download real
// Terraform binaries.
package acceptance

// Integration tests will be added here as commands are implemented.
// Example:
//
//   func TestInstallRealBinary(t *testing.T) {
//       // Download and verify a real Terraform release.
//   }
