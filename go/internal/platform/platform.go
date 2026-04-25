// Package platform detects the current OS, architecture, and platform-specific behaviour.
package platform

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// Platform represents the OS and architecture for Terraform binary downloads.
type Platform struct {
	OS   string // Terraform kernel string (darwin, linux, windows, freebsd)
	Arch string // Terraform arch string, respects TFENV_ARCH override
}

// Detect returns the current platform, using runtime.GOOS and runtime.GOARCH
// mapped to Terraform's naming conventions. If archOverride is non-empty, it
// is used instead of the detected architecture.
func Detect(archOverride string) Platform {
	p := Platform{
		OS:   mapOS(runtime.GOOS),
		Arch: mapArch(runtime.GOARCH),
	}
	if archOverride != "" {
		p.Arch = archOverride
	}
	return p
}

// String returns the platform in "os_arch" format (e.g. "linux_amd64").
func (p Platform) String() string {
	return fmt.Sprintf("%s_%s", p.OS, p.Arch)
}

// DownloadArch returns the architecture string to use for downloading a given
// Terraform version. When the arch is arm64, certain old versions don't have
// ARM64 binaries available, so this method falls back to amd64 for those.
func (p Platform) DownloadArch(version string) string {
	if p.Arch != "arm64" {
		return p.Arch
	}

	switch p.OS {
	case "linux":
		if needsLinuxARM64Fallback(version) {
			return "amd64"
		}
	case "darwin":
		if needsDarwinARM64Fallback(version) {
			return "amd64"
		}
	}

	return p.Arch
}

// needsLinuxARM64Fallback returns true if the given version requires falling
// back from arm64 to amd64 on Linux.
//
// Fallback ranges:
//   - Versions < 0.11.15
//   - Versions >= 0.12.0 and < 0.12.30
//   - Versions >= 0.13.0 and < 0.13.5
func needsLinuxARM64Fallback(version string) bool {
	major, minor, patch, ok := parseVersion(version)
	if !ok {
		return false
	}

	v := semver{major, minor, patch}

	// < 0.11.15
	if v.less(semver{0, 11, 15}) {
		return true
	}

	// >= 0.12.0 and < 0.12.30
	if !v.less(semver{0, 12, 0}) && v.less(semver{0, 12, 30}) {
		return true
	}

	// >= 0.13.0 and < 0.13.5
	if !v.less(semver{0, 13, 0}) && v.less(semver{0, 13, 5}) {
		return true
	}

	return false
}

// needsDarwinARM64Fallback returns true if the given version requires falling
// back from arm64 to amd64 on macOS (Apple Silicon).
//
// Fallback range: versions < 1.0.2
func needsDarwinARM64Fallback(version string) bool {
	major, minor, patch, ok := parseVersion(version)
	if !ok {
		return false
	}
	return (semver{major, minor, patch}).less(semver{1, 0, 2})
}

// semver is a minimal semver triple for comparison purposes.
type semver struct {
	major, minor, patch int
}

// less returns true if s < other.
func (s semver) less(other semver) bool {
	if s.major != other.major {
		return s.major < other.major
	}
	if s.minor != other.minor {
		return s.minor < other.minor
	}
	return s.patch < other.patch
}

// parseVersion extracts major.minor.patch from a version string. It strips
// any leading "v" and ignores pre-release suffixes (e.g. "-rc1").
func parseVersion(version string) (major, minor, patch int, ok bool) {
	version = strings.TrimPrefix(version, "v")

	// Strip pre-release suffix (e.g. "-alpha1", "-rc1")
	if idx := strings.IndexByte(version, '-'); idx >= 0 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, false
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, false
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, false
	}

	return major, minor, patch, true
}

// mapOS maps runtime.GOOS to the Terraform download kernel string.
func mapOS(goos string) string {
	switch goos {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	case "freebsd":
		return "freebsd"
	default:
		return goos
	}
}

// mapArch maps runtime.GOARCH to the Terraform download arch string.
func mapArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "386":
		return "386"
	case "arm":
		return "arm"
	default:
		return goarch
	}
}
