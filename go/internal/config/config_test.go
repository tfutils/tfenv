package config

import (
	"os"
	"path/filepath"
	"testing"
)

// clearEnv unsets all TFENV_* env vars to ensure a clean test environment.
func clearEnv(t *testing.T) {
	t.Helper()
	vars := []string{
		"TFENV_CONFIG_DIR",
		"TFENV_ROOT",
		"TFENV_REMOTE",
		"TFENV_ARCH",
		"TFENV_AUTO_INSTALL",
		"TFENV_CURL_OUTPUT",
		"TFENV_TERRAFORM_VERSION",
		"TFENV_NETRC_PATH",
		"TFENV_REVERSE_REMOTE",
		"TFENV_SORT_VERSIONS_REMOTE",
		"TFENV_SKIP_REMOTE_CHECK",
		"TFENV_DIR",
		"TFENV_PGP_KEY_PATH",
		"TFENV_LOG_LEVEL",
		"TFENV_LOG_FORMAT",
	}
	for _, v := range vars {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	tmpDir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.ConfigDir != tmpDir {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir, tmpDir)
	}
	if cfg.Remote != "https://releases.hashicorp.com" {
		t.Errorf("Remote = %q, want %q", cfg.Remote, "https://releases.hashicorp.com")
	}
	if cfg.AutoInstall != true {
		t.Errorf("AutoInstall = %v, want true", cfg.AutoInstall)
	}
	if cfg.CurlOutput != 0 {
		t.Errorf("CurlOutput = %d, want 0", cfg.CurlOutput)
	}
	if cfg.Root != "" {
		t.Errorf("Root = %q, want empty", cfg.Root)
	}
	if cfg.Arch != "" {
		t.Errorf("Arch = %q, want empty", cfg.Arch)
	}
	if cfg.TerraformVersion != "" {
		t.Errorf("TerraformVersion = %q, want empty", cfg.TerraformVersion)
	}
	if cfg.ReverseRemote != false {
		t.Errorf("ReverseRemote = %v, want false", cfg.ReverseRemote)
	}
	if cfg.SortVersionsRemote != false {
		t.Errorf("SortVersionsRemote = %v, want false", cfg.SortVersionsRemote)
	}
	if cfg.SkipRemoteCheck != false {
		t.Errorf("SkipRemoteCheck = %v, want false", cfg.SkipRemoteCheck)
	}
}

func TestLoad_DefaultConfigDir(t *testing.T) {
	clearEnv(t)

	// Use a temp dir as HOME to avoid polluting the real home directory.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	want := filepath.Join(tmpHome, ".tfenv")
	if cfg.ConfigDir != want {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir, want)
	}
}

func TestLoad_ConfigDirAutoCreated(t *testing.T) {
	clearEnv(t)

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "newdir", ".tfenv")
	t.Setenv("TFENV_CONFIG_DIR", configDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.ConfigDir != configDir {
		t.Errorf("ConfigDir = %q, want %q", cfg.ConfigDir, configDir)
	}

	// Verify the directory was actually created.
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("ConfigDir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("ConfigDir is not a directory")
	}

	// Verify versions/ subdirectory was created.
	versionsDir := filepath.Join(configDir, "versions")
	info, err = os.Stat(versionsDir)
	if err != nil {
		t.Fatalf("versions/ not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("versions/ is not a directory")
	}
}

func TestLoad_StringEnvVars(t *testing.T) {
	clearEnv(t)

	tmpDir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", tmpDir)
	t.Setenv("TFENV_ROOT", "/opt/tfenv")
	t.Setenv("TFENV_REMOTE", "https://my-mirror.example.com")
	t.Setenv("TFENV_ARCH", "arm64")
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.5.0")
	t.Setenv("TFENV_NETRC_PATH", "/home/user/.netrc")
	t.Setenv("TFENV_DIR", "/workspace")
	t.Setenv("TFENV_PGP_KEY_PATH", "/keys/custom.pgp")
	t.Setenv("TFENV_LOG_LEVEL", "debug")
	t.Setenv("TFENV_LOG_FORMAT", "json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Root", cfg.Root, "/opt/tfenv"},
		{"Remote", cfg.Remote, "https://my-mirror.example.com"},
		{"Arch", cfg.Arch, "arm64"},
		{"TerraformVersion", cfg.TerraformVersion, "1.5.0"},
		{"NetrcPath", cfg.NetrcPath, "/home/user/.netrc"},
		{"Dir", cfg.Dir, "/workspace"},
		{"PGPKeyPath", cfg.PGPKeyPath, "/keys/custom.pgp"},
		{"LogLevel", cfg.LogLevel, "debug"},
		{"LogFormat", cfg.LogFormat, "json"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestLoad_AutoInstallParsing(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    bool
		wantErr bool
	}{
		{"empty defaults true", "", true, false},
		{"true", "true", true, false},
		{"TRUE", "TRUE", true, false},
		{"1", "1", true, false},
		{"false", "false", false, false},
		{"FALSE", "FALSE", false, false},
		{"0", "0", false, false},
		{"invalid", "yes", false, true},
		{"invalid number", "2", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			tmpDir := t.TempDir()
			t.Setenv("TFENV_CONFIG_DIR", tmpDir)

			if tt.envVal != "" {
				t.Setenv("TFENV_AUTO_INSTALL", tt.envVal)
			}

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for TFENV_AUTO_INSTALL=%q, got nil", tt.envVal)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.AutoInstall != tt.want {
				t.Errorf("AutoInstall = %v, want %v", cfg.AutoInstall, tt.want)
			}
		})
	}
}

func TestLoad_CurlOutputParsing(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    int
		wantErr bool
	}{
		{"empty defaults 0", "", 0, false},
		{"0", "0", 0, false},
		{"1", "1", 1, false},
		{"2", "2", 2, false},
		{"invalid", "abc", 0, true},
		{"float", "1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)

			tmpDir := t.TempDir()
			t.Setenv("TFENV_CONFIG_DIR", tmpDir)

			if tt.envVal != "" {
				t.Setenv("TFENV_CURL_OUTPUT", tt.envVal)
			}

			cfg, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for TFENV_CURL_OUTPUT=%q, got nil", tt.envVal)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.CurlOutput != tt.want {
				t.Errorf("CurlOutput = %d, want %d", cfg.CurlOutput, tt.want)
			}
		})
	}
}

func TestLoad_BoolEnvVars(t *testing.T) {
	clearEnv(t)

	tmpDir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", tmpDir)
	t.Setenv("TFENV_REVERSE_REMOTE", "true")
	t.Setenv("TFENV_SORT_VERSIONS_REMOTE", "1")
	t.Setenv("TFENV_SKIP_REMOTE_CHECK", "TRUE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if cfg.ReverseRemote != true {
		t.Errorf("ReverseRemote = %v, want true", cfg.ReverseRemote)
	}
	if cfg.SortVersionsRemote != true {
		t.Errorf("SortVersionsRemote = %v, want true", cfg.SortVersionsRemote)
	}
	if cfg.SkipRemoteCheck != true {
		t.Errorf("SkipRemoteCheck = %v, want true", cfg.SkipRemoteCheck)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input      string
		defaultVal bool
		want       bool
		wantErr    bool
	}{
		{"", true, true, false},
		{"", false, false, false},
		{"true", false, true, false},
		{"True", false, true, false},
		{"TRUE", false, true, false},
		{"1", false, true, false},
		{"false", true, false, false},
		{"False", true, false, false},
		{"FALSE", true, false, false},
		{"0", true, false, false},
		{"yes", false, false, true},
		{"no", true, false, true},
		{"2", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBool(tt.input, tt.defaultVal)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseBool(%q, %v) = %v, want %v", tt.input, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input      string
		defaultVal int
		want       int
		wantErr    bool
	}{
		{"", 0, 0, false},
		{"", 5, 5, false},
		{"0", 5, 0, false},
		{"1", 0, 1, false},
		{"42", 0, 42, false},
		{"-1", 0, -1, false},
		{"abc", 0, 0, true},
		{"1.5", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseInt(tt.input, tt.defaultVal)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseInt(%q, %d) = %d, want %d", tt.input, tt.defaultVal, got, tt.want)
			}
		})
	}
}
