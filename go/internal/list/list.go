// Package list provides local and remote Terraform version listing.
package list

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	"golang.org/x/net/html"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/logging"
)

const (
	httpTimeout = 30 * time.Second
	userAgent   = "tfenv/go"
)

// ListRemote fetches available Terraform versions from the configured remote.
// Returns versions sorted by semver (newest first by default).
// cfg.ReverseRemote=true returns oldest first.
func ListRemote(cfg *config.Config) ([]string, error) {
	url := strings.TrimSuffix(cfg.Remote, "/") + "/terraform/"
	logging.Debug("fetching remote version index", "url", url)

	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", url, err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching remote index from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote index returned HTTP %d from %s", resp.StatusCode, url)
	}

	versions, err := parseVersionIndex(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing remote version index: %w", err)
	}

	sortVersions(versions, cfg.ReverseRemote)
	logging.Debug("fetched remote versions", "count", len(versions))
	return versions, nil
}

// ListLocal returns locally installed versions from ConfigDir/versions/.
// Each subdirectory name is an installed version. Returns versions sorted
// by semver (newest first by default). cfg.ReverseRemote=true returns oldest first.
func ListLocal(cfg *config.Config) ([]string, error) {
	versionsDir := filepath.Join(cfg.ConfigDir, "versions")
	logging.Debug("listing local versions", "dir", versionsDir)

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading versions directory %s: %w", versionsDir, err)
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, err := goversion.NewVersion(name); err == nil {
			versions = append(versions, name)
		}
	}

	sortVersions(versions, cfg.ReverseRemote)
	logging.Debug("found local versions", "count", len(versions))
	return versions, nil
}

// parseVersionIndex extracts version strings from HTML containing links in
// the format <a href="/terraform/X.Y.Z/">. Uses golang.org/x/net/html for
// proper HTML parsing.
func parseVersionIndex(r io.Reader) ([]string, error) {
	tokenizer := html.NewTokenizer(r)
	var versions []string

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return versions, nil
			}
			return nil, fmt.Errorf("tokenizing HTML: %w", err)

		case html.StartTagToken:
			tn, hasAttr := tokenizer.TagName()
			if string(tn) != "a" || !hasAttr {
				continue
			}
			version := extractVersionFromAttrs(tokenizer)
			if version != "" {
				versions = append(versions, version)
			}
		}
	}
}

// extractVersionFromAttrs scans the attributes of an <a> tag for an href
// matching /terraform/VERSION/ and returns the VERSION string. Returns ""
// if no matching href is found.
func extractVersionFromAttrs(tokenizer *html.Tokenizer) string {
	for {
		key, val, more := tokenizer.TagAttr()
		if string(key) == "href" {
			if v := versionFromHref(string(val)); v != "" {
				return v
			}
		}
		if !more {
			return ""
		}
	}
}

// versionFromHref extracts a version string from an href like
// "/terraform/1.5.0/" or "terraform/1.5.0-rc1/". Returns "" if the
// href does not match the expected pattern.
func versionFromHref(href string) string {
	// Strip leading slash and split.
	href = strings.TrimPrefix(href, "/")
	href = strings.TrimSuffix(href, "/")
	parts := strings.Split(href, "/")

	if len(parts) != 2 || parts[0] != "terraform" {
		return ""
	}

	candidate := parts[1]
	// Validate it parses as a semver version.
	if _, err := goversion.NewVersion(candidate); err != nil {
		return ""
	}
	return candidate
}

// sortVersions sorts a slice of version strings by semver. If reverse is
// false, sorts newest first (descending). If reverse is true, sorts oldest
// first (ascending).
func sortVersions(versions []string, reverse bool) {
	sort.Slice(versions, func(i, j int) bool {
		vi, _ := goversion.NewVersion(versions[i])
		vj, _ := goversion.NewVersion(versions[j])
		if reverse {
			return vi.LessThan(vj)
		}
		return vj.LessThan(vi)
	})
}
