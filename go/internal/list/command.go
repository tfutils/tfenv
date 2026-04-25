package list

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/logging"
	"github.com/tfutils/tfenv/go/internal/resolve"
)

func init() {
	cli.Register("list", "List installed Terraform versions", RunList)
	cli.Register("list-remote", "List installable Terraform versions", RunListRemote)
}

// RunList implements the tfenv list command. It displays all locally installed
// Terraform versions with architecture and active-version marking.
// Returns 0 on success, 1 on error.
func RunList(args []string) int {
	if len(args) > 0 {
		logging.Error("usage: tfenv list")
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	versions, err := ListLocal(cfg)
	if err != nil {
		logging.Error("failed to list installed versions", "err", err)
		return 1
	}

	if len(versions) == 0 {
		fmt.Fprintln(os.Stderr, "No versions available. Please install one with: tfenv install")
		return 1
	}

	// Resolve active version (may fail if no version is set).
	var activeVersion, versionSource string
	result, err := resolve.ResolveVersionFile(cfg)
	if err == nil {
		activeVersion = result.Version
		versionSource = result.Source
	}

	defaultSet := false
	for _, v := range versions {
		binaryPath := filepath.Join(cfg.ConfigDir, "versions", v, "terraform")
		arch := detectBinaryArch(binaryPath)
		if v == activeVersion {
			fmt.Fprintf(os.Stdout, "* %s (%s) (set by %s)\n", v, arch, versionSource)
			defaultSet = true
		} else {
			fmt.Fprintf(os.Stdout, "  %s (%s)\n", v, arch)
		}
	}

	if !defaultSet {
		logging.Info("No default set. Set with 'tfenv use <version>'")
	}

	return 0
}

// RunListRemote implements the tfenv list-remote command. It displays
// available Terraform versions from the remote mirror, optionally
// filtered by a regex pattern.
// Returns 0 on success, 1 on error.
func RunListRemote(args []string) int {
	if len(args) > 1 {
		logging.Error("usage: tfenv list-remote [<regex>]")
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		logging.Error("failed to load configuration", "err", err)
		return 1
	}

	versions, err := ListRemote(cfg)
	if err != nil {
		logging.Error("failed to list remote versions", "err", err)
		return 1
	}

	// Apply regex filter if provided.
	if len(args) == 1 {
		re, err := regexp.Compile(args[0])
		if err != nil {
			logging.Error("invalid regex", "pattern", args[0], "err", err)
			return 1
		}
		var filtered []string
		for _, v := range versions {
			if re.MatchString(v) {
				filtered = append(filtered, v)
			}
		}
		versions = filtered
	}

	for _, v := range versions {
		fmt.Fprintln(os.Stdout, v)
	}

	return 0
}

// detectBinaryArch inspects the binary at the given path and returns its
// architecture string. It tries ELF, Mach-O, and PE formats in order.
// Returns "unknown" if the file cannot be read or its format is not recognised.
func detectBinaryArch(binaryPath string) string {
	// Try ELF (Linux).
	if f, err := elf.Open(binaryPath); err == nil {
		machine := f.Machine
		f.Close()
		switch machine {
		case elf.EM_X86_64:
			return "amd64"
		case elf.EM_AARCH64:
			return "arm64"
		case elf.EM_386:
			return "386"
		case elf.EM_ARM:
			return "arm"
		default:
			return "unknown"
		}
	}

	// Try Mach-O (macOS).
	if f, err := macho.Open(binaryPath); err == nil {
		cpu := f.Cpu
		f.Close()
		switch cpu {
		case macho.CpuAmd64:
			return "amd64"
		case macho.CpuArm64:
			return "arm64"
		default:
			return "unknown"
		}
	}

	// Try PE (Windows).
	if f, err := pe.Open(binaryPath); err == nil {
		machine := f.Machine
		f.Close()
		switch machine {
		case pe.IMAGE_FILE_MACHINE_AMD64:
			return "amd64"
		case pe.IMAGE_FILE_MACHINE_ARM64:
			return "arm64"
		case pe.IMAGE_FILE_MACHINE_I386:
			return "386"
		default:
			return "unknown"
		}
	}

	return "unknown"
}
