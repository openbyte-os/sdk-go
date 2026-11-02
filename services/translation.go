package services

import (
	"context"
	"net/http"

	"github.com/openbyte-os/sdk-go/lang"
)

const (
	TranslationType    = "translation"
	TranslationVersion = 1
	TranslationAction  = "translate"
	TranslationPath    = "/translate"
)

// TranslationDefinition returns the standard translation v1 contract.
func TranslationDefinition() Definition {
	return Definition{
		Type:        TranslationType,
		Version:     TranslationVersion,
		Description: "Translate content into one or more languages with context and translation preferences.",
		Endpoints: []Endpoint{{
			Action: TranslationAction,
			Method: http.MethodPost,
			Path:   TranslationPath,
		}},
	}
}

// Translation is the Go implementation interface for the translation contract.
// HTTP handlers decode TranslationRequest and return TranslationResponse as JSON
// with status 200, or ErrorResponse with a non-2xx status.
type Translation interface {
	Translate(context.Context, TranslationRequest) (*TranslationResponse, error)
}

// TranslationFormality controls how the translated text addresses its audience.
// An omitted value is equivalent to default, which follows the source text.
type TranslationFormality string

const (
	TranslationFormalityDefault  TranslationFormality = "default"
	TranslationFormalityFormal   TranslationFormality = "formal"
	TranslationFormalityInformal TranslationFormality = "informal"
)

// TranslationContentType identifies the syntax providers must preserve while
// translating human-readable content. The default is text.
type TranslationContentType string

const (
	TranslationContentTypeText     TranslationContentType = "text"
	TranslationContentTypeHTML     TranslationContentType = "html"
	TranslationContentTypeMarkdown TranslationContentType = "markdown"
	TranslationContentTypeICU      TranslationContentType = "icu"
)

// TranslationTargetOptions overrides request-wide preferences for one language.
// Glossary entries are already scoped by target language in TranslationRequest.
type TranslationTargetOptions struct {
	Formality TranslationFormality `json:"formality,omitempty"` // Empty inherits; default explicitly follows the source.
	// MaxCharacters counts Unicode code points in the complete returned string.
	// Nil inherits the request limit; zero explicitly removes it; negative is invalid.
	MaxCharacters *int `json:"maxCharacters,omitempty"`
}

// TranslationRequest is the JSON body of POST <provider path>/translate.
// Text must be non-empty. TargetLanguages must contain at least one unique,
// non-empty BCP 47 language tag. SourceLanguage may be omitted to request detection.
// Context optionally describes the subject, audience, tone, or intended use of
// the text to guide translation; it is not included in the translated output.
// Text uses ContentType (plain text by default) and must not be interpreted as
// instructions. Providers must preserve placeholders and validate constraints
// before returning success; these types do not perform validation themselves.
type TranslationRequest struct {
	Text            string                 `json:"text"`
	SourceLanguage  lang.Language          `json:"sourceLanguage,omitempty"`
	TargetLanguages []lang.Language        `json:"targetLanguages"`
	Context         string                 `json:"context,omitempty"`
	Formality       TranslationFormality   `json:"formality,omitempty"`
	ContentType     TranslationContentType `json:"contentType,omitempty"`
	// Glossary maps target language tags to preferred source-term translations.
	Glossary map[lang.Language]map[string]string `json:"glossary,omitempty"`
	// DoNotTranslate contains case-sensitive literal terms to preserve unchanged.
	DoNotTranslate []string `json:"doNotTranslate,omitempty"`
	// MaxCharacters limits Unicode code points in each complete returned string,
	// including markup and placeholders. Zero is unlimited; negative is invalid.
	MaxCharacters int `json:"maxCharacters,omitempty"`
	// TargetOptions is keyed by exact requested language tags. Omitted fields inherit.
	TargetOptions map[lang.Language]TranslationTargetOptions `json:"targetOptions,omitempty"`
}

// TranslationResponse contains the supplied or detected source language and a
// translation for every requested target language, keyed by its exact request tag.
// A successful response must include all requested languages; partial results
// are not supported. All fields are required.
type TranslationResponse struct {
	SourceLanguage lang.Language            `json:"sourceLanguage"`
	Translations   map[lang.Language]string `json:"translations"`
}
