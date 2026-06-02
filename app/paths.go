package app

const PathHealthCheck = "/_kubex/health"    // returns HealthCheckResponse
const PathDefinition = "/_kubex/definition" // returns Definition
const PathBlueprints = "/_kubex/blueprints" // returns BlueprintsResponse
const PathHelp = "/_kubex/help"             // returns HelpArticlesResponse
const PathVersion = "/_kubex/version"       // returns Version

const PathAbout = "/_kubex/about" // Returns html information to show in the about view

const PathConfigure = "/_kubex/configure"
const PathActivate = "/_kubex/activate"

const PathPrefixWebhooks = "/_webhooks/"

// BlueprintsResponse should be returned as json from PathBlueprints.
type BlueprintsResponse struct {
	Blueprints []BlueprintDefinition `json:"blueprints"`
}

// HelpArticlesResponse should be returned as json from PathHelp.
type HelpArticlesResponse struct {
	Articles []HelpArticle `json:"articles"`
}
