package install

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/platform"
)

// --- Test helpers ---

// testPGPEntity generates an ephemeral PGP key for testing.
func testPGPEntity(t *testing.T) *openpgp.Entity {
	t.Helper()
	entity, err := openpgp.NewEntity("tfenv-test", "unit test key", "test@tfenv.test", &packet.Config{
		DefaultHash: crypto.SHA256,
	})
	if err != nil {
		t.Fatalf("generating test PGP key: %v", err)
	}
	return entity
}

// armorPubKey exports the entity's public key in ASCII-armored format.
func armorPubKey(t *testing.T, entity *openpgp.Entity) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := armor.Encode(&buf, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("creating armor encoder: %v", err)
	}
	if err := entity.Serialize(w); err != nil {
		t.Fatalf("serializing public key: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing armor encoder: %v", err)
	}
	return buf.Bytes()
}

// binaryPubKey exports the entity's public key in binary format.
func binaryPubKey(t *testing.T, entity *openpgp.Entity) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := entity.Serialize(&buf); err != nil {
		t.Fatalf("serializing binary public key: %v", err)
	}
	return buf.Bytes()
}

// detachSign creates a detached signature of data using the entity's private key.
func detachSign(t *testing.T, entity *openpgp.Entity, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := openpgp.DetachSign(&buf, entity, bytes.NewReader(data), &packet.Config{
		DefaultHash: crypto.SHA256,
	}); err != nil {
		t.Fatalf("creating detached signature: %v", err)
	}
	return buf.Bytes()
}

// buildTestZip creates a zip archive containing a stub terraform binary.
func buildTestZip(t *testing.T, version string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	stub := fmt.Sprintf("#!/bin/sh\necho \"Terraform v%s\"\n", version)
	hdr := &zip.FileHeader{
		Name:   "terraform",
		Method: zip.Deflate,
	}
	hdr.SetMode(0o755)
	fw, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatalf("creating zip entry: %v", err)
	}
	if _, err := fw.Write([]byte(stub)); err != nil {
		t.Fatalf("writing zip entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zip writer: %v", err)
	}
	return buf.Bytes()
}

// buildTestSHA256SUMS builds a SHA256SUMS file from filename→zipData pairs.
func buildTestSHA256SUMS(entries map[string][]byte) string {
	type entry struct {
		name string
		hash string
	}
	var sorted []entry
	for name, data := range entries {
		h := sha256.Sum256(data)
		sorted = append(sorted, entry{name: name, hash: fmt.Sprintf("%x", h)})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].name < sorted[j].name })
	var lines []string
	for _, e := range sorted {
		lines = append(lines, fmt.Sprintf("%s  %s", e.hash, e.name))
	}
	return strings.Join(lines, "\n") + "\n"
}

// mockReleaseServer creates an httptest.Server that serves HashiCorp-style
// release artifacts for the given version and platform, signed by the given
// PGP entity.
type mockRelease struct {
	version  string
	platStr  string
	zipData  []byte
	entity   *openpgp.Entity
	tampered bool // serve a modified SHA256SUMS sig to test failure
}

func startMockServer(t *testing.T, releases []mockRelease) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	for _, rel := range releases {
		rel := rel // capture

		zipFilename := buildZipFilename(rel.version, rel.platStr)
		sumsFilename := fmt.Sprintf("terraform_%s_SHA256SUMS", rel.version)
		sigFilename := fmt.Sprintf("%s%s.sig", sumsFilename, sigKeyPostfix)

		entries := map[string][]byte{zipFilename: rel.zipData}
		sumsContent := buildTestSHA256SUMS(entries)

		// Sign the SHA256SUMS content.
		var sigData []byte
		if rel.tampered {
			sigData = detachSign(t, rel.entity, []byte("tampered content"))
		} else {
			sigData = detachSign(t, rel.entity, []byte(sumsContent))
		}

		prefix := fmt.Sprintf("/terraform/%s/", rel.version)
		mux.HandleFunc(prefix+sumsFilename, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(sumsContent))
		})
		mux.HandleFunc(prefix+sigFilename, func(w http.ResponseWriter, r *http.Request) {
			w.Write(sigData)
		})
		mux.HandleFunc(prefix+zipFilename, func(w http.ResponseWriter, r *http.Request) {
			w.Write(rel.zipData)
		})
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// --- PGP Verification Tests ---

func TestVerifyPGPSignature_Valid(t *testing.T) {
	entity := testPGPEntity(t)
	pubKey := armorPubKey(t, entity)

	data := []byte("hello world sha256sums content")
	sig := detachSign(t, entity, data)

	if err := VerifyPGPSignature(data, sig, pubKey); err != nil {
		t.Fatalf("expected valid PGP signature to pass, got: %v", err)
	}
}

func TestVerifyPGPSignature_ValidBinaryKey(t *testing.T) {
	entity := testPGPEntity(t)
	pubKey := binaryPubKey(t, entity)

	data := []byte("some signed content")
	sig := detachSign(t, entity, data)

	if err := VerifyPGPSignature(data, sig, pubKey); err != nil {
		t.Fatalf("expected valid PGP signature with binary key to pass, got: %v", err)
	}
}

func TestVerifyPGPSignature_InvalidSignature(t *testing.T) {
	entity := testPGPEntity(t)
	pubKey := armorPubKey(t, entity)

	data := []byte("original content")
	sig := detachSign(t, entity, []byte("different content"))

	if err := VerifyPGPSignature(data, sig, pubKey); err == nil {
		t.Fatal("expected PGP verification to fail with wrong signature")
	}
}

func TestVerifyPGPSignature_WrongKey(t *testing.T) {
	signer := testPGPEntity(t)
	wrongKey := testPGPEntity(t)
	pubKey := armorPubKey(t, wrongKey)

	data := []byte("content signed by different key")
	sig := detachSign(t, signer, data)

	if err := VerifyPGPSignature(data, sig, pubKey); err == nil {
		t.Fatal("expected PGP verification to fail with wrong key")
	}
}

func TestVerifyPGPSignature_CorruptedSignature(t *testing.T) {
	entity := testPGPEntity(t)
	pubKey := armorPubKey(t, entity)

	data := []byte("content")
	if err := VerifyPGPSignature(data, []byte("not a signature"), pubKey); err == nil {
		t.Fatal("expected PGP verification to fail with corrupted signature")
	}
}

func TestVerifyPGPSignature_InvalidKey(t *testing.T) {
	data := []byte("content")
	sig := []byte("sig")
	if err := VerifyPGPSignature(data, sig, []byte("not a key")); err == nil {
		t.Fatal("expected error with invalid key data")
	}
}

// --- Embedded HashiCorp Key Test ---

func TestEmbeddedHashiCorpKey_NotEmpty(t *testing.T) {
	if len(hashicorpPGPKey) == 0 {
		t.Fatal("embedded HashiCorp PGP key is empty")
	}
}

func TestEmbeddedHashiCorpKey_Parseable(t *testing.T) {
	keyring, err := readPGPKey(hashicorpPGPKey)
	if err != nil {
		t.Fatalf("failed to parse embedded HashiCorp PGP key: %v", err)
	}
	if len(keyring) == 0 {
		t.Fatal("embedded HashiCorp PGP key contains no entities")
	}
}

// --- SHA256 Verification Tests ---

func TestVerifySHA256_Valid(t *testing.T) {
	content := []byte("terraform binary content")
	h := sha256.Sum256(content)
	hash := fmt.Sprintf("%x", h)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.zip")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	sums := fmt.Sprintf("%s  test.zip\n", hash)
	if err := VerifySHA256(path, []byte(sums), "test.zip"); err != nil {
		t.Fatalf("expected valid SHA256 to pass, got: %v", err)
	}
}

func TestVerifySHA256_Mismatch(t *testing.T) {
	content := []byte("actual content")
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.zip")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	sums := "0000000000000000000000000000000000000000000000000000000000000000  test.zip\n"
	err := VerifySHA256(path, []byte(sums), "test.zip")
	if err == nil {
		t.Fatal("expected SHA256 mismatch error")
	}
	if !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("expected mismatch error, got: %v", err)
	}
}

func TestVerifySHA256_MissingEntry(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.zip")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	sums := "abc123  other.zip\n"
	err := VerifySHA256(path, []byte(sums), "test.zip")
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
	if !strings.Contains(err.Error(), "no SHA256SUMS entry") {
		t.Fatalf("expected missing entry error, got: %v", err)
	}
}

func TestVerifySHA256_MultipleEntries(t *testing.T) {
	content := []byte("terraform binary")
	h := sha256.Sum256(content)
	hash := fmt.Sprintf("%x", h)

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "terraform_1.5.0_linux_amd64.zip")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	sums := fmt.Sprintf(
		"aaaa  terraform_1.5.0_darwin_amd64.zip\n%s  terraform_1.5.0_linux_amd64.zip\nbbbb  terraform_1.5.0_windows_amd64.zip\n",
		hash,
	)
	if err := VerifySHA256(path, []byte(sums), "terraform_1.5.0_linux_amd64.zip"); err != nil {
		t.Fatalf("expected SHA256 to pass with multiple entries, got: %v", err)
	}
}

// --- Zip Extraction Tests ---

func TestExtract_ValidZip(t *testing.T) {
	zipData := buildTestZip(t, "1.5.0")
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "terraform.zip")
	if err := os.WriteFile(zipPath, zipData, 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")
	if err := Extract(zipPath, destDir); err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	binPath := filepath.Join(destDir, "terraform")
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("terraform binary not found: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("terraform binary not executable: %v", info.Mode())
	}

	content, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Terraform v1.5.0") {
		t.Fatalf("unexpected content: %s", content)
	}
}

func TestExtract_ZipSlip_DotDot(t *testing.T) {
	// Create a zip with a "../evil" path entry.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("../evil")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("malicious"))
	zw.Close()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "evil.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")
	err = Extract(zipPath, destDir)
	if err == nil {
		t.Fatal("expected zip-slip to be rejected")
	}
	if !strings.Contains(err.Error(), "zip-slip") {
		t.Fatalf("expected zip-slip error, got: %v", err)
	}
}

func TestExtract_ZipSlip_AbsolutePath(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("/etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("malicious"))
	zw.Close()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "evil.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")
	err = Extract(zipPath, destDir)
	if err == nil {
		t.Fatal("expected absolute path zip entry to be rejected")
	}
	if !strings.Contains(err.Error(), "zip-slip") {
		t.Fatalf("expected zip-slip error, got: %v", err)
	}
}

func TestExtract_NoTerraformBinary(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create("readme.txt")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("not a binary"))
	zw.Close()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "nobin.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")
	err = Extract(zipPath, destDir)
	if err == nil {
		t.Fatal("expected error when no terraform binary in zip")
	}
	if !strings.Contains(err.Error(), "no terraform binary") {
		t.Fatalf("expected missing binary error, got: %v", err)
	}
}

func TestExtract_WindowsBinary(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	hdr := &zip.FileHeader{Name: "terraform.exe", Method: zip.Deflate}
	hdr.SetMode(0o755)
	fw, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("MZ windows binary"))
	zw.Close()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "terraform.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	destDir := filepath.Join(tmpDir, "dest")
	if err := Extract(zipPath, destDir); err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	binPath := filepath.Join(destDir, "terraform.exe")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("terraform.exe not found: %v", err)
	}
}

// --- Filename Construction Tests ---

func TestBuildZipFilename_Standard(t *testing.T) {
	got := buildZipFilename("1.5.0", "linux_amd64")
	want := "terraform_1.5.0_linux_amd64.zip"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildZipFilename_Alpha012(t *testing.T) {
	tests := []struct {
		version string
		want    string
	}{
		{"0.12.0-alpha3", "terraform_0.12.0-alpha3_terraform_0.12.0-alpha3_linux_amd64.zip"},
		{"0.12.0-alpha5", "terraform_0.12.0-alpha5_terraform_0.12.0-alpha5_linux_amd64.zip"},
		{"0.12.0-alpha9", "terraform_0.12.0-alpha9_terraform_0.12.0-alpha9_linux_amd64.zip"},
	}
	for _, tt := range tests {
		got := buildZipFilename(tt.version, "linux_amd64")
		if got != tt.want {
			t.Errorf("buildZipFilename(%q, linux_amd64) = %q, want %q", tt.version, got, tt.want)
		}
	}
}

func TestBuildZipFilename_NotAlpha012(t *testing.T) {
	tests := []string{
		"0.12.0-alpha1",
		"0.12.0-alpha2",
		"0.12.0-beta1",
		"0.12.0-rc1",
		"0.12.0",
		"1.0.0-alpha1",
	}
	for _, v := range tests {
		got := buildZipFilename(v, "linux_amd64")
		want := fmt.Sprintf("terraform_%s_linux_amd64.zip", v)
		if got != want {
			t.Errorf("buildZipFilename(%q, linux_amd64) = %q, want %q", v, got, want)
		}
	}
}

// --- Custom PGP Key Override Test ---

func TestLoadPGPKeyring_CustomKey(t *testing.T) {
	entity := testPGPEntity(t)
	keyData := armorPubKey(t, entity)

	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "custom.asc")
	if err := os.WriteFile(keyPath, keyData, 0o644); err != nil {
		t.Fatal(err)
	}

	keyring, err := loadPGPKeyring(keyPath)
	if err != nil {
		t.Fatalf("loading custom PGP key: %v", err)
	}
	if len(keyring) == 0 {
		t.Fatal("custom PGP keyring is empty")
	}
}

func TestLoadPGPKeyring_DefaultKey(t *testing.T) {
	keyring, err := loadPGPKeyring("")
	if err != nil {
		t.Fatalf("loading default PGP key: %v", err)
	}
	if len(keyring) == 0 {
		t.Fatal("default PGP keyring is empty")
	}
}

func TestLoadPGPKeyring_MissingFile(t *testing.T) {
	_, err := loadPGPKeyring("/nonexistent/path/key.asc")
	if err == nil {
		t.Fatal("expected error for missing PGP key file")
	}
}

// --- Netrc Tests ---

func TestLookupNetrc_Match(t *testing.T) {
	content := "machine example.com\n  login myuser\n  password mypass\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".netrc")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	login, pass, err := lookupNetrc(path, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if login != "myuser" || pass != "mypass" {
		t.Fatalf("got login=%q pass=%q, want myuser/mypass", login, pass)
	}
}

func TestLookupNetrc_NoMatch(t *testing.T) {
	content := "machine other.com\n  login user\n  password pass\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".netrc")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	login, _, err := lookupNetrc(path, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if login != "" {
		t.Fatalf("expected empty login for non-matching host, got %q", login)
	}
}

func TestLookupNetrc_Default(t *testing.T) {
	content := "default\n  login defuser\n  password defpass\n"
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".netrc")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	login, pass, err := lookupNetrc(path, "anything.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if login != "defuser" || pass != "defpass" {
		t.Fatalf("got login=%q pass=%q, want defuser/defpass", login, pass)
	}
}

// --- Full Download Pipeline Tests ---

func TestDownload_FullPipeline(t *testing.T) {
	entity := testPGPEntity(t)
	zipData := buildTestZip(t, "1.6.0")

	plat := platform.Platform{OS: "linux", Arch: "amd64"}
	server := startMockServer(t, []mockRelease{
		{version: "1.6.0", platStr: plat.String(), zipData: zipData, entity: entity},
	})

	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}

	result, err := Download(context.Background(), "1.6.0", plat, cfg)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if result.Version != "1.6.0" {
		t.Fatalf("got version %q, want 1.6.0", result.Version)
	}
	if result.Platform != "linux_amd64" {
		t.Fatalf("got platform %q, want linux_amd64", result.Platform)
	}
	if _, err := os.Stat(result.ZipPath); err != nil {
		t.Fatalf("zip file not found: %v", err)
	}

	// Clean up temp dir.
	os.RemoveAll(filepath.Dir(result.ZipPath))
}

func TestDownload_PGPVerificationFails(t *testing.T) {
	entity := testPGPEntity(t)
	zipData := buildTestZip(t, "1.6.0")

	plat := platform.Platform{OS: "linux", Arch: "amd64"}
	server := startMockServer(t, []mockRelease{
		{version: "1.6.0", platStr: plat.String(), zipData: zipData, entity: entity, tampered: true},
	})

	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}

	_, err := Download(context.Background(), "1.6.0", plat, cfg)
	if err == nil {
		t.Fatal("expected PGP verification failure")
	}
	if !strings.Contains(err.Error(), "PGP signature verification FAILED") {
		t.Fatalf("expected PGP failure message, got: %v", err)
	}
}

func TestDownload_HTTP404(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	entity := testPGPEntity(t)
	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}
	plat := platform.Platform{OS: "linux", Arch: "amd64"}

	_, err := Download(context.Background(), "99.99.99", plat, cfg)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 in error, got: %v", err)
	}
}

func TestDownload_HTTP500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	entity := testPGPEntity(t)
	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}
	plat := platform.Platform{OS: "linux", Arch: "amd64"}

	_, err := Download(context.Background(), "1.0.0", plat, cfg)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 in error, got: %v", err)
	}
}

func TestDownload_ContextCancellation(t *testing.T) {
	entity := testPGPEntity(t)
	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intentionally slow — the context should cancel before completion.
		select {}
	}))
	t.Cleanup(server.Close)

	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}
	plat := platform.Platform{OS: "linux", Arch: "amd64"}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := Download(ctx, "1.0.0", plat, cfg)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// --- Full Install Pipeline Test ---

func TestDownloadAndInstall_FullPipeline(t *testing.T) {
	entity := testPGPEntity(t)
	zipData := buildTestZip(t, "1.5.7")

	plat := platform.Platform{OS: "linux", Arch: "amd64"}
	server := startMockServer(t, []mockRelease{
		{version: "1.5.7", platStr: plat.String(), zipData: zipData, entity: entity},
	})

	configDir := t.TempDir()
	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: configDir,
		PGPKeyPath: pgpKeyPath,
	}

	result, err := Download(context.Background(), "1.5.7", plat, cfg)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	versionDir, err := Install(result, configDir)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	binPath := filepath.Join(versionDir, "terraform")
	info, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("terraform binary not found after install: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("terraform binary not executable: %v", info.Mode())
	}

	// Verify temp dir is cleaned up.
	if _, err := os.Stat(filepath.Dir(result.ZipPath)); !os.IsNotExist(err) {
		t.Fatal("temp directory was not cleaned up after install")
	}
}

// --- 0.12.0-alpha Edge Case Download Test ---

func TestDownload_Alpha012EdgeCase(t *testing.T) {
	entity := testPGPEntity(t)
	zipData := buildTestZip(t, "0.12.0-alpha5")

	plat := platform.Platform{OS: "linux", Arch: "amd64"}
	server := startMockServer(t, []mockRelease{
		{version: "0.12.0-alpha5", platStr: plat.String(), zipData: zipData, entity: entity},
	})

	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		PGPKeyPath: pgpKeyPath,
	}

	result, err := Download(context.Background(), "0.12.0-alpha5", plat, cfg)
	if err != nil {
		t.Fatalf("Download failed for alpha version: %v", err)
	}
	if result.Version != "0.12.0-alpha5" {
		t.Fatalf("got version %q, want 0.12.0-alpha5", result.Version)
	}

	// Clean up.
	os.RemoveAll(filepath.Dir(result.ZipPath))
}

// --- Download with Netrc Test ---

func TestDownload_WithNetrc(t *testing.T) {
	entity := testPGPEntity(t)
	zipData := buildTestZip(t, "1.6.0")
	plat := platform.Platform{OS: "linux", Arch: "amd64"}

	// Create a server that requires auth.
	var authChecked bool
	mux := http.NewServeMux()

	zipFilename := buildZipFilename("1.6.0", plat.String())
	sumsFilename := "terraform_1.6.0_SHA256SUMS"
	sigFilename := sumsFilename + sigKeyPostfix + ".sig"

	entries := map[string][]byte{zipFilename: zipData}
	sumsContent := buildTestSHA256SUMS(entries)
	sigData := detachSign(t, entity, []byte(sumsContent))

	authHandler := func(w http.ResponseWriter, r *http.Request, data []byte) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		authChecked = true
		w.Write(data)
	}

	prefix := "/terraform/1.6.0/"
	mux.HandleFunc(prefix+sumsFilename, func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, []byte(sumsContent))
	})
	mux.HandleFunc(prefix+sigFilename, func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, sigData)
	})
	mux.HandleFunc(prefix+zipFilename, func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, zipData)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Write netrc file.
	tmpDir := t.TempDir()
	netrcPath := filepath.Join(tmpDir, ".netrc")
	// The httptest server hostname is 127.0.0.1.
	netrcContent := "machine 127.0.0.1\n  login testuser\n  password testpass\n"
	if err := os.WriteFile(netrcPath, []byte(netrcContent), 0o600); err != nil {
		t.Fatal(err)
	}

	pgpKeyPath := writeTempKey(t, armorPubKey(t, entity))
	cfg := &config.Config{
		Remote:    server.URL,
		ConfigDir: t.TempDir(),
		NetrcPath: netrcPath,
		PGPKeyPath: pgpKeyPath,
	}

	result, err := Download(context.Background(), "1.6.0", plat, cfg)
	if err != nil {
		t.Fatalf("Download with netrc failed: %v", err)
	}
	if !authChecked {
		t.Fatal("netrc authentication was not used")
	}

	os.RemoveAll(filepath.Dir(result.ZipPath))
}

// --- SHA256SUMS Parsing Edge Cases ---

func TestFindSHA256Entry_TwoSpaceSeparator(t *testing.T) {
	sums := "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd  file.zip\n"
	hash, err := findSHA256Entry([]byte(sums), "file.zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd" {
		t.Fatalf("unexpected hash: %s", hash)
	}
}

func TestFindSHA256Entry_EmptyLines(t *testing.T) {
	sums := "\n\nabc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd  file.zip\n\n"
	hash, err := findSHA256Entry([]byte(sums), "file.zip")
	if err != nil {
		t.Fatal(err)
	}
	if hash != "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd" {
		t.Fatalf("unexpected hash: %s", hash)
	}
}

// --- Helper ---

func writeTempKey(t *testing.T, keyData []byte) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.asc")
	if err := os.WriteFile(path, keyData, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
