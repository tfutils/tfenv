package install

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract extracts the terraform binary from zipPath into destDir.
// It enforces zip-slip protection by rejecting entries with ".." path
// components or absolute paths. The extracted binary gets 0755 permissions.
func Extract(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("opening zip archive: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	found := false
	for _, f := range r.File {
		if err := validateZipEntry(f.Name); err != nil {
			return err
		}

		name := filepath.Base(f.Name)
		if !isTerraformBinary(name) {
			continue
		}

		destPath := filepath.Join(destDir, name)

		// Verify the resolved path is still within destDir (belt-and-suspenders).
		resolved, err := filepath.Abs(destPath)
		if err != nil {
			return fmt.Errorf("resolving path for %q: %w", f.Name, err)
		}
		absDestDir, err := filepath.Abs(destDir)
		if err != nil {
			return fmt.Errorf("resolving destination directory: %w", err)
		}
		if !strings.HasPrefix(resolved, absDestDir+string(os.PathSeparator)) && resolved != absDestDir {
			return fmt.Errorf("zip entry %q resolves outside destination directory (zip-slip)", f.Name)
		}

		if err := extractFile(f, destPath); err != nil {
			return err
		}
		found = true
	}

	if !found {
		return fmt.Errorf("no terraform binary found in zip archive")
	}
	return nil
}

// validateZipEntry checks a zip entry name for path traversal attacks.
func validateZipEntry(name string) error {
	if filepath.IsAbs(name) {
		return fmt.Errorf("zip entry %q has absolute path (zip-slip)", name)
	}
	clean := filepath.ToSlash(filepath.Clean(name))
	if strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return fmt.Errorf("zip entry %q contains path traversal (zip-slip)", name)
	}
	return nil
}

// isTerraformBinary returns true if the filename is a terraform binary.
func isTerraformBinary(name string) bool {
	return name == "terraform" || name == "terraform.exe"
}

// extractFile writes a single zip file entry to destPath with 0755 permissions.
func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("opening zip entry %q: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", destPath, err)
	}
	defer out.Close()

	// Limit extraction size to 1GB to prevent zip bombs.
	limited := io.LimitReader(rc, 1<<30)
	if _, err := io.Copy(out, limited); err != nil {
		return fmt.Errorf("extracting %q: %w", f.Name, err)
	}
	return nil
}
