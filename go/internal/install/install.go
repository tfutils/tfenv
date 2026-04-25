// Package install implements Terraform binary download, cryptographic
// verification (PGP signatures + SHA256 checksums), zip extraction, and
// atomic installation of Terraform binaries.
//
// The download pipeline:
//  1. Download SHA256SUMS and SHA256SUMS.sig
//  2. Verify PGP signature of SHA256SUMS against the embedded HashiCorp key
//  3. Download the platform-specific zip archive
//  4. Verify SHA256 checksum of the zip against the signed SHA256SUMS
//  5. Extract the terraform binary from the zip
//  6. Move to final ${ConfigDir}/versions/${version}/
//
// The HashiCorp PGP public key is embedded at compile time via go:embed,
// eliminating the Bash edition's fragile 3-tier verification fallback.
// Custom keys can be supplied via TFENV_PGP_KEY_PATH for private mirrors.
package install
