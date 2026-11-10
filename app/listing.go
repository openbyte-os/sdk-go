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
	// Brand styles the app's marketplace card cover and app page banner.
	Brand ListingBrand `json:"brand,omitzero"`
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

// ListingBrand is the app's look in the marketplace. Every field is optional:
// the marketplace draws a cover from the app's category colour and a pattern
// chosen from its ID when the vendor supplies none.
type ListingBrand struct {
	// Color is the accent as #RRGGBB. It fills the cover and tints the icon tile.
	Color string `json:"color,omitempty"`
	// Pattern is one of ListingPatterns, drawn faintly over the cover colour.
	Pattern ListingPattern `json:"pattern,omitempty"`
	// Cover is a wide image shown behind the card header and as the app page
	// banner, 1200x400 or wider at 3:1. It follows the same URL rules as
	// screenshots, and replaces the colour and pattern where it shows.
	Cover string `json:"cover,omitempty"`
	// Icon is a square image, 256x256 or larger, shown on the app's tile in the
	// marketplace in place of its definition icon. The rest of the UI keeps the
	// definition icon. It follows the same URL rules as screenshots.
	Icon string `json:"icon,omitempty"`
	// Hero is a large banner image, 2400x1000 or wider at 12:5, shown when the
	// app is promoted at the top of the marketplace. Keep the lower left calm:
	// the app's name, summary and icon are laid over it there. It follows the
	// same URL rules as screenshots.
	Hero string `json:"hero,omitempty"`
}

// ListingPattern is a cover texture drawn by the marketplace.
type ListingPattern string

const (
	PatternNone     ListingPattern = "none"
	PatternDots     ListingPattern = "dots"
	PatternGrid     ListingPattern = "grid"
	PatternDiagonal ListingPattern = "diagonal"
	PatternWaves    ListingPattern = "waves"
	PatternRings    ListingPattern = "rings"
	PatternCircuit  ListingPattern = "circuit"
)

// ListingPatterns are the patterns a listing may choose.
var ListingPatterns = []ListingPattern{PatternNone, PatternDots, PatternGrid, PatternDiagonal, PatternWaves, PatternRings, PatternCircuit}

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
