package app

import "github.com/openbyte-os/sdk-go/translation"

type HelpArticleType string

const (
	HelpArticleTypeArticle         HelpArticleType = "article"
	HelpArticleTypeHowTo           HelpArticleType = "how-to"
	HelpArticleTypeGlossary        HelpArticleType = "glossary"
	HelpArticleTypeFAQ             HelpArticleType = "faq"
	HelpArticleTypeTroubleshooting HelpArticleType = "troubleshooting"
	HelpArticleTypeReference       HelpArticleType = "reference"
)

type HelpArticle struct {
	ID      string           `json:"id"`
	Type    HelpArticleType  `json:"type,omitempty"`
	Title   translation.Text `json:"title"`
	Summary translation.Text `json:"summary,omitempty"`
	Content translation.Text `json:"content,omitempty"` // Main article body; plain text with paragraphs separated by blank lines

	Icon     string   `json:"icon,omitempty"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Keywords []string `json:"keywords,omitempty"`

	DestinationPath string `json:"destinationPath,omitempty"` // App-relative path to open this article
	ExternalURL     string `json:"externalUrl,omitempty"`     // Absolute URL when help lives outside the app

	Priority int       `json:"priority,omitempty"` // Base priority when ranking this article
	Pins     []HelpPin `json:"pins,omitempty"`     // Contextual ranking rules
}

type HelpPin struct {
	App       ScopedKey `json:"app,omitempty"`       // Optional app scope this pin applies to; empty means the owning app
	PathRegex string    `json:"pathRegex,omitempty"` // Regex matched against the active app-relative path
	Priority  int       `json:"priority,omitempty"`  // Additional priority when this pin matches
}
