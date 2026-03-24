package app

type SearchPattern struct {
	RegexMatch string `json:"regexMatch,omitempty"` // A basic regex pattern to match text input
	Path       string `json:"path,omitempty"`       // The endpoint to hit with the search query to fetch results
}
