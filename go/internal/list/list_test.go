package list

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tfutils/tfenv/go/internal/config"
)

// realisticHTML mirrors the format served by releases.hashicorp.com/terraform/.
const realisticHTML = `<!DOCTYPE html>
<html>
<head><title>Terraform Versions</title></head>
<body>
<ul>
  <li><a href="/terraform/1.10.0-alpha20240710/">terraform_1.10.0-alpha20240710</a></li>
  <li><a href="/terraform/1.9.2/">terraform_1.9.2</a></li>
  <li><a href="/terraform/1.9.1/">terraform_1.9.1</a></li>
  <li><a href="/terraform/1.9.0/">terraform_1.9.0</a></li>
  <li><a href="/terraform/1.9.0-rc2/">terraform_1.9.0-rc2</a></li>
  <li><a href="/terraform/1.9.0-rc1/">terraform_1.9.0-rc1</a></li>
  <li><a href="/terraform/1.9.0-beta1/">terraform_1.9.0-beta1</a></li>
  <li><a href="/terraform/1.8.5/">terraform_1.8.5</a></li>
  <li><a href="/terraform/1.8.4/">terraform_1.8.4</a></li>
  <li><a href="/terraform/1.7.0/">terraform_1.7.0</a></li>
  <li><a href="/terraform/0.14.11/">terraform_0.14.11</a></li>
  <li><a href="/terraform/0.13.7/">terraform_0.13.7</a></li>
</ul>
</body>
</html>`

func TestParseVersionIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		html    string
		want    []string
		wantErr bool
	}{
		{
			name: "realistic hashicorp HTML",
			html: realisticHTML,
			want: []string{
				"1.10.0-alpha20240710",
				"1.9.2", "1.9.1", "1.9.0",
				"1.9.0-rc2", "1.9.0-rc1", "1.9.0-beta1",
				"1.8.5", "1.8.4",
				"1.7.0",
				"0.14.11", "0.13.7",
			},
		},
		{
			name: "empty page",
			html: `<html><body></body></html>`,
			want: nil,
		},
		{
			name: "no terraform links",
			html: `<html><body><a href="/packer/1.0.0/">packer_1.0.0</a></body></html>`,
			want: nil,
		},
		{
			name: "links without trailing slash",
			html: `<html><body><a href="/terraform/1.5.0">terraform_1.5.0</a></body></html>`,
			want: []string{"1.5.0"},
		},
		{
			name: "mixed valid and invalid hrefs",
			html: `<html><body>
				<a href="/terraform/1.5.0/">terraform_1.5.0</a>
				<a href="/terraform/notaversion/">terraform_notaversion</a>
				<a href="/terraform/1.4.0/">terraform_1.4.0</a>
			</body></html>`,
			want: []string{"1.5.0", "1.4.0"},
		},
		{
			name: "pre-release versions included",
			html: `<html><body>
				<a href="/terraform/1.6.0-alpha1/">terraform_1.6.0-alpha1</a>
				<a href="/terraform/1.5.0-rc1/">terraform_1.5.0-rc1</a>
				<a href="/terraform/1.5.0-beta2/">terraform_1.5.0-beta2</a>
				<a href="/terraform/1.4.0/">terraform_1.4.0</a>
			</body></html>`,
			want: []string{"1.6.0-alpha1", "1.5.0-rc1", "1.5.0-beta2", "1.4.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseVersionIndex(strings.NewReader(tc.html))
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseVersionIndex() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if !slicesEqual(got, tc.want) {
				t.Errorf("parseVersionIndex() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSortVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []string
		reverse bool
		want    []string
	}{
		{
			name:    "descending (default)",
			input:   []string{"1.0.0", "1.2.0", "0.14.11", "1.1.0"},
			reverse: false,
			want:    []string{"1.2.0", "1.1.0", "1.0.0", "0.14.11"},
		},
		{
			name:    "ascending (reverse)",
			input:   []string{"1.0.0", "1.2.0", "0.14.11", "1.1.0"},
			reverse: true,
			want:    []string{"0.14.11", "1.0.0", "1.1.0", "1.2.0"},
		},
		{
			name:    "pre-releases sort before stable",
			input:   []string{"1.5.0", "1.5.0-rc1", "1.5.0-beta1", "1.4.0"},
			reverse: false,
			want:    []string{"1.5.0", "1.5.0-rc1", "1.5.0-beta1", "1.4.0"},
		},
		{
			name:    "pre-releases ascending",
			input:   []string{"1.5.0", "1.5.0-rc1", "1.5.0-beta1", "1.4.0"},
			reverse: true,
			want:    []string{"1.4.0", "1.5.0-beta1", "1.5.0-rc1", "1.5.0"},
		},
		{
			name:    "empty slice",
			input:   []string{},
			reverse: false,
			want:    []string{},
		},
		{
			name:    "single version",
			input:   []string{"1.0.0"},
			reverse: false,
			want:    []string{"1.0.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := make([]string, len(tc.input))
			copy(got, tc.input)
			sortVersions(got, tc.reverse)
			if !slicesEqual(got, tc.want) {
				t.Errorf("sortVersions() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestListRemote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		reverse    bool
		wantCount  int
		wantFirst  string
		wantLast   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successful fetch sorted descending",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != userAgent {
					t.Errorf("expected User-Agent %q, got %q", userAgent, r.Header.Get("User-Agent"))
				}
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(realisticHTML))
			},
			reverse:   false,
			wantCount: 12,
			wantFirst: "1.10.0-alpha20240710",
			wantLast:  "0.13.7",
		},
		{
			name: "successful fetch sorted ascending",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(realisticHTML))
			},
			reverse:   true,
			wantCount: 12,
			wantFirst: "0.13.7",
			wantLast:  "1.10.0-alpha20240710",
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`<html><body></body></html>`))
			},
			wantCount: 0,
		},
		{
			name: "HTTP 404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 404",
		},
		{
			name: "HTTP 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:    true,
			wantErrMsg: "HTTP 500",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(tc.handler)
			t.Cleanup(server.Close)

			cfg := &config.Config{
				Remote:        server.URL,
				ReverseRemote: tc.reverse,
			}

			got, err := ListRemote(cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrMsg != "" && !strings.Contains(err.Error(), tc.wantErrMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != tc.wantCount {
				t.Errorf("got %d versions, want %d", len(got), tc.wantCount)
			}
			if tc.wantFirst != "" && len(got) > 0 && got[0] != tc.wantFirst {
				t.Errorf("first version = %q, want %q", got[0], tc.wantFirst)
			}
			if tc.wantLast != "" && len(got) > 0 && got[len(got)-1] != tc.wantLast {
				t.Errorf("last version = %q, want %q", got[len(got)-1], tc.wantLast)
			}
		})
	}
}

func TestListLocal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns configDir
		reverse bool
		want    []string
		wantErr bool
	}{
		{
			name: "multiple versions sorted descending",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				versionsDir := filepath.Join(dir, "versions")
				for _, v := range []string{"1.0.0", "1.2.0", "0.14.11", "1.1.0"} {
					if err := os.MkdirAll(filepath.Join(versionsDir, v), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			want: []string{"1.2.0", "1.1.0", "1.0.0", "0.14.11"},
		},
		{
			name: "multiple versions sorted ascending",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				versionsDir := filepath.Join(dir, "versions")
				for _, v := range []string{"1.0.0", "1.2.0", "0.14.11"} {
					if err := os.MkdirAll(filepath.Join(versionsDir, v), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			reverse: true,
			want:    []string{"0.14.11", "1.0.0", "1.2.0"},
		},
		{
			name: "empty versions directory",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				if err := os.MkdirAll(filepath.Join(dir, "versions"), 0o755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: nil,
		},
		{
			name: "nonexistent versions directory",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			want: nil,
		},
		{
			name: "ignores non-directory entries",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				versionsDir := filepath.Join(dir, "versions")
				if err := os.MkdirAll(filepath.Join(versionsDir, "1.5.0"), 0o755); err != nil {
					t.Fatal(err)
				}
				// Create a regular file that looks like a version.
				if err := os.WriteFile(filepath.Join(versionsDir, "1.4.0"), []byte("not a dir"), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			want: []string{"1.5.0"},
		},
		{
			name: "ignores non-version directory names",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				versionsDir := filepath.Join(dir, "versions")
				for _, v := range []string{"1.5.0", "notaversion", ".hidden"} {
					if err := os.MkdirAll(filepath.Join(versionsDir, v), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			want: []string{"1.5.0"},
		},
		{
			name: "pre-release versions included",
			setup: func(t *testing.T) string {
				t.Helper()
				dir := t.TempDir()
				versionsDir := filepath.Join(dir, "versions")
				for _, v := range []string{"1.5.0", "1.5.0-rc1", "1.4.0"} {
					if err := os.MkdirAll(filepath.Join(versionsDir, v), 0o755); err != nil {
						t.Fatal(err)
					}
				}
				return dir
			},
			want: []string{"1.5.0", "1.5.0-rc1", "1.4.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			configDir := tc.setup(t)
			cfg := &config.Config{
				ConfigDir:     configDir,
				ReverseRemote: tc.reverse,
			}

			got, err := ListLocal(cfg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slicesEqual(got, tc.want) {
				t.Errorf("ListLocal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVersionFromHref(t *testing.T) {
	t.Parallel()

	tests := []struct {
		href string
		want string
	}{
		{"/terraform/1.5.0/", "1.5.0"},
		{"/terraform/1.5.0-rc1/", "1.5.0-rc1"},
		{"/terraform/0.14.11/", "0.14.11"},
		{"terraform/1.5.0/", "1.5.0"},
		{"/terraform/1.5.0", "1.5.0"},
		{"/packer/1.5.0/", ""},
		{"/terraform/", ""},
		{"/terraform/notaversion/", ""},
		{"", ""},
		{"/", ""},
		{"/terraform/1.5.0/extra/", ""},
	}

	for _, tc := range tests {
		t.Run(tc.href, func(t *testing.T) {
			t.Parallel()
			got := versionFromHref(tc.href)
			if got != tc.want {
				t.Errorf("versionFromHref(%q) = %q, want %q", tc.href, got, tc.want)
			}
		})
	}
}

// slicesEqual compares two string slices for equality. Treats nil and empty
// slices as equal.
func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
