package definition

import (
	"net/http"

	"github.com/openbyte-os/sdk-go/app"
)

// BlueprintProvider and HelpArticleProvider match K4's existing interfaces.
// Providers run per request and must be safe for concurrent calls.
type BlueprintProvider interface {
	AvailableBlueprints() []*app.BlueprintDefinition
}
type HelpArticleProvider interface {
	HelpArticles() app.HelpArticlesResponse
}
type ListingProvider interface{ Listing() app.Listing }
type VersionProvider interface{ Version() app.Version }
type HealthCheckProvider interface {
	HealthCheck() app.HealthCheckResponse
}

type endpoint func(*http.Request) (any, int)

func (b *Builder) dynamic(path string, fn endpoint) {
	if b.endpoints == nil {
		b.endpoints = make(map[string]endpoint)
	}
	b.endpoints[path] = fn
}

// WithMetadataProviders discovers supported provider interfaces on each value. Unsupported
// interfaces are ignored. The last provider for an endpoint wins, overriding
// static metadata for that endpoint. Providers remain live after Build.
func (b *Builder) WithMetadataProviders(providers ...any) *Builder {
	for _, provider := range providers {
		if p, ok := provider.(BlueprintProvider); ok {
			b.Definition.ProvideBlueprints = true
			b.dynamic(app.PathBlueprints, func(*http.Request) (any, int) {
				result := app.BlueprintsResponse{Blueprints: []app.BlueprintDefinition{}}
				for _, blueprint := range p.AvailableBlueprints() {
					if blueprint != nil {
						result.Blueprints = append(result.Blueprints, *blueprint)
					}
				}
				return result, http.StatusOK
			})
		}
		if p, ok := provider.(HelpArticleProvider); ok {
			b.Definition.ProvideHelp = true
			b.dynamic(app.PathHelp, func(*http.Request) (any, int) {
				result := p.HelpArticles()
				if result.Articles == nil {
					result.Articles = []app.HelpArticle{}
				}
				return result, http.StatusOK
			})
		}
		if p, ok := provider.(ListingProvider); ok {
			b.dynamic(app.PathListing, func(*http.Request) (any, int) { return p.Listing(), http.StatusOK })
		}
		if p, ok := provider.(VersionProvider); ok {
			b.dynamic(app.PathVersion, func(*http.Request) (any, int) { return p.Version(), http.StatusOK })
		}
		if p, ok := provider.(HealthCheckProvider); ok {
			b.WithHealthCheck(p.HealthCheck)
		}
	}
	return b
}

// WithHealthCheck enables a live health endpoint. Unhealthy responses use 503;
// healthy responses use 200. The handler supplies its current definition hash.
// The callback must be non-nil and safe for concurrent calls. Service lifecycle
// and dependency checks remain with the application.
func (b *Builder) WithHealthCheck(check func() app.HealthCheckResponse) *Builder {
	if check == nil {
		panic("definition: nil health check")
	}
	b.dynamic(app.PathHealthCheck, func(*http.Request) (any, int) {
		health := check()
		status := http.StatusOK
		if !health.IsHealthy {
			status = http.StatusServiceUnavailable
		}
		return health, status
	})
	return b
}
