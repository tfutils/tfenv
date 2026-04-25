package platform

import (
	"testing"
)

func TestDetect_UsesOverride(t *testing.T) {
	p := Detect("arm64")
	if p.Arch != "arm64" {
		t.Errorf("Detect(arm64).Arch = %q, want %q", p.Arch, "arm64")
	}
}

func TestDetect_EmptyOverrideUsesRuntime(t *testing.T) {
	p := Detect("")
	if p.OS == "" {
		t.Error("Detect(\"\").OS is empty, expected a value from runtime.GOOS")
	}
	if p.Arch == "" {
		t.Error("Detect(\"\").Arch is empty, expected a value from runtime.GOARCH")
	}
}

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		os   string
		arch string
		want string
	}{
		{"linux", "amd64", "linux_amd64"},
		{"darwin", "arm64", "darwin_arm64"},
		{"windows", "386", "windows_386"},
		{"freebsd", "arm", "freebsd_arm"},
	}

	for _, tt := range tests {
		p := Platform{OS: tt.os, Arch: tt.arch}
		got := p.String()
		if got != tt.want {
			t.Errorf("Platform{%q, %q}.String() = %q, want %q", tt.os, tt.arch, got, tt.want)
		}
	}
}

func TestMapOS(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"darwin", "darwin"},
		{"linux", "linux"},
		{"windows", "windows"},
		{"freebsd", "freebsd"},
		{"openbsd", "openbsd"}, // unknown passes through
	}

	for _, tt := range tests {
		got := mapOS(tt.goos)
		if got != tt.want {
			t.Errorf("mapOS(%q) = %q, want %q", tt.goos, got, tt.want)
		}
	}
}

func TestMapArch(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "amd64"},
		{"arm64", "arm64"},
		{"386", "386"},
		{"arm", "arm"},
		{"mips", "mips"}, // unknown passes through
	}

	for _, tt := range tests {
		got := mapArch(tt.goarch)
		if got != tt.want {
			t.Errorf("mapArch(%q) = %q, want %q", tt.goarch, got, tt.want)
		}
	}
}

func TestDownloadArch_NonARM64Passthrough(t *testing.T) {
	tests := []struct {
		arch    string
		version string
	}{
		{"amd64", "0.10.0"},
		{"386", "0.12.0"},
		{"arm", "1.0.0"},
	}

	for _, tt := range tests {
		p := Platform{OS: "linux", Arch: tt.arch}
		got := p.DownloadArch(tt.version)
		if got != tt.arch {
			t.Errorf("Platform{linux, %q}.DownloadArch(%q) = %q, want %q",
				tt.arch, tt.version, got, tt.arch)
		}
	}
}

func TestDownloadArch_LinuxARM64Fallback(t *testing.T) {
	tests := []struct {
		version string
		want    string
		desc    string
	}{
		// Versions < 0.11.15 → fallback to amd64
		{"0.10.0", "amd64", "pre-0.11 falls back"},
		{"0.11.0", "amd64", "0.11.0 falls back"},
		{"0.11.14", "amd64", "0.11.14 falls back"},

		// 0.11.15 is the boundary → arm64
		{"0.11.15", "arm64", "0.11.15 uses arm64"},
		{"0.11.16", "arm64", "0.11.16 uses arm64"},

		// Versions >= 0.12.0 and < 0.12.30 → fallback to amd64
		{"0.12.0", "amd64", "0.12.0 falls back"},
		{"0.12.15", "amd64", "0.12.15 falls back"},
		{"0.12.29", "amd64", "0.12.29 falls back"},

		// 0.12.30 is the boundary → arm64
		{"0.12.30", "arm64", "0.12.30 uses arm64"},
		{"0.12.31", "arm64", "0.12.31 uses arm64"},

		// Versions >= 0.13.0 and < 0.13.5 → fallback to amd64
		{"0.13.0", "amd64", "0.13.0 falls back"},
		{"0.13.4", "amd64", "0.13.4 falls back"},

		// 0.13.5 is the boundary → arm64
		{"0.13.5", "arm64", "0.13.5 uses arm64"},
		{"0.13.6", "arm64", "0.13.6 uses arm64"},

		// Modern versions → arm64
		{"0.14.0", "arm64", "0.14.0 uses arm64"},
		{"1.0.0", "arm64", "1.0.0 uses arm64"},
		{"1.5.0", "arm64", "1.5.0 uses arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := Platform{OS: "linux", Arch: "arm64"}
			got := p.DownloadArch(tt.version)
			if got != tt.want {
				t.Errorf("DownloadArch(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestDownloadArch_DarwinARM64Fallback(t *testing.T) {
	tests := []struct {
		version string
		want    string
		desc    string
	}{
		// Versions < 1.0.2 → fallback to amd64
		{"0.10.0", "amd64", "0.10.0 falls back"},
		{"0.14.11", "amd64", "0.14.11 falls back"},
		{"1.0.0", "amd64", "1.0.0 falls back"},
		{"1.0.1", "amd64", "1.0.1 falls back"},

		// 1.0.2 is the boundary → arm64
		{"1.0.2", "arm64", "1.0.2 uses arm64"},
		{"1.0.3", "arm64", "1.0.3 uses arm64"},
		{"1.5.0", "arm64", "1.5.0 uses arm64"},
		{"2.0.0", "arm64", "2.0.0 uses arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			p := Platform{OS: "darwin", Arch: "arm64"}
			got := p.DownloadArch(tt.version)
			if got != tt.want {
				t.Errorf("DownloadArch(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestDownloadArch_WindowsARM64NoFallback(t *testing.T) {
	// Windows ARM64 has no special fallback logic.
	p := Platform{OS: "windows", Arch: "arm64"}
	got := p.DownloadArch("0.10.0")
	if got != "arm64" {
		t.Errorf("Windows ARM64 DownloadArch(0.10.0) = %q, want %q", got, "arm64")
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input              string
		wantMajor          int
		wantMinor          int
		wantPatch          int
		wantOK             bool
	}{
		{"1.0.0", 1, 0, 0, true},
		{"0.12.30", 0, 12, 30, true},
		{"0.13.5", 0, 13, 5, true},
		{"v1.5.3", 1, 5, 3, true},
		{"1.0.0-rc1", 1, 0, 0, true},
		{"1.0.0-alpha20210623", 1, 0, 0, true},
		{"", 0, 0, 0, false},
		{"1.0", 0, 0, 0, false},
		{"abc", 0, 0, 0, false},
		{"1.2.abc", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			major, minor, patch, ok := parseVersion(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseVersion(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if major != tt.wantMajor || minor != tt.wantMinor || patch != tt.wantPatch {
				t.Errorf("parseVersion(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.input, major, minor, patch, tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
		})
	}
}

func TestSemver_Less(t *testing.T) {
	tests := []struct {
		a    semver
		b    semver
		want bool
	}{
		{semver{0, 0, 0}, semver{0, 0, 1}, true},
		{semver{0, 0, 1}, semver{0, 0, 0}, false},
		{semver{0, 0, 0}, semver{0, 0, 0}, false},
		{semver{0, 1, 0}, semver{0, 2, 0}, true},
		{semver{0, 2, 0}, semver{0, 1, 0}, false},
		{semver{1, 0, 0}, semver{2, 0, 0}, true},
		{semver{2, 0, 0}, semver{1, 0, 0}, false},
		{semver{0, 11, 14}, semver{0, 11, 15}, true},
		{semver{0, 11, 15}, semver{0, 11, 15}, false},
		{semver{0, 12, 29}, semver{0, 12, 30}, true},
		{semver{0, 13, 4}, semver{0, 13, 5}, true},
		{semver{1, 0, 1}, semver{1, 0, 2}, true},
	}

	for _, tt := range tests {
		got := tt.a.less(tt.b)
		if got != tt.want {
			t.Errorf("semver{%d,%d,%d}.less(semver{%d,%d,%d}) = %v, want %v",
				tt.a.major, tt.a.minor, tt.a.patch,
				tt.b.major, tt.b.minor, tt.b.patch,
				got, tt.want)
		}
	}
}

func TestDownloadArch_InvalidVersion(t *testing.T) {
	// Invalid versions should not cause fallback — return the arch as-is.
	p := Platform{OS: "linux", Arch: "arm64"}
	got := p.DownloadArch("invalid")
	if got != "arm64" {
		t.Errorf("DownloadArch(invalid) = %q, want %q", got, "arm64")
	}
}

func TestDownloadArch_PreReleaseVersion(t *testing.T) {
	// Pre-release versions should parse correctly.
	p := Platform{OS: "darwin", Arch: "arm64"}

	// 1.0.2-rc1 should be >= 1.0.2 base, so arm64
	got := p.DownloadArch("1.0.2-rc1")
	if got != "arm64" {
		t.Errorf("DownloadArch(1.0.2-rc1) = %q, want %q", got, "arm64")
	}

	// 1.0.1-rc1 base is 1.0.1 which is < 1.0.2, so amd64
	got = p.DownloadArch("1.0.1-rc1")
	if got != "amd64" {
		t.Errorf("DownloadArch(1.0.1-rc1) = %q, want %q", got, "amd64")
	}
}
