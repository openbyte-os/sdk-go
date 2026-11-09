package app

// Listing is an app's marketplace listing. An app serves it from PathListing
// so each deploy can update its listing, and the developer centre imports and
// exports the same format as listing.json.
type Listing struct {
	// Tagline is the one-line pitch, up to 140 characters.
	Tagline string `json:"tagline,omitempty"`
	// Overview is the long description; paragraphs are separated by blank lines.
	Overview    string              `json:"overview,omitempty"`
	Highlights  []ListingHighlight  `json:"highlights,omitempty"`
	Screenshots []ListingScreenshot `json:"screenshots,omitempty"`
	// Scopes describe the data the app reads or changes in a workspace.
	Scopes    []ListingScope `json:"scopes,omitempty"`
	Languages []string       `json:"languages,omitempty"`
	DocsURL   string         `json:"docsURL,omitempty"`
	// Releases are notes per version. The marketplace shows a version's notes
	// once the platform has seen that version running.
	Releases []ListingRelease `json:"releases,omitempty"`
}

type ListingHighlight struct {
	// Icon is a Lucide icon name with the @lu suffix, e.g. "zap@lu".
	Icon  string `json:"icon,omitempty"`
	Title string `json:"title"`
	Text  string `json:"text,omitempty"`
}

// ListingScreenshot URL is an https URL, or a path the platform fetches from
// the app (served from PathListing) or from the import zip (listing files).
type ListingScreenshot struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
}

type ListingScope struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

type ListingRelease struct {
	Version string   `json:"version"`
	Notes   []string `json:"notes"`
}
