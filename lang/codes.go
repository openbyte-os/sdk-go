package lang

type Language string

// --- Major Two-Letter (ISO 639-1) Codes ---
const (
	// European/Western Languages
	EN Language = "en" // English
	FR Language = "fr" // French
	DE Language = "de" // German
	ES Language = "es" // Spanish
	IT Language = "it" // Italian
	PT Language = "pt" // Portuguese
	RU Language = "ru" // Russian
	NL Language = "nl" // Dutch
	SV Language = "sv" // Swedish
	NO Language = "no" // Norwegian
	DA Language = "da" // Danish
	FI Language = "fi" // Finnish
	PL Language = "pl" // Polish
	CS Language = "cs" // Czech
	HU Language = "hu" // Hungarian
	TR Language = "tr" // Turkish

	// Asian/Middle Eastern Languages
	ZH Language = "zh" // Chinese (generic)
	JA Language = "ja" // Japanese
	KO Language = "ko" // Korean
	AR Language = "ar" // Arabic
	HE Language = "he" // Hebrew
	HI Language = "hi" // Hindi
	ID Language = "id" // Indonesian
	TH Language = "th" // Thai

	// Other Major Languages
	VI Language = "vi" // Vietnamese
	EL Language = "el" // Greek
)

// --- Common Three-Letter (ISO 639-2) Codes ---
// Used for less common languages or to denote a language family/group.
const (
	AAR Language = "aar" // Afar
	AFR Language = "afr" // Afrikaans
	AMH Language = "amh" // Amharic
	AST Language = "ast" // Asturian
	BER Language = "ber" // Berber
	BOS Language = "bos" // Bosnian
	CAT Language = "cat" // Catalan
	EUS Language = "eus" // Basque
	FAS Language = "fas" // Persian (Farsi)
	GLA Language = "gla" // Gaelic (Scottish)
	IND Language = "ind" // Indonesian (alternative)
	LAT Language = "lat" // Latin
	MAY Language = "may" // Malay (alternative)
	SWA Language = "swa" // Swahili
	TUR Language = "tur" // Turkish (alternative)
	WEL Language = "wel" // Welsh
	YOR Language = "yor" // Yoruba
)

// --- Common Locales (Language-Region BCP 47 Tags) ---
const (
	// English Locales
	EN_US Language = "en-us" // English (United States)
	EN_GB Language = "en-gb" // English (Great Britain)
	EN_CA Language = "en-ca" // English (Canada)
	EN_AU Language = "en-au" // English (Australia)
	EN_IE Language = "en-ie" // English (Ireland)
	EN_NZ Language = "en-nz" // English (New Zealand)

	// Spanish Locales
	ES_ES Language = "es-es" // Spanish (Spain)
	ES_MX Language = "es-mx" // Spanish (Mexico)
	ES_AR Language = "es-ar" // Spanish (Argentina)
	ES_CL Language = "es-cl" // Spanish (Chile)

	// Portuguese Locales
	PT_BR Language = "pt-br" // Portuguese (Brazil)
	PT_PT Language = "pt-pt" // Portuguese (Portugal)

	// French Locales
	FR_FR Language = "fr-fr" // French (France)
	FR_CA Language = "fr-ca" // French (Canada)
	FR_BE Language = "fr-be" // French (Belgium)

	// Chinese Locales (Script and Region)
	ZH_HANS Language = "zh-hans" // Chinese (Simplified Script)
	ZH_HANT Language = "zh-hant" // Chinese (Traditional Script)
	ZH_CN   Language = "zh-cn"   // Chinese (China, Simplified)
	ZH_TW   Language = "zh-tw"   // Chinese (Taiwan, Traditional)
	ZH_HK   Language = "zh-hk"   // Chinese (Hong Kong, Traditional)

	// Arabic Locales (Often grouped by region)
	AR_EG Language = "ar-eg" // Arabic (Egypt)
	AR_SA Language = "ar-sa" // Arabic (Saudi Arabia)

	// Other Locales
	DE_AT Language = "de-at" // German (Austria)
	DE_CH Language = "de-ch" // German (Switzerland)
	NL_BE Language = "nl-be" // Dutch (Belgium) - Flemish
	NB_NO Language = "nb-no" // Norwegian Bokmål (Norway)
)
