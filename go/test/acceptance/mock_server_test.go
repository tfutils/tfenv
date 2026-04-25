package acceptance

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sort"
	"strings"
	"testing"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// MockRelease describes a single Terraform version served by the mock server.
type MockRelease struct {
	Version   string
	Platforms []MockPlatform
}

// MockPlatform describes an os/arch combination for a release.
type MockPlatform struct {
	OS   string
	Arch string
}

// MockServer wraps an httptest.Server and exposes the test PGP key for
// verification testing.
type MockServer struct {
	Server       *httptest.Server
	URL          string
	PGPEntity    *openpgp.Entity
	ArmoredPubKey string
}

// Close shuts down the mock server.
func (m *MockServer) Close() {
	m.Server.Close()
}

// defaultMockReleases returns a set of fake versions covering common test
// scenarios: multiple minor versions, patch levels, and a pre-release.
func defaultMockReleases() []MockRelease {
	os := runtime.GOOS
	arch := runtime.GOARCH
	return []MockRelease{
		{Version: "1.6.1", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
		{Version: "1.6.0", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
		{Version: "1.5.7", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
		{Version: "1.5.7-rc1", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
		{Version: "1.5.6", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
		{Version: "1.4.0", Platforms: []MockPlatform{{OS: os, Arch: arch}}},
	}
}

// startMockReleaseServer creates and starts an httptest.Server serving mock
// HashiCorp-style release artifacts. The server is automatically closed via
// t.Cleanup.
func startMockReleaseServer(t *testing.T) *MockServer {
	t.Helper()

	releases := defaultMockReleases()
	entity := generateTestPGPKey(t)

	armoredPub := armorPublicKey(t, entity)

	mux := http.NewServeMux()

	// Version index at /terraform/ — HTML with links like the real
	// releases.hashicorp.com page.
	mux.HandleFunc("/terraform/", func(w http.ResponseWriter, r *http.Request) {
		// If the path has more segments, it's a per-version directory request.
		trimmed := strings.TrimPrefix(r.URL.Path, "/terraform/")
		trimmed = strings.TrimSuffix(trimmed, "/")

		if trimmed == "" {
			// Root index page.
			serveVersionIndex(w, releases)
			return
		}

		// Try to match version/filename.
		parts := strings.SplitN(trimmed, "/", 2)
		version := parts[0]
		var release *MockRelease
		for i := range releases {
			if releases[i].Version == version {
				release = &releases[i]
				break
			}
		}
		if release == nil {
			http.NotFound(w, r)
			return
		}

		if len(parts) == 1 {
			// Version directory listing — not strictly needed but helpful.
			serveVersionDirectory(w, *release)
			return
		}

		filename := parts[1]
		serveReleaseArtifact(t, w, r, *release, filename, entity)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return &MockServer{
		Server:        server,
		URL:           server.URL,
		PGPEntity:     entity,
		ArmoredPubKey: armoredPub,
	}
}

// serveVersionIndex writes an HTML page with links to each version, matching
// the format found on releases.hashicorp.com/terraform/.
func serveVersionIndex(w http.ResponseWriter, releases []MockRelease) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	buf.WriteString("<html><body>\n")
	for _, r := range releases {
		fmt.Fprintf(&buf, "<a href=\"/terraform/%s/\">terraform_%s</a>\n", r.Version, r.Version)
	}
	buf.WriteString("</body></html>\n")
	w.Write(buf.Bytes())
}

// serveVersionDirectory writes an HTML page listing artifacts for a version.
func serveVersionDirectory(w http.ResponseWriter, release MockRelease) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	buf.WriteString("<html><body>\n")
	prefix := "terraform_" + release.Version
	for _, p := range release.Platforms {
		name := fmt.Sprintf("%s_%s_%s.zip", prefix, p.OS, p.Arch)
		fmt.Fprintf(&buf, "<a href=\"%s\">%s</a>\n", name, name)
	}
	fmt.Fprintf(&buf, "<a href=\"%s_SHA256SUMS\">%s_SHA256SUMS</a>\n", prefix, prefix)
	fmt.Fprintf(&buf, "<a href=\"%s_SHA256SUMS.72D7468F.sig\">%s_SHA256SUMS.72D7468F.sig</a>\n", prefix, prefix)
	buf.WriteString("</body></html>\n")
	w.Write(buf.Bytes())
}

// serveReleaseArtifact serves an individual file from a version directory.
func serveReleaseArtifact(t *testing.T, w http.ResponseWriter, r *http.Request, release MockRelease, filename string, entity *openpgp.Entity) {
	t.Helper()
	prefix := "terraform_" + release.Version

	switch {
	case filename == prefix+"_SHA256SUMS":
		serveSHA256SUMS(t, w, release)
	case filename == prefix+"_SHA256SUMS.72D7468F.sig":
		serveSHA256SUMSSig(t, w, release, entity)
	case strings.HasSuffix(filename, ".zip"):
		serveZip(t, w, release, filename)
	default:
		http.NotFound(w, r)
	}
}

// serveSHA256SUMS generates and serves the SHA256SUMS file for a release.
func serveSHA256SUMS(t *testing.T, w http.ResponseWriter, release MockRelease) {
	t.Helper()
	content := buildSHA256SUMS(t, release)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

// serveSHA256SUMSSig generates and serves a PGP detached signature of the
// SHA256SUMS file.
func serveSHA256SUMSSig(t *testing.T, w http.ResponseWriter, release MockRelease, entity *openpgp.Entity) {
	t.Helper()
	content := buildSHA256SUMS(t, release)

	var sigBuf bytes.Buffer
	err := openpgp.DetachSign(&sigBuf, entity, strings.NewReader(content), &packet.Config{
		DefaultHash: crypto.SHA256,
	})
	if err != nil {
		t.Fatalf("failed to create detached signature: %v", err)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(sigBuf.Bytes())
}

// serveZip serves a zip archive containing a stub terraform binary.
func serveZip(t *testing.T, w http.ResponseWriter, release MockRelease, filename string) {
	t.Helper()

	// Parse platform from filename: terraform_VERSION_OS_ARCH.zip
	prefix := "terraform_" + release.Version + "_"
	suffix := ".zip"
	if !strings.HasPrefix(filename, prefix) || !strings.HasSuffix(filename, suffix) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	zipContent := buildZip(t, release.Version)
	w.Header().Set("Content-Type", "application/zip")
	w.Write(zipContent)
}

// buildZip creates a zip archive containing a stub terraform script that
// prints "Terraform vVERSION" when executed.
func buildZip(t *testing.T, version string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Create a stub shell script as the "terraform" binary.
	stub := fmt.Sprintf("#!/bin/sh\necho \"Terraform v%s\"\n", version)

	header := &zip.FileHeader{
		Name:   "terraform",
		Method: zip.Deflate,
	}
	// Set executable permissions.
	header.SetMode(0o755)

	fw, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := fw.Write([]byte(stub)); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

// buildSHA256SUMS generates the SHA256SUMS content for a release. Each line
// has the format "hash  filename".
func buildSHA256SUMS(t *testing.T, release MockRelease) string {
	t.Helper()

	// Collect entries sorted by filename for deterministic output.
	type entry struct {
		hash     string
		filename string
	}
	var entries []entry

	for _, p := range release.Platforms {
		zipData := buildZip(t, release.Version)
		h := sha256.Sum256(zipData)
		filename := fmt.Sprintf("terraform_%s_%s_%s.zip", release.Version, p.OS, p.Arch)
		entries = append(entries, entry{
			hash:     fmt.Sprintf("%x", h),
			filename: filename,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].filename < entries[j].filename
	})

	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s  %s", e.hash, e.filename))
	}

	return strings.Join(lines, "\n") + "\n"
}

// generateTestPGPKey creates a new PGP entity for test signing. The key is
// ephemeral and used only within the test run.
func generateTestPGPKey(t *testing.T) *openpgp.Entity {
	t.Helper()

	entity, err := openpgp.NewEntity("tfenv-test", "test signing key", "test@tfenv.test", &packet.Config{
		DefaultHash: crypto.SHA256,
	})
	if err != nil {
		t.Fatalf("failed to generate test PGP key: %v", err)
	}

	return entity
}

// armorPublicKey exports the public key of an entity in ASCII-armored format.
func armorPublicKey(t *testing.T, entity *openpgp.Entity) string {
	t.Helper()

	var buf bytes.Buffer
	w, err := armor.Encode(&buf, openpgp.PublicKeyType, nil)
	if err != nil {
		t.Fatalf("failed to create armor encoder: %v", err)
	}
	if err := entity.Serialize(w); err != nil {
		t.Fatalf("failed to serialize public key: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close armor encoder: %v", err)
	}

	return buf.String()
}
