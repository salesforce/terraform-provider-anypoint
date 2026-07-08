package anypoint

import (
	"sort"

	version "github.com/hashicorp/go-version"
)

// DriftSeverity classifies the magnitude of a version drift between a declared
// version and the latest known version. Values mirror SemVer bumps; "unknown"
// covers the case where either input cannot be parsed as SemVer and "none" covers
// declared >= latest (declared is not behind).
const (
	DriftSeverityNone    = "none"
	DriftSeverityPatch   = "patch"
	DriftSeverityMinor   = "minor"
	DriftSeverityMajor   = "major"
	DriftSeverityUnknown = "unknown"
)

// VersionDriftResult captures the classification of a single (declared, latest)
// pair against the full available version set.
type VersionDriftResult struct {
	LatestVersion     string
	AvailableVersions []string // sorted descending; non-SemVer values sorted lexicographically at the tail
	DriftSeverity     string
	IsOutdated        bool
}

// ClassifyVersionDrift derives the latest version from the available set and
// classifies the drift between it and the declared version. The available set
// is normalized (deduplicated + sorted desc) regardless of caller order. When
// the caller already knows the platform-reported "latest" (e.g. Exchange returns
// it explicitly) it should be passed via knownLatest so the result reflects the
// platform's choice rather than our SemVer sort.
//
// Returns:
//   - LatestVersion: knownLatest if non-empty, otherwise the highest SemVer in
//     available (ties broken lexicographically); empty string if available is empty.
//   - AvailableVersions: deduped + sorted desc. Non-SemVer values are placed at
//     the tail, sorted lexicographically among themselves.
//   - DriftSeverity: one of the DriftSeverity* constants. "unknown" if either
//     declared or latest is not SemVer-parseable, "none" if declared >= latest.
//   - IsOutdated: declared < latest (under SemVer). False on "unknown" and "none".
func ClassifyVersionDrift(declared, knownLatest string, available []string) VersionDriftResult {
	sorted, semverHead := normalizeVersions(available)

	latest := knownLatest
	if latest == "" && len(semverHead) > 0 {
		latest = semverHead[0]
	} else if latest == "" && len(sorted) > 0 {
		latest = sorted[0]
	}

	result := VersionDriftResult{
		LatestVersion:     latest,
		AvailableVersions: sorted,
	}

	if declared == "" || latest == "" {
		result.DriftSeverity = DriftSeverityUnknown
		return result
	}

	dv, dErr := version.NewVersion(declared)
	lv, lErr := version.NewVersion(latest)
	if dErr != nil || lErr != nil {
		result.DriftSeverity = DriftSeverityUnknown
		return result
	}

	if dv.GreaterThanOrEqual(lv) {
		result.DriftSeverity = DriftSeverityNone
		return result
	}

	result.IsOutdated = true
	result.DriftSeverity = classifySegmentDelta(dv.Segments(), lv.Segments())
	return result
}

// classifySegmentDelta walks the segment slices (major, minor, patch, ...) and
// returns the severity tag of the highest-significance bump.
func classifySegmentDelta(declaredSegs, latestSegs []int) string {
	max := len(declaredSegs)
	if len(latestSegs) > max {
		max = len(latestSegs)
	}
	for i := 0; i < max; i++ {
		d := segAt(declaredSegs, i)
		l := segAt(latestSegs, i)
		if d == l {
			continue
		}
		switch i {
		case 0:
			return DriftSeverityMajor
		case 1:
			return DriftSeverityMinor
		default:
			return DriftSeverityPatch
		}
	}
	return DriftSeverityNone
}

func segAt(s []int, i int) int {
	if i >= len(s) {
		return 0
	}
	return s[i]
}

// normalizeVersions deduplicates the input, splits SemVer-parseable values from
// non-parseable ones, sorts the SemVer head descending and the non-SemVer tail
// lexicographically descending, and returns (full sorted slice, semver-only slice).
// The semver-only slice is useful when the caller wants to pick a "latest" while
// ignoring junk values (e.g. "v1", "snapshot").
func normalizeVersions(in []string) ([]string, []string) {
	seen := make(map[string]struct{}, len(in))
	semvers := make([]*version.Version, 0, len(in))
	semverStrings := make([]string, 0, len(in))
	others := make([]string, 0, len(in))
	for _, raw := range in {
		if raw == "" {
			continue
		}
		if _, dup := seen[raw]; dup {
			continue
		}
		seen[raw] = struct{}{}
		if v, err := version.NewVersion(raw); err == nil {
			semvers = append(semvers, v)
			semverStrings = append(semverStrings, raw)
		} else {
			others = append(others, raw)
		}
	}
	sort.SliceStable(semvers, func(i, j int) bool { return semvers[i].GreaterThan(semvers[j]) })
	sortedSemvers := make([]string, len(semvers))
	for i, v := range semvers {
		sortedSemvers[i] = v.Original()
	}
	sort.Sort(sort.Reverse(sort.StringSlice(others)))
	full := make([]string, 0, len(sortedSemvers)+len(others))
	full = append(full, sortedSemvers...)
	full = append(full, others...)
	_ = semverStrings // retained for future use; semvers slice is canonical
	return full, sortedSemvers
}
