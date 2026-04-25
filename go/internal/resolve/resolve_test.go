package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tfutils/tfenv/go/internal/config"
)

// ---------------------------------------------------------------------------
// parseVersionContent unit tests
// ---------------------------------------------------------------------------

func TestParseVersionContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple version",
			content: "1.5.0\n",
			want:    "1.5.0",
		},
		{
			name:    "latest keyword",
			content: "latest\n",
			want:    "latest",
		},
		{
			name:    "latest with regex",
			content: "latest:^1.5\n",
			want:    "latest:^1.5",
		},
		{
			name:    "v prefix stripped",
			content: "v1.5.0\n",
			want:    "1.5.0",
		},
		{
			name:    "comment stripped",
			content: "1.5.0 # pinned for CI\n",
			want:    "1.5.0",
		},
		{
			name:    "full line comment skipped",
			content: "# this is a comment\n1.5.0\n",
			want:    "1.5.0",
		},
		{
			name:    "carriage return stripped",
			content: "1.5.0\r\n",
			want:    "1.5.0",
		},
		{
			name:    "windows line endings",
			content: "# comment\r\n1.5.0\r\n",
			want:    "1.5.0",
		},
		{
			name:    "leading and trailing whitespace",
			content: "  1.5.0  \n",
			want:    "1.5.0",
		},
		{
			name:    "blank lines skipped",
			content: "\n\n1.5.0\n\n",
			want:    "1.5.0",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
		{
			name:    "only comments",
			content: "# comment\n# another\n",
			want:    "",
		},
		{
			name:    "only whitespace",
			content: "   \n  \n",
			want:    "",
		},
		{
			name:    "v prefix with leading space",
			content: "  v1.5.0\n",
			want:    "1.5.0",
		},
		{
			name:    "min-required keyword",
			content: "min-required\n",
			want:    "min-required",
		},
		{
			name:    "latest-allowed keyword",
			content: "latest-allowed\n",
			want:    "latest-allowed",
		},
		{
			name:    "inline comment with tabs",
			content: "1.5.0\t# pinned\n",
			want:    "1.5.0",
		},
		{
			name:    "multiple versions returns first",
			content: "1.5.0\n1.6.0\n",
			want:    "1.5.0",
		},
		{
			name:    "carriage return only line",
			content: "\r\n1.5.0\n",
			want:    "1.5.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseVersionContent(tc.content)
			if got != tc.want {
				t.Errorf("parseVersionContent(%q) = %q, want %q", tc.content, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// cleanVersion unit tests
// ---------------------------------------------------------------------------

func TestCleanVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.5.0", "1.5.0"},
		{"v1.5.0", "1.5.0"},
		{" 1.5.0 ", "1.5.0"},
		{"1.5.0\r", "1.5.0"},
		{"v1.5.0\r\n", "1.5.0"},
		{"latest", "latest"},
		{"latest:^1.5", "latest:^1.5"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := cleanVersion(tc.input)
			if got != tc.want {
				t.Errorf("cleanVersion(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ResolveVersionFile integration tests
// ---------------------------------------------------------------------------

// writeFile is a test helper that writes content to a file, creating
// parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func TestResolveVersionFile_EnvVarHighestPrecedence(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Place a version file that should be ignored.
	writeFile(t, filepath.Join(startDir, versionFileName), "1.0.0\n")

	cfg := &config.Config{
		TerraformVersion: "1.9.0",
		ConfigDir:        configDir,
		Dir:              startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.9.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.9.0")
	}
	if result.Source != "TFENV_TERRAFORM_VERSION" {
		t.Errorf("Source = %q, want %q", result.Source, "TFENV_TERRAFORM_VERSION")
	}
}

func TestResolveVersionFile_EnvVarVPrefixStripped(t *testing.T) {
	tmp := t.TempDir()
	cfg := &config.Config{
		TerraformVersion: "v1.9.0",
		ConfigDir:        filepath.Join(tmp, "config"),
		Dir:              tmp,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.9.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.9.0")
	}
}

func TestResolveVersionFile_CurrentDir(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "1.5.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
	if result.Source != filepath.Join(startDir, versionFileName) {
		t.Errorf("Source = %q, want %q", result.Source, filepath.Join(startDir, versionFileName))
	}
}

func TestResolveVersionFile_ParentDir(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	parent := filepath.Join(tmp, "project")
	child := filepath.Join(parent, "modules", "vpc")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(parent, versionFileName), "1.4.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       child,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.4.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.4.0")
	}
	if result.Source != filepath.Join(parent, versionFileName) {
		t.Errorf("Source = %q, want %q", result.Source, filepath.Join(parent, versionFileName))
	}
}

func TestResolveVersionFile_GrandparentDir(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	grandchild := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(grandchild, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(tmp, versionFileName), "1.3.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       grandchild,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.3.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.3.0")
	}
}

func TestResolveVersionFile_ConfigDirFallback(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(configDir, "version"), "1.2.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.2.0")
	}
	if result.Source != filepath.Join(configDir, "version") {
		t.Errorf("Source = %q, want %q", result.Source, filepath.Join(configDir, "version"))
	}
}

func TestResolveVersionFile_NoVersionFound(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	_, err := ResolveVersionFile(cfg)
	if err == nil {
		t.Fatal("expected error when no version file found, got nil")
	}
}

func TestResolveVersionFile_EmptyFileReturnsError(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write an empty version file — should be skipped, not returned.
	writeFile(t, filepath.Join(startDir, versionFileName), "")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	_, err := ResolveVersionFile(cfg)
	if err == nil {
		t.Fatal("expected error for empty version file, got nil")
	}
}

func TestResolveVersionFile_CommentOnlyFileSkipped(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	parent := filepath.Join(tmp)
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Comment-only file in startDir — should be skipped.
	writeFile(t, filepath.Join(startDir, versionFileName), "# just a comment\n")
	// Real version in parent.
	writeFile(t, filepath.Join(parent, versionFileName), "1.6.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.6.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.6.0")
	}
}

func TestResolveVersionFile_CarriageReturnsStripped(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "1.5.0\r\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
}

func TestResolveVersionFile_VPrefixInFile(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "v1.5.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
}

func TestResolveVersionFile_CommentsStrippedInFile(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "# this project\n1.5.0 # pinned\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
}

func TestResolveVersionFile_WindowsLineEndings(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "# comment\r\n1.5.0\r\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
}

func TestResolveVersionFile_TFENVDirOverride(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")

	// Two different directories with different versions.
	dirA := filepath.Join(tmp, "dirA")
	dirB := filepath.Join(tmp, "dirB")
	if err := os.MkdirAll(dirA, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dirB, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(dirA, versionFileName), "1.1.0\n")
	writeFile(t, filepath.Join(dirB, versionFileName), "1.2.0\n")

	// Config.Dir is set to dirB — should use dirB's version.
	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       dirB,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.2.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.2.0")
	}
}

func TestResolveVersionFile_LatestKeyword(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "latest:^1.5\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "latest:^1.5" {
		t.Errorf("Version = %q, want %q", result.Version, "latest:^1.5")
	}
}

func TestResolveVersionFile_PrecedenceLocalOverParent(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	parent := filepath.Join(tmp, "project")
	child := filepath.Join(parent, "sub")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(parent, versionFileName), "1.0.0\n")
	writeFile(t, filepath.Join(child, versionFileName), "1.1.0\n")

	cfg := &config.Config{
		ConfigDir: configDir,
		Dir:       child,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.1.0" {
		t.Errorf("Version = %q, want %q (local should override parent)", result.Version, "1.1.0")
	}
}

func TestResolveVersionFile_PrecedenceEnvOverLocal(t *testing.T) {
	tmp := t.TempDir()
	configDir := filepath.Join(tmp, "config")
	startDir := filepath.Join(tmp, "project")
	if err := os.MkdirAll(startDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(startDir, versionFileName), "1.0.0\n")

	cfg := &config.Config{
		TerraformVersion: "1.9.0",
		ConfigDir:        configDir,
		Dir:              startDir,
	}

	result, err := ResolveVersionFile(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Version != "1.9.0" {
		t.Errorf("Version = %q, want %q (env should override local file)", result.Version, "1.9.0")
	}
	if result.Source != "TFENV_TERRAFORM_VERSION" {
		t.Errorf("Source = %q, want TFENV_TERRAFORM_VERSION", result.Source)
	}
}

// ---------------------------------------------------------------------------
// findVersionFileWalk unit tests
// ---------------------------------------------------------------------------

func TestFindVersionFileWalk_StopsAtRoot(t *testing.T) {
	// Walk from a temp dir with no version file — should return nil, nil.
	tmp := t.TempDir()
	deep := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := findVersionFileWalk(deep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result when no version file exists, got %+v", result)
	}
}

func TestFindVersionFileWalk_FindsFileInStartDir(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, versionFileName), "1.5.0\n")

	result, err := findVersionFileWalk(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Version != "1.5.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.5.0")
	}
}

func TestFindVersionFileWalk_SkipsEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	child := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}

	// Empty file in child — should be skipped, find parent's.
	writeFile(t, filepath.Join(child, versionFileName), "")
	writeFile(t, filepath.Join(tmp, versionFileName), "1.4.0\n")

	result, err := findVersionFileWalk(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Version != "1.4.0" {
		t.Errorf("Version = %q, want %q", result.Version, "1.4.0")
	}
}
