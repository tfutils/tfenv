package acceptance

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/openpgp"
)

// TestTfenvVersion validates that the binary under test prints a version
// string and exits 0.
func TestTfenvVersion(t *testing.T) {
	t.Parallel()

	stdout, _, exitCode := runTfenv(t, "version")
	assertExitCode(t, exitCode, 0)
	assertOutputContains(t, stdout, "tfenv")
}

// TestTfenvHelp validates that the binary prints help text and exits 0.
func TestTfenvHelp(t *testing.T) {
	t.Parallel()

	stdout, _, exitCode := runTfenv(t, "help")
	assertExitCode(t, exitCode, 0)
	assertOutputContains(t, stdout, "Usage:")
}

// TestTfenvUnknownCommand validates that the binary exits non-zero and
// prints an error when given an unknown subcommand.
func TestTfenvUnknownCommand(t *testing.T) {
	t.Parallel()

	_, stderr, exitCode := runTfenv(t, "not-a-command")

	if exitCode == 0 {
		t.Errorf("expected non-zero exit code for unknown command, got 0")
	}
	assertOutputContains(t, stderr, "unknown command")
}

// TestMockServerServesVersionIndex validates that the mock release server
// returns an HTML page with version links that are parseable by the same
// regex the Bash edition uses.
func TestMockServerServesVersionIndex(t *testing.T) {
	t.Parallel()

	mock := startMockReleaseServer(t)

	resp, err := http.Get(mock.URL + "/terraform/")
	if err != nil {
		t.Fatalf("failed to GET version index: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	html := string(body)

	// The Bash edition uses this regex to extract versions.
	re := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+(-(rc|beta|alpha|oci)-?[0-9]*)?`)
	matches := re.FindAllString(html, -1)

	if len(matches) == 0 {
		t.Fatalf("expected version matches in index page, got none.\nBody:\n%s", html)
	}

	// Check that our known versions appear.
	found := make(map[string]bool)
	for _, m := range matches {
		found[m] = true
	}

	for _, want := range []string{"1.6.1", "1.6.0", "1.5.7", "1.5.7-rc1"} {
		if !found[want] {
			t.Errorf("expected version %q in index, found: %v", want, found)
		}
	}
}

// TestMockServerServesZip validates that a zip archive can be downloaded
// for a known version and platform.
func TestMockServerServesZip(t *testing.T) {
	t.Parallel()

	mock := startMockReleaseServer(t)

	releases := defaultMockReleases()
	r := releases[0] // 1.6.1
	p := r.Platforms[0]

	url := mock.URL + "/terraform/" + r.Version + "/terraform_" + r.Version + "_" + p.OS + "_" + p.Arch + ".zip"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET zip: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read zip body: %v", err)
	}

	if len(body) == 0 {
		t.Fatal("zip body is empty")
	}

	// Quick validation: zip files start with PK signature (0x504b).
	if body[0] != 0x50 || body[1] != 0x4b {
		t.Fatalf("downloaded content does not look like a zip (first bytes: %x %x)", body[0], body[1])
	}
}

// TestMockServerServesSHA256SUMS validates that the SHA256SUMS file is
// served correctly and contains expected hash/filename lines.
func TestMockServerServesSHA256SUMS(t *testing.T) {
	t.Parallel()

	mock := startMockReleaseServer(t)

	url := mock.URL + "/terraform/1.6.1/terraform_1.6.1_SHA256SUMS"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET SHA256SUMS: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read SHA256SUMS body: %v", err)
	}

	content := string(body)
	if !strings.Contains(content, "terraform_1.6.1_") {
		t.Errorf("SHA256SUMS does not reference expected filename.\nContent:\n%s", content)
	}

	// Each line should be: <hex hash>  <filename>
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			t.Errorf("malformed SHA256SUMS line: %q", line)
			continue
		}
		if len(parts[0]) != 64 {
			t.Errorf("hash length is %d, want 64: %q", len(parts[0]), parts[0])
		}
	}
}

// TestMockServerServesValidSignature validates that the PGP signature
// served by the mock server can be verified against its public key.
func TestMockServerServesValidSignature(t *testing.T) {
	t.Parallel()

	mock := startMockReleaseServer(t)

	// Fetch SHA256SUMS.
	sumsResp, err := http.Get(mock.URL + "/terraform/1.6.1/terraform_1.6.1_SHA256SUMS")
	if err != nil {
		t.Fatalf("failed to GET SHA256SUMS: %v", err)
	}
	defer sumsResp.Body.Close()
	sumsBody, err := io.ReadAll(sumsResp.Body)
	if err != nil {
		t.Fatalf("failed to read SHA256SUMS: %v", err)
	}

	// Fetch signature.
	sigResp, err := http.Get(mock.URL + "/terraform/1.6.1/terraform_1.6.1_SHA256SUMS.72D7468F.sig")
	if err != nil {
		t.Fatalf("failed to GET signature: %v", err)
	}
	defer sigResp.Body.Close()
	sigBody, err := io.ReadAll(sigResp.Body)
	if err != nil {
		t.Fatalf("failed to read signature: %v", err)
	}

	// Build a keyring from the mock server's entity.
	keyring := openpgp.EntityList{mock.PGPEntity}

	// Verify the detached signature.
	signer, err := openpgp.CheckDetachedSignature(
		keyring,
		strings.NewReader(string(sumsBody)),
		strings.NewReader(string(sigBody)),
	)
	if err != nil {
		t.Fatalf("signature verification failed: %v", err)
	}

	if signer == nil {
		t.Fatal("signer is nil — verification succeeded but no signer returned")
	}
}
