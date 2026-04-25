// Package install implements Terraform binary download, verification, and installation.
package install

import (
	_ "embed"
	"fmt"
	"os"

	"golang.org/x/crypto/openpgp"
)

// hashicorpPGPKey is the HashiCorp PGP public key, embedded at compile time.
// This eliminates the Bash edition's fragile 3-tier verification fallback
// (keybase → gpgv → gpg) and ensures PGP verification is always available.
//
//go:embed keys/hashicorp.pgp
var hashicorpPGPKey []byte

// loadPGPKeyring returns an openpgp.EntityList from the embedded HashiCorp PGP
// key, or from a custom key file if pgpKeyPath is non-empty.
func loadPGPKeyring(pgpKeyPath string) (openpgp.EntityList, error) {
	keyData := hashicorpPGPKey
	if pgpKeyPath != "" {
		custom, err := os.ReadFile(pgpKeyPath)
		if err != nil {
			return nil, fmt.Errorf("reading custom PGP key %q: %w", pgpKeyPath, err)
		}
		keyData = custom
	}
	return readPGPKey(keyData)
}
