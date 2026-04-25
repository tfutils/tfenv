package install

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/platform"
)

const (
	// downloadTimeout is the HTTP client timeout for large downloads.
	downloadTimeout = 5 * time.Minute

	// userAgent identifies tfenv/go in HTTP requests.
	userAgent = "tfenv/go"

	// sigKeyPostfix is the PGP key fingerprint suffix used in signature filenames.
	sigKeyPostfix = ".72D7468F"
)

// alpha012Re matches version strings for 0.12.0-alpha3 through 0.12.0-alpha9.
var alpha012Re = regexp.MustCompile(`^0\.12\.0-alpha[3-9]$`)

// DownloadResult contains the paths and metadata from a successful download.
type DownloadResult struct {
	ZipPath  string // Path to downloaded zip file
	Version  string // The version that was downloaded
	Platform string // "linux_amd64" etc.
}

// Download fetches the Terraform zip for the given version and platform,
// verifies its PGP signature and SHA256 checksum, and returns the paths.
// All files are downloaded to a temporary directory. The caller is responsible
// for calling Install() to move verified artifacts to the final location.
func Download(ctx context.Context, version string, plat platform.Platform, cfg *config.Config) (*DownloadResult, error) {
	arch := plat.DownloadArch(version)
	platStr := fmt.Sprintf("%s_%s", plat.OS, arch)

	zipFilename := buildZipFilename(version, platStr)
	sumsFilename := fmt.Sprintf("terraform_%s_SHA256SUMS", version)
	sigFilename := fmt.Sprintf("%s%s.sig", sumsFilename, sigKeyPostfix)

	baseURL := strings.TrimRight(cfg.Remote, "/")
	versionURL := fmt.Sprintf("%s/terraform/%s", baseURL, version)

	client := buildHTTPClient(cfg)

	// Create temporary directory for all downloads.
	tmpDir, err := os.MkdirTemp("", "tfenv-download-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}

	logging.Debug("Downloading artifacts", "version", version, "platform", platStr, "tmpdir", tmpDir)

	// Step 1: Download SHA256SUMS.
	sumsPath := filepath.Join(tmpDir, sumsFilename)
	sumsURL := fmt.Sprintf("%s/%s", versionURL, sumsFilename)
	logging.Info("Downloading SHA256SUMS", "url", sumsURL)
	if err := downloadFile(ctx, client, sumsURL, sumsPath); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("downloading SHA256SUMS for Terraform %s: %w", version, err)
	}

	// Step 2: Download SHA256SUMS signature.
	sigPath := filepath.Join(tmpDir, sigFilename)
	sigURL := fmt.Sprintf("%s/%s", versionURL, sigFilename)
	logging.Info("Downloading SHA256SUMS signature", "url", sigURL)
	if err := downloadFile(ctx, client, sigURL, sigPath); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("downloading SHA256SUMS signature for Terraform %s: %w", version, err)
	}

	// Step 3: Verify PGP signature of SHA256SUMS.
	logging.Info("Verifying PGP signature")
	sumsData, err := os.ReadFile(sumsPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("reading SHA256SUMS: %w", err)
	}
	sigData, err := os.ReadFile(sigPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("reading SHA256SUMS signature: %w", err)
	}

	keyring, err := loadPGPKeyring(cfg.PGPKeyPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("loading PGP keyring: %w", err)
	}
	if err := verifyDetachedSig(keyring, sumsData, sigData); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf(
			"PGP signature verification FAILED for Terraform %s — the download may be tampered with: %w",
			version, err,
		)
	}
	logging.Info("PGP signature verified successfully")

	// Step 4: Download zip archive.
	zipPath := filepath.Join(tmpDir, zipFilename)
	zipURL := fmt.Sprintf("%s/%s", versionURL, zipFilename)
	logging.Info("Downloading Terraform", "version", version, "platform", platStr, "url", zipURL)
	if err := downloadFile(ctx, client, zipURL, zipPath); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("downloading Terraform %s zip: %w", version, err)
	}

	// Step 5: Verify SHA256 checksum of zip.
	logging.Info("Verifying SHA256 checksum")
	if err := VerifySHA256(zipPath, sumsData, zipFilename); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("SHA256 verification failed for Terraform %s: %w", version, err)
	}
	logging.Info("SHA256 checksum verified successfully")

	return &DownloadResult{
		ZipPath:  zipPath,
		Version:  version,
		Platform: platStr,
	}, nil
}

// Install extracts a verified download into the final versions directory and
// cleans up the temporary download directory. This should only be called after
// Download() returns successfully (all verification passed).
func Install(result *DownloadResult, configDir string) (string, error) {
	versionDir := filepath.Join(configDir, "versions", result.Version)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return "", fmt.Errorf("creating version directory: %w", err)
	}

	logging.Info("Extracting Terraform", "version", result.Version, "dest", versionDir)
	if err := Extract(result.ZipPath, versionDir); err != nil {
		return "", fmt.Errorf("extracting Terraform %s: %w", result.Version, err)
	}

	// Clean up temp directory.
	tmpDir := filepath.Dir(result.ZipPath)
	os.RemoveAll(tmpDir)

	logging.Info("Terraform installed", "version", result.Version, "path", versionDir)
	return versionDir, nil
}

// buildZipFilename returns the expected zip filename for a version and platform.
// Handles the 0.12.0-alpha edge case with double-prefixed filenames.
func buildZipFilename(version, platStr string) string {
	if alpha012Re.MatchString(version) {
		return fmt.Sprintf("terraform_%s_terraform_%s_%s.zip", version, version, platStr)
	}
	return fmt.Sprintf("terraform_%s_%s.zip", version, platStr)
}

// buildHTTPClient creates an HTTP client with appropriate timeouts and netrc
// support if configured.
func buildHTTPClient(cfg *config.Config) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext

	client := &http.Client{
		Transport: transport,
		Timeout:   downloadTimeout,
	}

	if cfg.NetrcPath != "" {
		client.Transport = &netrcTransport{
			base:      transport,
			netrcPath: cfg.NetrcPath,
		}
	}
	return client
}

// downloadFile fetches a URL and writes it to the given path.
func downloadFile(ctx context.Context, client *http.Client, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", destPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing response to %q: %w", destPath, err)
	}
	return nil
}

// verifyDetachedSig checks a PGP detached signature against the given keyring.
func verifyDetachedSig(keyring openpgp.EntityList, data, sig []byte) error {
	_, err := openpgp.CheckDetachedSignature(keyring, bytes.NewReader(data), bytes.NewReader(sig))
	if err != nil {
		return fmt.Errorf("PGP signature verification failed: %w", err)
	}
	return nil
}
