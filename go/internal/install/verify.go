package install

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

// readPGPKey parses a PGP public key from binary or ASCII-armored data.
func readPGPKey(data []byte) (openpgp.EntityList, error) {
	// Try binary first.
	entities, err := openpgp.ReadKeyRing(bytes.NewReader(data))
	if err == nil && len(entities) > 0 {
		return entities, nil
	}

	// Try ASCII-armored.
	block, err := armor.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("PGP key is neither valid binary nor armored format: %w", err)
	}
	entities, err = openpgp.ReadKeyRing(block.Body)
	if err != nil {
		return nil, fmt.Errorf("reading armored PGP key: %w", err)
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("PGP key file contains no keys")
	}
	return entities, nil
}

// VerifyPGPSignature verifies a detached PGP signature. data is the signed
// content, signature is the raw detached signature bytes, and publicKey is
// the PGP public key in binary or armored format.
func VerifyPGPSignature(data, signature, publicKey []byte) error {
	keyring, err := readPGPKey(publicKey)
	if err != nil {
		return fmt.Errorf("loading PGP public key: %w", err)
	}

	_, err = openpgp.CheckDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(signature))
	if err != nil {
		return fmt.Errorf("PGP signature verification failed: %w", err)
	}
	return nil
}

// VerifySHA256 verifies that the SHA256 hash of the file at filePath matches
// the expected hash for expectedFilename in the sha256sumsContent. The
// sha256sumsContent format is "hash  filename\n" per line (two-space separator).
func VerifySHA256(filePath string, sha256sumsContent []byte, expectedFilename string) error {
	expectedHash, err := findSHA256Entry(sha256sumsContent, expectedFilename)
	if err != nil {
		return err
	}

	actualHash, err := hashFile(filePath)
	if err != nil {
		return err
	}

	if actualHash != expectedHash {
		return fmt.Errorf(
			"SHA256 checksum mismatch for %s — expected %s, got %s",
			expectedFilename, expectedHash, actualHash,
		)
	}
	return nil
}

// findSHA256Entry parses the SHA256SUMS content and returns the hex hash for
// the given filename. Returns an error if no entry matches.
func findSHA256Entry(sha256sumsContent []byte, filename string) (string, error) {
	lines := strings.Split(strings.TrimSpace(string(sha256sumsContent)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "hash  filename" (two spaces).
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[1]) == filename {
			hash := strings.TrimSpace(parts[0])
			if len(hash) != 64 {
				return "", fmt.Errorf("invalid SHA256 hash length for %s: %d", filename, len(hash))
			}
			return hash, nil
		}
	}
	return "", fmt.Errorf("no SHA256SUMS entry found for %s", filename)
}

// hashFile computes the SHA256 hex digest of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file for SHA256: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("computing SHA256: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
