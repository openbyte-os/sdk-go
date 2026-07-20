package semver

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Version
	}{
		{name: "core version", input: "1.2.3", want: Version{Major: 1, Minor: 2, Patch: 3}},
		{
			name:  "pre-release and build",
			input: "10.20.30-alpha.1+linux.amd64",
			want: Version{
				Major:      10,
				Minor:      20,
				Patch:      30,
				PreRelease: []string{"alpha", "1"},
				Build:      []string{"linux", "amd64"},
			},
		},
		{name: "surrounding whitespace", input: " 1.2.3-rc.1 ", want: Version{Major: 1, Minor: 2, Patch: 3, PreRelease: []string{"rc", "1"}}},
		{name: "missing patch", input: "1.2", want: Version{}},
		{name: "non-numeric core", input: "one.2.3", want: Version{}},
		{name: "component overflow", input: "65536.2.3", want: Version{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}

func TestVersionAccessors(t *testing.T) {
	version := Version{
		Major:      1,
		Minor:      2,
		Patch:      3,
		PreRelease: []string{"rc", "1"},
		Build:      []string{"linux", "amd64"},
	}

	if got := version.GetMajor(); got != 1 {
		t.Errorf("GetMajor() = %d, want 1", got)
	}
	if got := version.GetMinor(); got != 2 {
		t.Errorf("GetMinor() = %d, want 2", got)
	}
	if got := version.GetPatch(); got != 3 {
		t.Errorf("GetPatch() = %d, want 3", got)
	}
	if got := version.String(); got != "1.2.3-rc.1+linux.amd64" {
		t.Errorf("String() = %q, want %q", got, "1.2.3-rc.1+linux.amd64")
	}

	preRelease := version.GetPreRelease()
	preRelease[0] = "changed"
	if version.PreRelease[0] != "rc" {
		t.Error("GetPreRelease returned the Version's backing slice")
	}

	build := version.GetBuild()
	build[0] = "changed"
	if version.Build[0] != "linux" {
		t.Error("GetBuild returned the Version's backing slice")
	}
}
