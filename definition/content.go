package definition

import (
	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/translation"
)

// HelpArticleBuilder composes an article. Article exposes all fields, including
// translations. Build returns a shallow value; configure before concurrent use.
type HelpArticleBuilder struct{ Article app.HelpArticle }

// NewHelpArticle creates a plain-text article with fallback title and content.
func NewHelpArticle(id, title, content string) *HelpArticleBuilder {
	return &HelpArticleBuilder{Article: app.HelpArticle{
		ID: id, Type: app.HelpArticleTypeArticle,
		Title: translation.String(title), Content: translation.String(content),
	}}
}

func (b *HelpArticleBuilder) Configure(fn func(*app.HelpArticle)) *HelpArticleBuilder {
	fn(&b.Article)
	return b
}

func (b *HelpArticleBuilder) WithSummary(summary string) *HelpArticleBuilder {
	b.Article.Summary = translation.String(summary)
	return b
}

func (b *HelpArticleBuilder) WithTags(tags ...string) *HelpArticleBuilder {
	b.Article.Tags = append(b.Article.Tags, tags...)
	return b
}

func (b *HelpArticleBuilder) WithKeywords(keywords ...string) *HelpArticleBuilder {
	b.Article.Keywords = append(b.Article.Keywords, keywords...)
	return b
}

func (b *HelpArticleBuilder) WithPins(pins ...app.HelpPin) *HelpArticleBuilder {
	b.Article.Pins = append(b.Article.Pins, pins...)
	return b
}

func (b *HelpArticleBuilder) Build() app.HelpArticle { return b.Article }

// BlueprintBuilder composes a blueprint. Blueprint exposes all fields.
// Build returns a shallow value; configure before concurrent use.
type BlueprintBuilder struct{ Blueprint app.BlueprintDefinition }

func NewBlueprint(id, name, version string) *BlueprintBuilder {
	return &BlueprintBuilder{Blueprint: app.BlueprintDefinition{
		ID: id, Name: name, Version: version,
		Apps: []app.BlueprintApp{}, Roles: []app.BlueprintRole{}, Integrations: []app.BlueprintIntegration{},
	}}
}

func (b *BlueprintBuilder) Configure(fn func(*app.BlueprintDefinition)) *BlueprintBuilder {
	fn(&b.Blueprint)
	return b
}

func (b *BlueprintBuilder) WithApps(values ...app.BlueprintApp) *BlueprintBuilder {
	b.Blueprint.Apps = append(b.Blueprint.Apps, values...)
	return b
}

func (b *BlueprintBuilder) WithRoles(values ...app.BlueprintRole) *BlueprintBuilder {
	b.Blueprint.Roles = append(b.Blueprint.Roles, values...)
	return b
}

func (b *BlueprintBuilder) WithIntegrations(values ...app.BlueprintIntegration) *BlueprintBuilder {
	b.Blueprint.Integrations = append(b.Blueprint.Integrations, values...)
	return b
}

func (b *BlueprintBuilder) Build() app.BlueprintDefinition { return b.Blueprint }
