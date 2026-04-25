// Version resolution engine — resolves version specifiers (latest,
// latest:<regex>, latest-allowed, min-required, exact) into concrete
// Terraform version numbers.
package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/tfutils/tfenv/go/internal/config"
	"github.com/tfutils/tfenv/go/internal/logging"
)

// ResolvedVersion holds the final resolved version and metadata about
// where the specifier originated.
type ResolvedVersion struct {
	Version string // Concrete version (e.g. "1.5.7")
	Source  string // Where the specifier came from
}

// exactVersionRe matches a version string that starts with digits in
// semver-ish form: MAJOR.MINOR.PATCH with optional pre-release suffix.
var exactVersionRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// requiredVersionRe extracts the constraint string from a
// required_version line in an HCL or JSON file.  Matches both forms:
//
//	required_version = ">= 1.0.0"   (HCL)
//	"required_version": ">= 1.3.0"  (JSON)
var requiredVersionRe = regexp.MustCompile(
	`"?required_version"?\s*[:=]\s*"([^"]+)"`,
)

// ResolveVersion takes a version specifier and the list of available
// versions (from remote or local listing) and resolves it to a single
// concrete version.
//
// Specifiers:
//
//   - "1.5.0"           — exact, returned as-is
//   - "latest"          — newest stable (no pre-release)
//   - "latest:<regex>"  — newest matching RE2 regex (pre-releases included)
//   - "latest-allowed"  — newest satisfying required_version in .tf files
//   - "min-required"    — minimum satisfying required_version in .tf files
func ResolveVersion(
	specifier string,
	availableVersions []string,
	source string,
	cfg *config.Config,
) (*ResolvedVersion, error) {
	// Strip v prefix.
	specifier = strings.TrimPrefix(specifier, "v")

	logging.Debug("resolving version specifier",
		"specifier", specifier, "candidates", len(availableVersions))

	switch {
	case specifier == "latest":
		return resolveLatest(availableVersions, source)

	case strings.HasPrefix(specifier, "latest:"):
		pattern := strings.TrimPrefix(specifier, "latest:")
		return resolveLatestRegex(pattern, availableVersions, source)

	case specifier == "latest-allowed":
		return resolveLatestAllowed(availableVersions, source, cfg)

	case specifier == "min-required":
		return resolveMinRequired(availableVersions, source, cfg)

	case exactVersionRe.MatchString(specifier):
		return &ResolvedVersion{
			Version: specifier,
			Source:  source,
		}, nil

	default:
		return nil, fmt.Errorf(
			"unrecognised version specifier: %q", specifier)
	}
}

// resolveLatest returns the newest stable version (excluding
// pre-releases: alpha, beta, rc).
func resolveLatest(
	versions []string, source string,
) (*ResolvedVersion, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available")
	}

	stable := filterStable(versions)
	if len(stable) == 0 {
		return nil, fmt.Errorf(
			"no stable versions found among %d candidates",
			len(versions))
	}

	newest := newestVersion(stable)
	logging.Debug("resolved latest", "version", newest)

	return &ResolvedVersion{Version: newest, Source: source}, nil
}

// resolveLatestRegex returns the newest version matching the given RE2
// regex.  Pre-release versions are included if they match the pattern.
func resolveLatestRegex(
	pattern string, versions []string, source string,
) (*ResolvedVersion, error) {
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex %q: %w", pattern, err)
	}

	var matched []string
	for _, v := range versions {
		if re.MatchString(v) {
			matched = append(matched, v)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf(
			"no version matches regex %q among %d candidates",
			pattern, len(versions))
	}

	newest := newestVersion(matched)
	logging.Debug("resolved latest:<regex>",
		"pattern", pattern, "version", newest)

	return &ResolvedVersion{Version: newest, Source: source}, nil
}

// resolveLatestAllowed parses required_version from .tf files and
// returns the latest available version satisfying the constraint.
func resolveLatestAllowed(
	versions []string, source string, cfg *config.Config,
) (*ResolvedVersion, error) {
	constraintStr, err := findRequiredVersion(cfg)
	if err != nil {
		return nil, fmt.Errorf("latest-allowed: %w", err)
	}

	logging.Debug("parsed required_version constraint",
		"constraint", constraintStr)

	constraint, err := goversion.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing version constraint %q: %w", constraintStr, err)
	}

	stable := filterStable(versions)
	matched := filterByConstraint(stable, constraint)

	if len(matched) == 0 {
		return nil, fmt.Errorf(
			"no stable version satisfies constraint %q "+
				"among %d candidates",
			constraintStr, len(versions))
	}

	newest := newestVersion(matched)
	logging.Debug("resolved latest-allowed",
		"constraint", constraintStr, "version", newest)

	return &ResolvedVersion{Version: newest, Source: source}, nil
}

// resolveMinRequired parses required_version from .tf files and returns
// the minimum available version satisfying the constraint.
func resolveMinRequired(
	versions []string, source string, cfg *config.Config,
) (*ResolvedVersion, error) {
	constraintStr, err := findRequiredVersion(cfg)
	if err != nil {
		return nil, fmt.Errorf("min-required: %w", err)
	}

	logging.Debug("parsed required_version constraint",
		"constraint", constraintStr)

	constraint, err := goversion.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing version constraint %q: %w", constraintStr, err)
	}

	stable := filterStable(versions)
	matched := filterByConstraint(stable, constraint)

	if len(matched) == 0 {
		return nil, fmt.Errorf(
			"no stable version satisfies constraint %q "+
				"among %d candidates",
			constraintStr, len(versions))
	}

	oldest := oldestVersion(matched)
	logging.Debug("resolved min-required",
		"constraint", constraintStr, "version", oldest)

	return &ResolvedVersion{Version: oldest, Source: source}, nil
}

// -----------------------------------------------------------------------
// HCL constraint discovery
// -----------------------------------------------------------------------

// findRequiredVersion searches .tf and .tf.json files in the working
// directory for required_version attributes.  It collects all constraints
// and joins them with commas so go-version can intersect them.
func findRequiredVersion(cfg *config.Config) (string, error) {
	dir := cfg.Dir
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("determining working directory: %w", err)
		}
		dir = wd
	}

	logging.Debug("searching for required_version", "dir", dir)

	patterns := []string{
		filepath.Join(dir, "*.tf"),
		filepath.Join(dir, "*.tf.json"),
	}

	var constraints []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("globbing %q: %w", pattern, err)
		}

		for _, path := range matches {
			found, err := extractConstraints(path)
			if err != nil {
				logging.Debug("error reading tf file",
					"path", path, "err", err)
				continue
			}
			constraints = append(constraints, found...)
		}
	}

	if len(constraints) == 0 {
		return "", fmt.Errorf(
			"no required_version constraint found in %s", dir)
	}

	// Join all found constraints with commas for intersection.
	combined := strings.Join(constraints, ", ")
	logging.Debug("combined constraints", "result", combined)

	return combined, nil
}

// extractConstraints reads a single file and returns all
// required_version constraint strings found.
func extractConstraints(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	content := string(data)
	var constraints []string

	for _, line := range strings.Split(content, "\n") {
		// Skip comment lines.
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, "//") {
			continue
		}

		matches := requiredVersionRe.FindStringSubmatch(line)
		if len(matches) >= 2 {
			constraint := strings.TrimSpace(matches[1])
			if constraint != "" {
				constraints = append(constraints, constraint)
			}
		}
	}

	return constraints, nil
}

// -----------------------------------------------------------------------
// Version filtering and sorting helpers
// -----------------------------------------------------------------------

// filterStable returns only versions with no pre-release segment.
func filterStable(versions []string) []string {
	var stable []string
	for _, v := range versions {
		parsed, err := goversion.NewVersion(v)
		if err != nil {
			continue
		}
		if parsed.Prerelease() == "" {
			stable = append(stable, v)
		}
	}
	return stable
}

// filterByConstraint returns versions satisfying the given constraint.
func filterByConstraint(
	versions []string, c goversion.Constraints,
) []string {
	var matched []string
	for _, v := range versions {
		parsed, err := goversion.NewVersion(v)
		if err != nil {
			continue
		}
		if c.Check(parsed) {
			matched = append(matched, v)
		}
	}
	return matched
}

// newestVersion returns the semver-highest version from a non-empty
// slice.
func newestVersion(versions []string) string {
	sorted := sortedVersions(versions)
	return sorted[len(sorted)-1]
}

// oldestVersion returns the semver-lowest version from a non-empty
// slice.
func oldestVersion(versions []string) string {
	sorted := sortedVersions(versions)
	return sorted[0]
}

// sortedVersions returns a copy sorted ascending by semver.
func sortedVersions(versions []string) []string {
	cp := make([]string, len(versions))
	copy(cp, versions)

	sort.Slice(cp, func(i, j int) bool {
		vi, _ := goversion.NewVersion(cp[i])
		vj, _ := goversion.NewVersion(cp[j])
		return vi.LessThan(vj)
	})

	return cp
}
