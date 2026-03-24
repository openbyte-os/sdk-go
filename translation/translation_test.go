package translation

import "testing"

func TestFromMap(t *testing.T) {
	m := map[string]string{
		"EN": "English",
		"FR": "French",
	}

	enDefault := FromMap(m, "EN")
	if enDefault.Fallback != "English" {
		t.Error("incorrect fallback for English")
	}

	if enDefault.Get("FR") != "French" {
		t.Error("incorrect French translation for English")
	}

	frDefault := FromMap(m, "FR")
	if frDefault.Fallback != "French" {
		t.Error("incorrect fallback for French")
	}

	esDefault := FromMap(m, "ES")
	if esDefault.Fallback != "" {
		t.Error("incorrect fallback for ES")
	}
	if esDefault.Get("EN") != "English" {
		t.Error("incorrect English translation for ES")
	}
}

func TestString(t *testing.T) {
	txt := String("hello")
	if txt.Fallback != "hello" {
		t.Errorf("expected fallback %q, got %q", "hello", txt.Fallback)
	}
	if txt.Translations != nil {
		t.Error("expected nil translations map")
	}
}

func TestString_Empty(t *testing.T) {
	txt := String("")
	if txt.Fallback != "" {
		t.Errorf("expected empty fallback, got %q", txt.Fallback)
	}
}

func TestText_Get_ReturnsTranslation(t *testing.T) {
	txt := Text{
		Fallback:     "default",
		Translations: map[string]string{"FR": "bonjour", "DE": "hallo"},
	}
	if got := txt.Get("FR"); got != "bonjour" {
		t.Errorf("expected %q, got %q", "bonjour", got)
	}
	if got := txt.Get("DE"); got != "hallo" {
		t.Errorf("expected %q, got %q", "hallo", got)
	}
}

func TestText_Get_ReturnsFallbackWhenLanguageNotFound(t *testing.T) {
	txt := Text{
		Fallback:     "default",
		Translations: map[string]string{"FR": "bonjour"},
	}
	if got := txt.Get("ES"); got != "default" {
		t.Errorf("expected fallback %q, got %q", "default", got)
	}
}

func TestText_Get_NilTranslations(t *testing.T) {
	txt := Text{Fallback: "fallback", Translations: nil}
	if got := txt.Get("EN"); got != "fallback" {
		t.Errorf("expected fallback %q, got %q", "fallback", got)
	}
}

func TestText_Get_EmptyTranslations(t *testing.T) {
	txt := Text{Fallback: "fallback", Translations: map[string]string{}}
	if got := txt.Get("EN"); got != "fallback" {
		t.Errorf("expected fallback %q, got %q", "fallback", got)
	}
}

func TestFromMap_EmptyMap(t *testing.T) {
	txt := FromMap(map[string]string{}, "EN")
	if txt.Fallback != "" {
		t.Errorf("expected empty fallback, got %q", txt.Fallback)
	}
	if got := txt.Get("EN"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFromMap_MissingDefaultLanguageKey(t *testing.T) {
	m := map[string]string{"FR": "bonjour", "DE": "hallo"}
	txt := FromMap(m, "EN")
	if txt.Fallback != "" {
		t.Errorf("expected empty fallback when default language missing, got %q", txt.Fallback)
	}
	if got := txt.Get("FR"); got != "bonjour" {
		t.Errorf("expected %q, got %q", "bonjour", got)
	}
	if got := txt.Get("EN"); got != "" {
		t.Errorf("expected empty string for missing language, got %q", got)
	}
}
