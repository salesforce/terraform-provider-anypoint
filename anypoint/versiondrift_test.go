package anypoint

import (
	"reflect"
	"testing"
)

func TestClassifyVersionDrift(t *testing.T) {
	cases := []struct {
		name        string
		declared    string
		knownLatest string
		available   []string
		want        VersionDriftResult
	}{
		{
			name:        "patch drift with knownLatest",
			declared:    "1.1.0",
			knownLatest: "1.1.1",
			available:   []string{"1.1.0", "1.0.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.1",
				AvailableVersions: []string{"1.1.0", "1.0.0"},
				DriftSeverity:     DriftSeverityPatch,
				IsOutdated:        true,
			},
		},
		{
			name:        "minor drift",
			declared:    "1.0.5",
			knownLatest: "",
			available:   []string{"1.1.0", "1.0.5", "1.0.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.0",
				AvailableVersions: []string{"1.1.0", "1.0.5", "1.0.0"},
				DriftSeverity:     DriftSeverityMinor,
				IsOutdated:        true,
			},
		},
		{
			name:        "major drift",
			declared:    "1.5.0",
			knownLatest: "",
			available:   []string{"2.0.0", "1.5.0"},
			want: VersionDriftResult{
				LatestVersion:     "2.0.0",
				AvailableVersions: []string{"2.0.0", "1.5.0"},
				DriftSeverity:     DriftSeverityMajor,
				IsOutdated:        true,
			},
		},
		{
			name:        "none — declared equals latest",
			declared:    "1.1.1",
			knownLatest: "1.1.1",
			available:   []string{"1.1.1", "1.1.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.1",
				AvailableVersions: []string{"1.1.1", "1.1.0"},
				DriftSeverity:     DriftSeverityNone,
				IsOutdated:        false,
			},
		},
		{
			name:        "none — declared ahead of latest",
			declared:    "1.2.0",
			knownLatest: "1.1.1",
			available:   []string{"1.1.1", "1.1.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.1",
				AvailableVersions: []string{"1.1.1", "1.1.0"},
				DriftSeverity:     DriftSeverityNone,
				IsOutdated:        false,
			},
		},
		{
			name:        "unknown — declared not semver",
			declared:    "snapshot",
			knownLatest: "1.1.1",
			available:   []string{"1.1.1"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.1",
				AvailableVersions: []string{"1.1.1"},
				DriftSeverity:     DriftSeverityUnknown,
				IsOutdated:        false,
			},
		},
		{
			name:        "unknown — empty declared",
			declared:    "",
			knownLatest: "1.1.1",
			available:   []string{"1.1.1"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.1",
				AvailableVersions: []string{"1.1.1"},
				DriftSeverity:     DriftSeverityUnknown,
				IsOutdated:        false,
			},
		},
		{
			name:        "unknown — no versions known",
			declared:    "1.0.0",
			knownLatest: "",
			available:   []string{},
			want: VersionDriftResult{
				LatestVersion:     "",
				AvailableVersions: []string{},
				DriftSeverity:     DriftSeverityUnknown,
				IsOutdated:        false,
			},
		},
		{
			name:        "non-semver values sorted to tail; semver picked as latest",
			declared:    "1.0.0",
			knownLatest: "",
			available:   []string{"snapshot", "2.0.0", "1.0.0"},
			want: VersionDriftResult{
				LatestVersion:     "2.0.0",
				AvailableVersions: []string{"2.0.0", "1.0.0", "snapshot"},
				DriftSeverity:     DriftSeverityMajor,
				IsOutdated:        true,
			},
		},
		{
			name:        "deduplicates",
			declared:    "1.0.0",
			knownLatest: "",
			available:   []string{"1.1.0", "1.1.0", "1.0.0", "1.0.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.1.0",
				AvailableVersions: []string{"1.1.0", "1.0.0"},
				DriftSeverity:     DriftSeverityMinor,
				IsOutdated:        true,
			},
		},
		{
			name:        "knownLatest overrides semver sort",
			declared:    "1.0.0",
			knownLatest: "1.0.5",
			available:   []string{"1.1.0", "1.0.5", "1.0.0"},
			want: VersionDriftResult{
				LatestVersion:     "1.0.5",
				AvailableVersions: []string{"1.1.0", "1.0.5", "1.0.0"},
				DriftSeverity:     DriftSeverityPatch,
				IsOutdated:        true,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyVersionDrift(tc.declared, tc.knownLatest, tc.available)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ClassifyVersionDrift mismatch\n want: %+v\n  got: %+v", tc.want, got)
			}
		})
	}
}
