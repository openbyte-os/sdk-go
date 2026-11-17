package definition

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/openbyte-os/sdk-go/app"
)

// LoadVersionFile reads a version.json document over explicit defaults (for
// example Environment). Missing or invalid files return an error; no build time
// is invented. Both SDK buildID and the legacy K4 build key are accepted.
func LoadVersionFile(path string, defaults app.Version) (app.Version, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return app.Version{}, fmt.Errorf("definition: read version file: %w", err)
	}
	value := struct {
		app.Version
		Build string `json:"build"`
	}{Version: defaults}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return app.Version{}, fmt.Errorf("definition: decode version file: %w", err)
	}
	if fields == nil {
		return app.Version{}, fmt.Errorf("definition: version file must be a JSON object")
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return app.Version{}, fmt.Errorf("definition: decode version file: %w", err)
	}
	if _, explicit := fields["buildID"]; !explicit && value.Build != "" {
		value.BuildID = value.Build
	}
	return value.Version, nil
}

// WithVersionFile loads the file during Build. Errors are returned by Build.
// WithVersion replaces this source; providers take precedence over both.
func (b *Builder) WithVersionFile(path string, defaults app.Version) *Builder {
	b.version = nil
	b.versionFile = &versionFile{path: path, defaults: defaults}
	return b
}

type versionFile struct {
	path     string
	defaults app.Version
}
