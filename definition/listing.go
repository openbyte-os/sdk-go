package definition

import "github.com/openbyte-os/sdk-go/app"

// ListingBuilder composes a marketplace listing. Listing exposes all fields.
// Like Builder, configure it before concurrent use; Build returns a shallow value.
type ListingBuilder struct{ Listing app.Listing }

func NewListing(tagline string) *ListingBuilder {
	return &ListingBuilder{Listing: app.Listing{Tagline: tagline}}
}

func (b *ListingBuilder) WithOverview(overview string) *ListingBuilder {
	b.Listing.Overview = overview
	return b
}

func (b *ListingBuilder) WithHighlights(values ...app.ListingHighlight) *ListingBuilder {
	b.Listing.Highlights = append(b.Listing.Highlights, values...)
	return b
}

func (b *ListingBuilder) WithScreenshots(values ...app.ListingScreenshot) *ListingBuilder {
	b.Listing.Screenshots = append(b.Listing.Screenshots, values...)
	return b
}

func (b *ListingBuilder) WithScopes(values ...app.ListingScope) *ListingBuilder {
	b.Listing.Scopes = append(b.Listing.Scopes, values...)
	return b
}

func (b *ListingBuilder) WithLanguages(values ...string) *ListingBuilder {
	b.Listing.Languages = append(b.Listing.Languages, values...)
	return b
}

func (b *ListingBuilder) WithDocsURL(url string) *ListingBuilder {
	b.Listing.DocsURL = url
	return b
}

func (b *ListingBuilder) WithBrand(brand app.ListingBrand) *ListingBuilder {
	b.Listing.Brand = brand
	return b
}

func (b *ListingBuilder) WithReleases(values ...app.ListingRelease) *ListingBuilder {
	b.Listing.Releases = append(b.Listing.Releases, values...)
	return b
}

func (b *ListingBuilder) Build() app.Listing { return b.Listing }
