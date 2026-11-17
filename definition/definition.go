// Package definition builds app metadata and serves the standard /_kubex endpoints.
package definition

import (
	"encoding/json"
	"fmt"

	"github.com/openbyte-os/sdk-go/app"
)

// Builder composes metadata before Build takes a snapshot. Configure builders
// before serving requests; builders themselves are not safe for concurrent use.
// Definition exposes every SDK field, including fields without a fluent helper.
type Builder struct {
	Definition  app.Definition
	listing     *app.Listing
	blueprints  *app.BlueprintsResponse
	help        *app.HelpArticlesResponse
	version     *app.Version
	versionFile *versionFile
	endpoints   map[string]endpoint
}

// New starts with an existing definition, which may also be an empty struct.
// Input maps and slices may be shared until Build takes its snapshot.
func New(def app.Definition) *Builder { return &Builder{Definition: def} }

// Configure edits any definition fields and returns the builder for chaining.
func (b *Builder) Configure(fn func(*app.Definition)) *Builder {
	fn(&b.Definition)
	return b
}

// WithListing sets or replaces the marketplace listing.
func (b *Builder) WithListing(listing app.Listing) *Builder {
	b.listing = &listing
	return b
}

// WithHelp appends articles and enables the help endpoint, even with no articles.
func (b *Builder) WithHelp(articles ...app.HelpArticle) *Builder {
	if b.help == nil {
		b.help = &app.HelpArticlesResponse{Articles: []app.HelpArticle{}}
	}
	b.help.Articles = append(b.help.Articles, articles...)
	b.Definition.ProvideHelp = true
	return b
}

// WithBlueprints appends blueprints and enables the endpoint, even when empty.
func (b *Builder) WithBlueprints(blueprints ...app.BlueprintDefinition) *Builder {
	if b.blueprints == nil {
		b.blueprints = &app.BlueprintsResponse{Blueprints: []app.BlueprintDefinition{}}
	}
	b.blueprints.Blueprints = append(b.blueprints.Blueprints, blueprints...)
	b.Definition.ProvideBlueprints = true
	return b
}

// WithVersion sets or replaces the version endpoint.
func (b *Builder) WithVersion(version app.Version) *Builder {
	b.versionFile = nil
	b.version = &version
	return b
}

// Build encodes static metadata for concurrent HTTP use and captures provider callbacks. Subsequent edits
// to static metadata do not affect the handler; providers remain live. The definition hash is
// recomputed, ignoring any stale supplied hash. Other metadata is not in that hash.
// Build checks JSON encoding; it does not validate platform business rules.
func (b *Builder) Build() (*Handler, error) {
	def := b.Definition
	if b.help != nil || b.endpoints[app.PathHelp] != nil {
		def.ProvideHelp = true
	}
	if b.blueprints != nil || b.endpoints[app.PathBlueprints] != nil {
		def.ProvideBlueprints = true
	}
	def.Hash = ""
	def.GetHash(true)
	h := &Handler{documents: make(map[string][]byte), hash: def.Hash, endpoints: make(map[string]endpoint)}
	for path, fn := range b.endpoints {
		h.endpoints[path] = fn
	}
	add := func(path string, value any) error {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("definition: encode %s: %w", path, err)
		}
		h.documents[path] = data
		return nil
	}
	if err := add(app.PathDefinition, def); err != nil {
		return nil, err
	}
	if b.listing != nil {
		if err := add(app.PathListing, b.listing); err != nil {
			return nil, err
		}
	}
	if b.help != nil {
		if err := add(app.PathHelp, b.help); err != nil {
			return nil, err
		}
	}
	if b.blueprints != nil {
		if err := add(app.PathBlueprints, b.blueprints); err != nil {
			return nil, err
		}
	}
	version := b.version
	if b.versionFile != nil && b.endpoints[app.PathVersion] == nil {
		loaded, err := LoadVersionFile(b.versionFile.path, b.versionFile.defaults)
		if err != nil {
			return nil, err
		}
		version = &loaded
	}
	if version != nil {
		if err := add(app.PathVersion, version); err != nil {
			return nil, err
		}
	}
	return h, nil
}
