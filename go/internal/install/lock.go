package install

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tfutils/tfenv/go/internal/logging"
)

const (
	// lockRetryInterval is the time between lock acquisition attempts.
	lockRetryInterval = 1 * time.Second

	// lockMaxRetries is the maximum number of lock acquisition attempts.
	lockMaxRetries = 60

	// lockStaleThreshold is the age at which a lock is considered stale
	// and will be forcibly removed. A 10-minute threshold accounts for
	// slow network downloads while still recovering from crashed processes.
	lockStaleThreshold = 10 * time.Minute
)

// lockDir returns the path to the lock directory for a given version.
func lockDir(configDir, version string) string {
	return filepath.Join(configDir, fmt.Sprintf(".install-lock-%s", version))
}

// acquireLock attempts to acquire a filesystem-based lock for installing the
// given version. It uses os.Mkdir as the atomic operation — the call succeeds
// only if the directory does not already exist. Returns the lock path on
// success for use with releaseLock.
func acquireLock(configDir, version string) (string, error) {
	dir := lockDir(configDir, version)

	for attempt := 0; attempt < lockMaxRetries; attempt++ {
		err := os.Mkdir(dir, 0o755)
		if err == nil {
			logging.Debug("acquired install lock", "version", version, "path", dir)
			return dir, nil
		}

		if !os.IsExist(err) {
			return "", fmt.Errorf("creating lock directory %q: %w", dir, err)
		}

		// Lock exists — check if stale.
		if removedStale := removeIfStale(dir); removedStale {
			// Stale lock removed; retry immediately.
			continue
		}

		if attempt == 0 {
			logging.Info("Another process is installing Terraform. Waiting...",
				"version", version)
		}

		time.Sleep(lockRetryInterval)
	}

	return "", fmt.Errorf(
		"timed out waiting for install lock for Terraform %s after %d attempts",
		version, lockMaxRetries,
	)
}

// releaseLock removes the lock directory. It is safe to call multiple times.
func releaseLock(dir string) {
	if dir == "" {
		return
	}
	if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to release install lock", "path", dir, "err", err)
	}
}

// removeIfStale checks if the lock directory's modification time exceeds
// the stale threshold and removes it if so. Returns true if a stale lock
// was removed.
func removeIfStale(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	age := time.Since(info.ModTime())
	if age <= lockStaleThreshold {
		return false
	}

	logging.Warn("removing stale install lock",
		"path", dir, "age", age.Round(time.Second))

	if err := os.Remove(dir); err != nil {
		logging.Warn("failed to remove stale lock", "path", dir, "err", err)
		return false
	}

	return true
}
