package semver

import (
	"strconv"
	"strings"
)

// Major.Minor.Patch[-PreRelease][+Build]
type Version struct {
	Major      uint16
	Minor      uint16
	Patch      uint16
	PreRelease []string
	Build      []string
}

func Parse(input string) Version {
	var version Version

	coreAndPreRelease, build, hasBuild := strings.Cut(strings.TrimSpace(input), "+")
	core, preRelease, hasPreRelease := strings.Cut(coreAndPreRelease, "-")

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}
	}

	major, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return Version{}
	}
	minor, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return Version{}
	}
	patch, err := strconv.ParseUint(parts[2], 10, 16)
	if err != nil {
		return Version{}
	}

	version.Major = uint16(major)
	version.Minor = uint16(minor)
	version.Patch = uint16(patch)
	if hasPreRelease && preRelease != "" {
		version.PreRelease = strings.Split(preRelease, ".")
	}
	if hasBuild && build != "" {
		version.Build = strings.Split(build, ".")
	}

	return version
}

func (v Version) GetMajor() uint16 { return v.Major }

func (v Version) GetMinor() uint16 { return v.Minor }

func (v Version) GetPatch() uint16 { return v.Patch }

func (v Version) GetPreRelease() []string { return append([]string(nil), v.PreRelease...) }

func (v Version) GetBuild() []string { return append([]string(nil), v.Build...) }

func (v Version) String() string {
	var version strings.Builder
	version.WriteString(strconv.FormatUint(uint64(v.Major), 10))
	version.WriteByte('.')
	version.WriteString(strconv.FormatUint(uint64(v.Minor), 10))
	version.WriteByte('.')
	version.WriteString(strconv.FormatUint(uint64(v.Patch), 10))
	if len(v.PreRelease) > 0 {
		version.WriteByte('-')
		version.WriteString(strings.Join(v.PreRelease, "."))
	}
	if len(v.Build) > 0 {
		version.WriteByte('+')
		version.WriteString(strings.Join(v.Build, "."))
	}
	return version.String()
}
