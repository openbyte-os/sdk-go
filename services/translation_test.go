package services_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/openbyte-os/sdk-go/app"
	"github.com/openbyte-os/sdk-go/lang"
	"github.com/openbyte-os/sdk-go/services"
)

func TestTranslationWireContract(t *testing.T) {
	request := services.TranslationRequest{Text: "Hello", TargetLanguages: []lang.Language{lang.FR, lang.DE}}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"text":"Hello","targetLanguages":["fr","de"]}` {
		t.Fatalf("request JSON: %s", data)
	}
	var decoded services.TranslationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, request) {
		t.Fatalf("request did not round trip: %+v", decoded)
	}
	response := services.TranslationResponse{
		SourceLanguage: lang.EN,
		Translations:   map[lang.Language]string{lang.FR: "Bonjour", lang.DE: "Hallo"},
	}
	data, err = json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"sourceLanguage":"en","translations":{"de":"Hallo","fr":"Bonjour"}}` {
		t.Fatalf("response JSON: %s", data)
	}
	var result services.TranslationResponse
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, response) {
		t.Fatalf("response did not round trip: %+v", result)
	}
}

func TestTranslationWithContext(t *testing.T) {
	request := services.TranslationRequest{
		Text: "Save", SourceLanguage: lang.EN,
		TargetLanguages: []lang.Language{lang.FR},
		Context:         "A button that saves a document. Use a concise label.",
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"text":"Save","sourceLanguage":"en","targetLanguages":["fr"],"context":"A button that saves a document. Use a concise label."}` {
		t.Fatalf("request JSON: %s", data)
	}
	var decoded services.TranslationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, request) {
		t.Fatalf("contextual request did not round trip: %+v", decoded)
	}
}

func TestTranslationPreferencesWireContract(t *testing.T) {
	unlimited := 0
	request := services.TranslationRequest{
		Text: "Save {name} in OpenByte", SourceLanguage: lang.EN,
		TargetLanguages: []lang.Language{lang.FR, lang.DE},
		Context:         "Document editor", Formality: services.TranslationFormalityFormal,
		ContentType: services.TranslationContentTypeText,
		Glossary: map[lang.Language]map[string]string{
			lang.FR: {"Save": "Enregistrer"}, lang.DE: {"Save": "Speichern"},
		},
		DoNotTranslate: []string{"OpenByte"}, MaxCharacters: 40,
		TargetOptions: map[lang.Language]services.TranslationTargetOptions{
			lang.DE: {Formality: services.TranslationFormalityDefault, MaxCharacters: &unlimited},
			lang.FR: {Formality: services.TranslationFormalityInformal},
		},
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"text":"Save {name} in OpenByte","sourceLanguage":"en","targetLanguages":["fr","de"],"context":"Document editor","formality":"formal","contentType":"text","glossary":{"de":{"Save":"Speichern"},"fr":{"Save":"Enregistrer"}},"doNotTranslate":["OpenByte"],"maxCharacters":40,"targetOptions":{"de":{"formality":"default","maxCharacters":0},"fr":{"formality":"informal"}}}`
	if string(data) != want {
		t.Fatalf("preferences JSON: %s", data)
	}
	var decoded services.TranslationRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, request) {
		t.Fatalf("preferences did not round trip: %+v", decoded)
	}
}

func TestTranslationDefinition(t *testing.T) {
	def := services.TranslationDefinition()
	if def.Type != "translation" || def.Version != 1 || len(def.Endpoints) != 1 {
		t.Fatalf("unexpected definition: %+v", def)
	}
	endpoint := def.Endpoints[0]
	if endpoint.Action != "translate" || endpoint.Method != http.MethodPost || endpoint.Path != "/translate" {
		t.Fatalf("unexpected endpoint: %+v", endpoint)
	}
	def.Endpoints[0].Path = "/changed"
	if services.TranslationDefinition().Endpoints[0].Path != "/translate" {
		t.Fatal("caller mutated the shared contract")
	}
}

func TestServiceErrorWireContract(t *testing.T) {
	response := services.ErrorResponse{Error: services.Error{
		Code: services.ErrorCodeUnsupported, Message: "Unsupported target language",
	}}
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"error":{"code":"unsupported","message":"Unsupported target language"}}` {
		t.Fatalf("error JSON: %s", data)
	}
}

func TestServiceDeclarationsRoundTripAndHash(t *testing.T) {
	def := app.Definition{
		Providers: []app.ProviderRegistration{{
			Type: services.TranslationType, Key: "default", Version: services.TranslationVersion,
			Path: "/_services/translation/v1", Actions: []string{services.TranslationAction},
		}},
		ProviderDependencies: []app.ProviderDependency{{
			Type: services.TranslationType, Version: services.TranslationVersion,
			Required: true, RequiredActions: []string{services.TranslationAction},
		}},
	}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := app.FromJson(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(def.Providers, decoded.Providers) || !reflect.DeepEqual(def.ProviderDependencies, decoded.ProviderDependencies) {
		t.Fatal("service declarations did not round trip")
	}
	initial := def.GetHash(false)
	def.Providers[0].Version++
	if initial == def.GetHash(false) {
		t.Fatal("provider version did not affect definition hash")
	}
	def.Providers[0].Version--
	def.Providers[0].Path += "/other"
	if initial == def.GetHash(false) {
		t.Fatal("provider path did not affect definition hash")
	}
	def.Providers[0].Path = decoded.Providers[0].Path
	def.ProviderDependencies[0].Version++
	if initial == def.GetHash(false) {
		t.Fatal("dependency version did not affect definition hash")
	}
}

func TestLegacyProviderJSONUnchanged(t *testing.T) {
	data, err := json.Marshal(app.ProviderRegistration{Type: "ai", Key: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"type":"ai","key":"default"}` {
		t.Fatalf("legacy provider JSON changed: %s", data)
	}
	data, err = json.Marshal(app.ProviderDependency{Type: "ai", Required: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"type":"ai","required":true}` {
		t.Fatalf("legacy dependency JSON changed: %s", data)
	}
}
