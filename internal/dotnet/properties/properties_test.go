package properties

import (
	"testing"
)

func TestIsValidFixedVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
		isFixed bool
	}{
		{"fixed version", "10.1.1", true},
		{"fixed beta version", "10.1.1-beta", true},
		{"fixed beta patch", "10.1.1-beta.1", true},
		{"fixed version label zzz", "1.0.1-zzz", true},
		{"fixed version RC with label", "1.0.1-rc.10", true},
		{"fixed version RC with label 2", "1.0.1-rc.2", true},
		{"fixed version with label open", "1.0.1-open", true},
		{"fixed version alpha", "1.0.1-alpha2", true},
		{"fixed version RC with label aaa", "1.0.1-aaa", true},
		{"fixed version range", "[1.0]", true},
		{"version as variable", "$(ComponentDetectionPackageVersion)", true},
		{"version range with inclusive min", "[1.0,)", false},
		{"version range with inclusive min without brackets", "1.0", false},
		{"version range with exclusive min", "(1.0,)", false},
		{"version range with inclusive max", "(,1.0]", false},
		{"version range with exclusive max", "[,1.0)", false},
		{"Exact range, inclusive", "[1.0,2.0]", false},
		{"Exact range, exclusive", "(1.0,2.0)", false},
		{"Mixed inclusive minimum and exclusive maximum version", "(1.0,2.0)", false},
		{"invalid", "(1.0)", false},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			isFixed := isValidFixedVersion(tt.version)
			if tt.isFixed != isFixed {
				t.Errorf("expected %v. Got %v", tt.isFixed, isFixed)
			}
		})
	}
}
