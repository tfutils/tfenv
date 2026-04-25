package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tfutils/tfenv/go/internal/config"
)

// sampleVersions is a realistic set of available versions for tests.
var sampleVersions = []string{
	"1.0.0",
	"1.0.1",
	"1.1.0",
	"1.2.0",
	"1.3.0",
	"1.4.0",
	"1.4.1",
	"1.5.0",
	"1.5.1",
	"1.5.2",
	"1.5.3",
	"1.5.4",
	"1.5.5",
	"1.5.6",
	"1.5.7",
	"1.6.0",
	"1.6.1",
	"1.6.2",
	"1.6.3",
	"1.6.4",
	"1.6.5",
	"1.6.6",
	"1.7.0",
	"1.7.1",
	"1.7.2",
	"1.7.3",
	"1.7.4",
	"1.7.5",
	"1.8.0",
	"1.8.1",
	"1.8.2",
	"1.8.3",
	"1.8.4",
	"1.8.5",
	"1.9.0",
	"1.9.1",
	"1.9.2",
	"1.9.3",
	"1.9.4",
	"1.9.5",
	"1.9.6",
	"1.9.7",
	"1.9.8",
	"1.10.0",
	"1.10.1",
	"1.10.2",
	"1.10.3",
	"2.0.0-alpha1",
	"2.0.0-beta1",
	"2.0.0-rc1",
	"2.0.0",
}

// ---------------------------------------------------------------------------
// Exact version resolution
// ---------------------------------------------------------------------------

func TestResolveVersion_Exact(t *testing.T) {
	tests := []struct {
		name      string
		specifier string
		want      string
	}{
		{"simple version", "1.5.0", "1.5.0"},
		{"with patch", "1.5.7", "1.5.7"},
		{"v prefix stripped", "v1.5.0", "1.5.0"},
		{"pre-release exact", "2.0.0-alpha1", "2.0.0-alpha1"},
		{"two-dot version", "1.10.3", "1.10.3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ResolveVersion(
				tc.specifier, sampleVersions, "test", &config.Config{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Version != tc.want {
				t.Errorf("got %q, want %q", result.Version, tc.want)
			}
			if result.Source != "test" {
				t.Errorf("source = %q, want %q", result.Source, "test")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// latest — newest stable
// ---------------------------------------------------------------------------

func TestResolveVersion_Latest(t *testing.T) {
	result, err := ResolveVersion(
		"latest", sampleVersions, "test", &config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "2.0.0" {
		t.Errorf("got %q, want %q", result.Version, "2.0.0")
	}
}

func TestResolveVersion_LatestExcludesPrerelease(t *testing.T) {
	// Only pre-release versions + one stable.
	versions := []string{
		"1.0.0",
		"2.0.0-alpha1",
		"2.0.0-beta1",
		"2.0.0-rc1",
	}
	result, err := ResolveVersion(
		"latest", versions, "test", &config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.0.0" {
		t.Errorf("got %q, want %q — latest should exclude pre-releases",
			result.Version, "1.0.0")
	}
}

func TestResolveVersion_LatestAllPrerelease(t *testing.T) {
	versions := []string{
		"2.0.0-alpha1",
		"2.0.0-beta1",
	}
	_, err := ResolveVersion(
		"latest", versions, "test", &config.Config{})
	if err == nil {
		t.Fatal("expected error when all versions are pre-release")
	}
}

func TestResolveVersion_LatestEmptyVersions(t *testing.T) {
	_, err := ResolveVersion(
		"latest", nil, "test", &config.Config{})
	if err == nil {
		t.Fatal("expected error with empty versions")
	}
}

// ---------------------------------------------------------------------------
// latest:<regex> — regex matching
// ---------------------------------------------------------------------------

func TestResolveVersion_LatestRegex(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    string
		wantErr bool
	}{
		{
			name: "match 1.5.x",
			spec: `latest:^1\.5\.`,
			want: "1.5.7",
		},
		{
			name: "match 1.5.x does not match 1.50.x",
			spec: `latest:^1\.5\.`,
			want: "1.5.7",
		},
		{
			name: "match 1.6.x",
			spec: `latest:^1\.6\.`,
			want: "1.6.6",
		},
		{
			name: "match 2.x pre-releases included",
			spec: `latest:^2\.0\.0`,
			want: "2.0.0",
		},
		{
			name: "match only rc",
			spec: `latest:rc`,
			want: "2.0.0-rc1",
		},
		{
			name: "match alpha or beta",
			spec: `latest:alpha|beta`,
			want: "2.0.0-beta1",
		},
		{
			name:    "no match",
			spec:    `latest:^99\.`,
			wantErr: true,
		},
		{
			name:    "invalid regex",
			spec:    `latest:[invalid`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ResolveVersion(
				tc.spec, sampleVersions, "test", &config.Config{})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Version != tc.want {
				t.Errorf("got %q, want %q", result.Version, tc.want)
			}
		})
	}
}

func TestResolveVersion_LatestRegex_DoesNotMatch150(t *testing.T) {
	// Verify ^1\.5\. does NOT match 1.50.x (regex anchoring).
	versions := []string{"1.5.0", "1.5.1", "1.50.0", "1.50.1"}
	result, err := ResolveVersion(
		`latest:^1\.5\.`, versions, "test", &config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.1" {
		t.Errorf("got %q, want %q — must not match 1.50.x",
			result.Version, "1.5.1")
	}
}

func TestResolveVersion_LatestRegexEmptyVersions(t *testing.T) {
	_, err := ResolveVersion(
		`latest:^1\.5\.`, nil, "test", &config.Config{})
	if err == nil {
		t.Fatal("expected error with empty versions")
	}
}

// ---------------------------------------------------------------------------
// latest-allowed — constraint from .tf files
// ---------------------------------------------------------------------------

func TestResolveVersion_LatestAllowed(t *testing.T) {
	tests := []struct {
		name       string
		tfContent  string
		versions   []string
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name: "pessimistic ~> 1.5.0",
			tfContent: `terraform {
  required_version = "~> 1.5.0"
}`,
			versions: sampleVersions,
			want:     "1.5.7",
		},
		{
			name: "range >= 1.0.0, < 2.0.0",
			tfContent: `terraform {
  required_version = ">= 1.0.0, < 2.0.0"
}`,
			versions: sampleVersions,
			want:     "1.10.3",
		},
		{
			name: "exact = 1.5.0",
			tfContent: `terraform {
  required_version = "= 1.5.0"
}`,
			versions: sampleVersions,
			want:     "1.5.0",
		},
		{
			name: "pessimistic ~> 1.5",
			tfContent: `terraform {
  required_version = "~> 1.5"
}`,
			versions: sampleVersions,
			want:     "1.10.3",
		},
		{
			name: "greater than or equal",
			tfContent: `terraform {
  required_version = ">= 1.9.0"
}`,
			versions: sampleVersions,
			want:     "2.0.0",
		},
		{
			name: "excludes pre-releases",
			tfContent: `terraform {
  required_version = ">= 1.0.0"
}`,
			versions: []string{
				"1.0.0",
				"2.0.0-alpha1",
				"2.0.0-rc1",
				"2.0.0",
			},
			want: "2.0.0",
		},
		{
			name: "no matching version",
			tfContent: `terraform {
  required_version = ">= 99.0.0"
}`,
			versions: sampleVersions,
			wantErr:  true,
		},
		{
			name:       "no tf files",
			tfContent:  "",
			versions:   sampleVersions,
			wantErr:    true,
			errContain: "no required_version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			if tc.tfContent != "" {
				writeFile(t,
					filepath.Join(tmp, "main.tf"),
					tc.tfContent)
			}

			cfg := &config.Config{Dir: tmp}
			result, err := ResolveVersion(
				"latest-allowed", tc.versions, "test", cfg)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContain != "" {
					if got := err.Error(); !contains(got, tc.errContain) {
						t.Errorf("error %q should contain %q",
							got, tc.errContain)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Version != tc.want {
				t.Errorf("got %q, want %q", result.Version, tc.want)
			}
		})
	}
}

func TestResolveVersion_LatestAllowed_MultipleFiles(t *testing.T) {
	tmp := t.TempDir()

	// First file constrains >= 1.5.0
	writeFile(t, filepath.Join(tmp, "providers.tf"), `terraform {
  required_version = ">= 1.5.0"
}`)

	// Second file constrains < 1.8.0
	writeFile(t, filepath.Join(tmp, "backend.tf"), `terraform {
  required_version = "< 1.8.0"
}`)

	cfg := &config.Config{Dir: tmp}
	result, err := ResolveVersion(
		"latest-allowed", sampleVersions, "test", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Intersection: >= 1.5.0 AND < 1.8.0 → newest is 1.7.5
	if result.Version != "1.7.5" {
		t.Errorf("got %q, want %q", result.Version, "1.7.5")
	}
}

func TestResolveVersion_LatestAllowed_InvalidConstraint(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.tf"), `terraform {
  required_version = "not-a-constraint"
}`)

	cfg := &config.Config{Dir: tmp}
	_, err := ResolveVersion(
		"latest-allowed", sampleVersions, "test", cfg)
	if err == nil {
		t.Fatal("expected error for invalid constraint")
	}
}

// ---------------------------------------------------------------------------
// min-required — minimum satisfying version
// ---------------------------------------------------------------------------

func TestResolveVersion_MinRequired(t *testing.T) {
	tests := []struct {
		name      string
		tfContent string
		versions  []string
		want      string
		wantErr   bool
	}{
		{
			name: "pessimistic ~> 1.5.0",
			tfContent: `terraform {
  required_version = "~> 1.5.0"
}`,
			versions: sampleVersions,
			want:     "1.5.0",
		},
		{
			name: "range >= 1.6.0, < 1.8.0",
			tfContent: `terraform {
  required_version = ">= 1.6.0, < 1.8.0"
}`,
			versions: sampleVersions,
			want:     "1.6.0",
		},
		{
			name: "exact = 1.5.0",
			tfContent: `terraform {
  required_version = "= 1.5.0"
}`,
			versions: sampleVersions,
			want:     "1.5.0",
		},
		{
			name: "no matching version",
			tfContent: `terraform {
  required_version = ">= 99.0.0"
}`,
			versions: sampleVersions,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			writeFile(t,
				filepath.Join(tmp, "main.tf"),
				tc.tfContent)

			cfg := &config.Config{Dir: tmp}
			result, err := ResolveVersion(
				"min-required", tc.versions, "test", cfg)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Version != tc.want {
				t.Errorf("got %q, want %q", result.Version, tc.want)
			}
		})
	}
}

func TestResolveVersion_MinRequired_MultipleFiles(t *testing.T) {
	tmp := t.TempDir()

	writeFile(t, filepath.Join(tmp, "providers.tf"), `terraform {
  required_version = ">= 1.5.0"
}`)

	writeFile(t, filepath.Join(tmp, "backend.tf"), `terraform {
  required_version = "< 1.8.0"
}`)

	cfg := &config.Config{Dir: tmp}
	result, err := ResolveVersion(
		"min-required", sampleVersions, "test", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Intersection: >= 1.5.0 AND < 1.8.0 → oldest is 1.5.0
	if result.Version != "1.5.0" {
		t.Errorf("got %q, want %q", result.Version, "1.5.0")
	}
}

func TestResolveVersion_MinRequired_ExcludesPrerelease(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.tf"), `terraform {
  required_version = ">= 1.0.0"
}`)

	versions := []string{
		"2.0.0-alpha1",
		"2.0.0-rc1",
		"2.0.0",
		"2.1.0",
	}

	cfg := &config.Config{Dir: tmp}
	result, err := ResolveVersion(
		"min-required", versions, "test", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "2.0.0" {
		t.Errorf("got %q, want %q — should exclude pre-releases",
			result.Version, "2.0.0")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestResolveVersion_UnrecognisedSpecifier(t *testing.T) {
	_, err := ResolveVersion(
		"foobar", sampleVersions, "test", &config.Config{})
	if err == nil {
		t.Fatal("expected error for unrecognised specifier")
	}
}

func TestResolveVersion_VPrefixStripped(t *testing.T) {
	result, err := ResolveVersion(
		"v1.5.0", sampleVersions, "test", &config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("got %q, want %q", result.Version, "1.5.0")
	}
}

// ---------------------------------------------------------------------------
// findRequiredVersion tests
// ---------------------------------------------------------------------------

func TestFindRequiredVersion_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.tf"), `terraform {
  required_version = "~> 1.5.0"
}`)

	cfg := &config.Config{Dir: tmp}
	got, err := findRequiredVersion(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "~> 1.5.0" {
		t.Errorf("got %q, want %q", got, "~> 1.5.0")
	}
}

func TestFindRequiredVersion_MultipleFilesIntersected(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "a.tf"), `terraform {
  required_version = ">= 1.5.0"
}`)
	writeFile(t, filepath.Join(tmp, "b.tf"), `terraform {
  required_version = "< 2.0.0"
}`)

	cfg := &config.Config{Dir: tmp}
	got, err := findRequiredVersion(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should contain both constraints joined by comma.
	if !contains(got, ">= 1.5.0") || !contains(got, "< 2.0.0") {
		t.Errorf("got %q, want both constraints combined", got)
	}
}

func TestFindRequiredVersion_SkipsComments(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.tf"), `# required_version = ">= 0.12"
// required_version = ">= 0.11"
terraform {
  required_version = "~> 1.5.0"
}`)

	cfg := &config.Config{Dir: tmp}
	got, err := findRequiredVersion(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "~> 1.5.0" {
		t.Errorf("got %q, want %q", got, "~> 1.5.0")
	}
}

func TestFindRequiredVersion_TfJsonFile(t *testing.T) {
	tmp := t.TempDir()
	// .tf.json files can also have required_version.
	writeFile(t, filepath.Join(tmp, "main.tf.json"),
		`{ "terraform": { "required_version": ">= 1.3.0" } }`)

	cfg := &config.Config{Dir: tmp}
	got, err := findRequiredVersion(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(got, ">= 1.3.0") {
		t.Errorf("got %q, want to contain %q", got, ">= 1.3.0")
	}
}

func TestFindRequiredVersion_NoFiles(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{Dir: tmp}
	_, err := findRequiredVersion(cfg)
	if err == nil {
		t.Fatal("expected error when no .tf files exist")
	}
}

func TestFindRequiredVersion_NoConstraintInFiles(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.tf"), `resource "aws_instance" "web" {
  ami = "ami-12345"
}`)

	cfg := &config.Config{Dir: tmp}
	_, err := findRequiredVersion(cfg)
	if err == nil {
		t.Fatal("expected error when no required_version found")
	}
}

// ---------------------------------------------------------------------------
// filterStable tests
// ---------------------------------------------------------------------------

func TestFilterStable(t *testing.T) {
	input := []string{
		"1.0.0",
		"1.1.0-alpha1",
		"1.1.0-beta1",
		"1.1.0-rc1",
		"1.1.0",
		"2.0.0-alpha1",
	}
	got := filterStable(input)
	want := []string{"1.0.0", "1.1.0"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, v, want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// newestVersion / oldestVersion tests
// ---------------------------------------------------------------------------

func TestNewestVersion(t *testing.T) {
	versions := []string{"1.0.0", "1.2.0", "1.1.0", "0.9.0"}
	got := newestVersion(versions)
	if got != "1.2.0" {
		t.Errorf("got %q, want %q", got, "1.2.0")
	}
}

func TestOldestVersion(t *testing.T) {
	versions := []string{"1.0.0", "1.2.0", "1.1.0", "0.9.0"}
	got := oldestVersion(versions)
	if got != "0.9.0" {
		t.Errorf("got %q, want %q", got, "0.9.0")
	}
}

// ---------------------------------------------------------------------------
// extractConstraints tests
// ---------------------------------------------------------------------------

func TestExtractConstraints(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "main.tf")

	content := `terraform {
  required_version = ">= 1.0.0, < 2.0.0"
}

# This module also requires:
# required_version = ">= 0.12"

resource "aws_instance" "web" {
  ami = "ami-12345"
}
`
	writeFile(t, path, content)

	got, err := extractConstraints(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d constraints, want 1: %v", len(got), got)
	}
	if got[0] != ">= 1.0.0, < 2.0.0" {
		t.Errorf("got %q, want %q", got[0], ">= 1.0.0, < 2.0.0")
	}
}

func TestExtractConstraints_MultipleBlocks(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "versions.tf")

	content := `terraform {
  required_version = ">= 1.3.0"
}

# In another block context (unusual but valid for our grep)
terraform {
  required_version = "< 2.0.0"
}
`
	writeFile(t, path, content)

	got, err := extractConstraints(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d constraints, want 2: %v", len(got), got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeVersionFile is a test helper that creates a file. Reuses writeFile
// from resolve_test.go (same package).

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure writeFile from the test package is reachable (it's in
// resolve_test.go in the same package). We reference it above already.
// If the build fails because writeFile is in a different test file,
// the Go test runner compiles all _test.go files in a package together,
// so it will be available.

// ---------------------------------------------------------------------------
// Integration: ResolveVersionFile → ResolveVersion pipeline
// ---------------------------------------------------------------------------

func TestResolveVersionFileThenResolveVersion(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	projectDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a version file specifying "latest".
	writeFile(t, filepath.Join(projectDir, ".terraform-version"), "latest\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       projectDir,
	}

	// Step 1: Resolve version file.
	vr, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("ResolveVersionFile: %v", err)
	}
	if vr.Version != "latest" {
		t.Fatalf("version file returned %q, want %q", vr.Version, "latest")
	}

	// Step 2: Resolve version specifier.
	versions := []string{"1.0.0", "1.1.0", "2.0.0-rc1", "2.0.0"}
	result, err := ResolveVersion(
		vr.Version, versions, vr.Source, cfg)
	if err != nil {
		t.Fatalf("ResolveVersion: %v", err)
	}
	if result.Version != "2.0.0" {
		t.Errorf("got %q, want %q", result.Version, "2.0.0")
	}
}

func TestResolveVersionFileThenResolveVersion_LatestAllowed(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	projectDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write version file.
	writeFile(t, filepath.Join(projectDir, ".terraform-version"),
		"latest-allowed\n")

	// Write .tf file with constraint.
	writeFile(t, filepath.Join(projectDir, "main.tf"), `terraform {
  required_version = "~> 1.5.0"
}`)

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       projectDir,
	}

	vr, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("ResolveVersionFile: %v", err)
	}

	result, err := ResolveVersion(
		vr.Version, sampleVersions, vr.Source, cfg)
	if err != nil {
		t.Fatalf("ResolveVersion: %v", err)
	}
	if result.Version != "1.5.7" {
		t.Errorf("got %q, want %q", result.Version, "1.5.7")
	}
}
