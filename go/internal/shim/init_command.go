package shim

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/logging"
)

func init() {
	cli.Register("init",
		"Create terraform hardlink/copy next to tfenv binary",
		RunInit)
}

// RunInit creates a "terraform" hardlink (or copy) next to the running
// tfenv binary, enabling the multi-call binary to intercept terraform
// invocations. It is idempotent — re-running when the link already
// exists and points to the same file is a successful no-op.
func RunInit(args []string) int {
	exe, err := os.Executable()
	if err != nil {
		logging.Error("failed to determine executable path",
			"err", err)
		return 1
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		logging.Error("failed to resolve symlinks",
			"path", exe, "err", err)
		return 1
	}

	dir := filepath.Dir(exe)
	terraformName := "terraform"
	if runtime.GOOS == "windows" {
		terraformName = "terraform.exe"
	}
	terraformPath := filepath.Join(dir, terraformName)

	// Idempotent: if terraform already exists and is the same file,
	// nothing to do.
	if isSameFile(exe, terraformPath) {
		fmt.Fprintf(os.Stderr, "Already exists: %s\n", terraformPath)
		return 0
	}

	// Remove a stale terraform binary if present (might be from an
	// older version or a different binary entirely).
	if _, statErr := os.Stat(terraformPath); statErr == nil {
		if removeErr := os.Remove(terraformPath); removeErr != nil {
			logging.Error("failed to remove existing terraform",
				"path", terraformPath, "err", removeErr)
			return 1
		}
	}

	// Try hardlink first — same inode, no extra disk space.
	if linkErr := os.Link(exe, terraformPath); linkErr != nil {
		logging.Debug("hardlink failed, falling back to copy",
			"err", linkErr)

		if copyErr := copyFile(exe, terraformPath); copyErr != nil {
			logging.Error("failed to create terraform binary",
				"path", terraformPath, "err", copyErr)
			return 1
		}
	}

	fmt.Fprintf(os.Stderr, "Created %s\n", terraformPath)
	return 0
}

// isSameFile checks whether two paths refer to the same underlying file
// (same device and inode on Unix).
func isSameFile(a, b string) bool {
	infoA, errA := os.Stat(a)
	infoB, errB := os.Stat(b)
	if errA != nil || errB != nil {
		return false
	}
	return os.SameFile(infoA, infoB)
}

// copyFile copies src to dst, preserving the source's permission bits.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source %s: %w", src, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat source %s: %w", src, err)
	}

	out, err := os.OpenFile(
		dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("creating destination %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %s to %s: %w", src, dst, err)
	}

	return nil
}
