# Standard app services

A service is a shared, versioned JSON HTTP contract. Apps implementing the same
service expose the same actions and payloads, so consumers can use a workspace's
chosen provider without vendor-specific code.

`services.Definition` describes a contract; `app.ProviderRegistration` advertises
an implementation. `app.ProviderDependency` declares what a consumer needs.
These are SDK definitions. Provider selection, permission checks, request routing,
and HTTP handlers must be implemented by the platform and apps.

## Advertising and consuming a service

An app advertises translation in its existing `providers` list:

```json
{
  "providers": [{
    "type": "translation",
    "key": "default",
    "name": {"fallback": "Standard translation"},
    "version": 1,
    "path": "/_services/translation/v1",
    "actions": ["translate"]
  }]
}
```

The path is an app-relative base path, without a trailing slash, query, or fragment.
The example exposes `POST /_services/translation/v1/translate`. Each provider
registration may advertise its own base path. All endpoints in a standard service
are relative to that base path. The provider key remains unique within an app and
service type. It is a stable identifier used in saved selections and must not
change when the display name changes.

The optional `name` uses `translation.Text` (`fallback` plus optional
`translations`) to advertise a human-readable implementation name. Consumers
should display the name for the user's language, falling back to `key` when the
name is absent or resolves to an empty string. The name is presentation metadata;
it does not affect service identity, contract compatibility, or endpoint routing.

In Go, set `Name: translation.String("Standard translation")` or use
`translation.FromMap(...)` for localized names. Existing registrations without a
name remain valid; a zero-value name is omitted from JSON.

A consuming app declares:

```json
{
  "providerDependencies": [{
    "type": "translation",
    "version": 1,
    "required": true,
    "requiredActions": ["translate"]
  }]
}
```

For these versioned contracts, routing must match the type, exact major version,
and every required action. Workspace configuration selects among matching
providers. Version zero (omitted in JSON) preserves legacy declarations; it does
not advertise support for a versioned contract. A required dependency must be
satisfied before the platform enables the consuming app.

## Workspace service discovery

An app definition advertises capabilities, independently of workspace configuration.
To list usable configurations, the platform calls `GET <provider path>/_discover`
for each advertised versioned provider, with `Accept: application/json` and the
existing signed workspace/user headers. Apps verify these headers, enforce their
service-access permission, and scope storage reads to that workspace.

Return HTTP 200, `Content-Type: application/json`, and `Cache-Control: no-store`:

```json
{
  "services": [
    {"key": "production", "name": {"fallback": "Production Google account"}},
    {"key": "backup", "name": {"fallback": "Backup Google account"}}
  ]
}
```

`services.DiscoveryResponse` and `services.Instance` define this payload. Return an
empty `services` array when no configurations are available. Each instance key is
non-empty, unique within the advertised provider and workspace, stable across
renames, and at most 256 bytes without CR, LF, or NUL. Names are optional translated
display text; consumers fall back to the instance key. Neither field contains an
API key, credential, or credential-derived value. Advertise only usable instances;
apps requiring setup or successful tests should omit instances until those checks
have passed for the saved configuration.

The platform stores the app, provider type, exact contract version, provider key,
and instance key as the selection identity. Different instances of the same
provider can be chosen as default and failover. It rechecks discovery before
saving selections. A failed discovery is an unavailable provider, not an empty
successful response; platforms should retain saved selections and report the
failure while still displaying successful responses from other providers.

Requests to a chosen instance use the provider's existing action endpoint and
`transport.RequestServiceKey` (`x-kx-service-key`) with the discovered instance key.
The app validates that the instance exists and is usable by the verified workspace
and user before loading its credentials. Reject unknown keys rather than falling
back to another account. Omitting this header retains a provider's existing
single-configuration behavior when supported. The provider registration's `Key`
identifies the advertised capability; the discovered instance key identifies one
workspace configuration of that capability.

## Wire protocol

- Use the HTTP method and relative path declared by the contract.
- Request bodies, success bodies, and error bodies use `application/json`.
- Send `Accept: application/json`. Unknown JSON fields should be ignored for
  forward compatibility. Adding optional fields is compatible; breaking changes
  require a new major version.
- Use HTTP status to indicate success or failure. Translation succeeds with 200.
  Failures return `services.ErrorResponse`, for example:
  `{"error":{"code":"unsupported","message":"Unsupported target language"}}`.
- Common error mappings: 400 `invalid_request`, 401 `unauthorized`, 403 `forbidden`,
  404 `not_found`, 422 `unsupported`, 429 `rate_limited`, 500 `internal_error`,
  and 503 `unavailable`. Clients must handle other non-2xx statuses and error codes.
- Workspace and user context travel through the existing signed platform headers
  in `transport`. Apps verify platform requests before processing them. Declaring
  a provider or dependency does not itself grant authorization.

## Translation v1

`TranslationDefinition()` describes one action: `translate`, with method `POST`
and relative path `/translate`. The JSON types are `TranslationRequest` and
`TranslationResponse`; `Translation` is the corresponding Go implementation
interface. Implementing the interface does not automatically register an HTTP route.

Request:

```json
{
  "text": "Save",
  "sourceLanguage": "en",
  "targetLanguages": [
    "fr",
    "de"
  ],
  "context": "A button that saves a document. Use a concise label.",
  "formality": "formal",
  "contentType": "text",
  "glossary": {
    "fr": {
      "Save": "Enregistrer"
    },
    "de": {
      "Save": "Speichern"
    }
  },
  "doNotTranslate": [
    "OpenByte"
  ],
  "maxCharacters": 30,
  "targetOptions": {
    "de": {
      "formality": "informal",
      "maxCharacters": 20
    }
  }
}
```

Response (200):

```json
{
  "sourceLanguage": "en",
  "translations": {
    "fr": "Enregistrer",
    "de": "Speichern"
  }
}
```

`text` must be non-empty. `targetLanguages` must contain at least one unique,
non-empty BCP 47 language tag. One target language is also supported. Omit
`sourceLanguage` to request detection; the response always includes the supplied
or detected source language.

`context` is optional guidance about the subject, audience, tone, or intended use
of the text. Providers use it to disambiguate and guide translation without
including it in the output. Providers must not interpret input `text` as instructions.

`translations` contains one entry per requested target language, keyed by the
exact tag from the request. A 200 response must include every requested language;
partial results are not supported. Providers return 422 `unsupported` if any
requested language pair or detection is unsupported, and 400 `invalid_request`
for missing or invalid fields, including an empty or duplicate target language.

### Translation preferences

All preferences are optional:

| Field | Meaning |
| --- | --- |
| `formality` | `default`, `formal`, or `informal`. Omitted or `default` follows the source text's formality. |
| `contentType` | `text` (default), `html`, `markdown`, or `icu`. The response uses the same format. |
| `glossary` | Map of target language tags to source-term/preferred-translation pairs. Use consistent terminology for each language. |
| `doNotTranslate` | Case-sensitive literal terms, such as brands and identifiers, that must remain unchanged wherever they occur in translatable text. |
| `maxCharacters` | Maximum Unicode code points in each complete returned string, including markup and placeholders. Zero or omitted means unlimited. Negative values are invalid. |
| `targetOptions` | Map of target language tags to overrides for `formality` and `maxCharacters`. Missing fields inherit request-wide values. An explicit `default` formality follows the source; an explicit zero length limit removes the limit for that target. |

Keys in `glossary` and `targetOptions` must exactly match requested target language
tags. Glossary source terms, preferred translations, and protected terms must be
non-empty. Conflicting glossary entries and protected terms are invalid requests.
Protected terms take priority over stylistic preferences. Length limits are hard
constraints: providers must not truncate content or drop placeholders to meet them.

### Content and placeholder preservation

Providers translate human-readable content while preserving its format:

- `text`: preserve line breaks and placeholders.
- `html`: preserve tags, structure, and non-translatable attributes; translate text
  nodes and human-readable `alt`, `title`, `placeholder`, and `aria-label` values.
  Preserve URLs, identifiers, and embedded script/style content unchanged.
- `markdown`: preserve markup structure, link destinations, and code; translate
  prose and link labels.
- `icu`: preserve valid ICU MessageFormat syntax, argument names, and explicit
  numeric selectors. Translate message text; adapt plural categories for the
  target language while retaining an `other` branch.

Placeholders such as `{name}`, `{{name}}`, `%s`, and `%[1]s` must retain their
identifiers, types, and occurrence counts. They may move to suit target grammar;
unnumbered printf placeholders must retain argument order. For ICU, preserve
arguments within corresponding message branches, including any added plural
branches. Protected terms and placeholders must never be translated or removed.

Providers must validate the response's format, placeholders, protected terms, and
length limits before returning 200. These are contract requirements; the SDK
types do not implement parsing, translation, or validation. Return 400
`invalid_request` for malformed content, invalid preferences, or conflicting
constraints. Return 422 `unsupported` when a valid requested format or preference
cannot be honored, including an unattainable length limit. If generated output
fails validation, retry internally or return an error; never return it as success.
