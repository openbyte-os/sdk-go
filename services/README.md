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
- Request bodies, success bodies, and error bodies use `application/json`. The LLM
  `stream` action responds with `text/event-stream`.
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

## LLM v1

`LLMDefinition()` describes three actions on a provider's base path, for example
`/_services/llm/v1`:

| Action | Endpoint | Body |
| --- | --- | --- |
| `generate` | `POST /generate` | `LLMRequest` → `LLMResponse` |
| `stream` | `POST /stream` | `LLMRequest` → server-sent events ending in an `LLMResponse` |
| `models` | `POST /models` | `{}` → `LLMModelsResponse` |

Every action is a POST to `/<action>`, the form Rubix Gateway forwards service
calls in. `LLM` is the corresponding Go implementation interface. A provider advertises the
actions it implements; a consumer lists the ones it needs in `requiredActions`.
One provider registration covers every model of a vendor account: the workspace
selects the provider instance, and each request selects a model.

Request:

```json
{
  "model": "example-large",
  "system": "You are a concise support assistant.",
  "messages": [
    {"role": "user", "content": [{"type": "text", "text": "Where is order 1042?"}]}
  ],
  "maxOutputTokens": 1024,
  "temperature": 0.2,
  "stopSequences": ["END"]
}
```

Response (200):

```json
{
  "model": "example-large",
  "content": [{"type": "text", "text": "Order 1042 shipped on Tuesday."}],
  "stopReason": "end",
  "usage": {"inputTokens": 21, "outputTokens": 9}
}
```

`messages` alternate between `user` and `assistant`, beginning and ending with
`user`. Every message holds one or more content blocks. To continue a
conversation, append the response's `content` as an `assistant` message, then the
next `user` message. A response whose `content` is empty (possible with
`max_tokens` and `content_filter`) has no assistant turn to append. Providers hold
no conversation state between requests.

| Field | Meaning |
| --- | --- |
| `model` | An ID from `models`. Omitted selects the instance's default model. An unknown ID is 400 `invalid_request`. |
| `system` | Instruction framing the whole conversation. |
| `maxOutputTokens` | Cap on generated tokens. Zero or omitted uses the provider's default. |
| `temperature` | Sampling hint from 0 (most deterministic) to 1 (most varied), scaled to the vendor's range. Ignored by models with fixed sampling. |
| `stopSequences` | Up to 16 strings that end generation. The matched sequence is excluded from the response. |
| `tools`, `toolChoice` | See Tools. |
| `outputSchema` | See Structured output. |

`stopReason` is `end`, `max_tokens` (the content is truncated), `stop_sequence`,
`tool_use`, or `content_filter` (the vendor withheld or cut short the answer).
All five are 200 responses; callers check `stopReason` before using the content.
`usage` reports the vendor's billed token counts for the request.

Message content is data for the model. Providers pass it through and never act
on instructions inside it.

### Content blocks

| `type` | Fields | Allowed in |
| --- | --- | --- |
| `text` | `text`, optional `opaque` | user, assistant |
| `image` | `mediaType` (`image/png`, `image/jpeg`, `image/gif`, `image/webp`), `data` (standard padded base64, at most 5 MiB decoded) | user |
| `tool_use` | `toolUseId`, `name`, `input` (JSON object), optional `opaque` | assistant |
| `tool_result` | `toolUseId`, `text` (the tool's output), `isError` | user |

A response block may carry `opaque`, provider state such as a vendor's reasoning
signature. Return response blocks unchanged, `opaque` included, when sending them
back in `messages`; a provider may be unable to continue a tool exchange without it.

### Tools

```json
{
  "tools": [{
    "name": "get_order",
    "description": "Look up an order by number.",
    "inputSchema": {"type": "object", "properties": {"number": {"type": "integer"}}, "required": ["number"]}
  }],
  "toolChoice": {"type": "auto"}
}
```

Tool names are unique, 1–64 characters of letters, digits, `_`, or `-`.
`toolChoice.type` is `auto` (default), `required` (call at least one tool),
`none`, or `tool` with the `name` of the tool to call. When the model calls
tools, the response has `stopReason` `tool_use` and one `tool_use` block per call.
The caller runs them and answers every one with a `tool_result` block in the next
user message:

```json
{"role": "assistant", "content": [{"type": "tool_use", "toolUseId": "call-1", "name": "get_order", "input": {"number": 1042}}]},
{"role": "user", "content": [{"type": "tool_result", "toolUseId": "call-1", "text": "{\"status\":\"shipped\"}"}]}
```

Providers never execute tools.

### Structured output

`outputSchema` is a JSON Schema object. The response's text is then one JSON
document conforming to it, produced with the vendor's constrained decoding.
Providers confirm the text parses as JSON before returning 200 with `stopReason`
`end`. A `max_tokens` response may hold an incomplete document. Vendors support
differing subsets of JSON Schema; a schema the model cannot honor is 422
`unsupported`.

### Models

```json
{
  "models": [
    {"id": "example-large", "name": "Example Large", "default": true,
     "capabilities": ["tools", "images", "structured_output"],
     "inputTokenLimit": 200000, "outputTokenLimit": 16000},
    {"id": "example-small", "name": "Example Small", "capabilities": ["tools"]}
  ]
}
```

`models` takes an empty JSON object, which providers do not read, and lists what
the selected instance can use; at most one is the `default`.
Token limits are omitted when the vendor does not publish them. A request using a
feature its model lacks (tools, images, or structured output) is 422 `unsupported`.

### Streaming

`stream` takes the same request as `generate`. Send `Accept: text/event-stream`.
A request that fails before generation starts returns the usual JSON error with
a non-2xx status. Otherwise the response is 200 `text/event-stream`:

```text
event: text
data: {"text":"Order 1042 "}

event: text
data: {"text":"shipped on Tuesday."}

event: done
data: {"response":{"model":"example-large","content":[{"type":"text","text":"Order 1042 shipped on Tuesday."}],"stopReason":"end","usage":{"inputTokens":21,"outputTokens":9}}}
```

| Event | Data | Meaning |
| --- | --- | --- |
| `text` | `{"text": "..."}` | The next increment of response text. |
| `done` | `{"response": LLMResponse}` | The complete response, identical to what `generate` returns. Always last. |
| `error` | `{"error": {"code": "...", "message": "..."}}` | Generation failed after the stream began. Always last. |

The concatenated `text` events equal the text blocks of the final response.
Tool calls arrive whole in the `done` response. A stream that closes without
`done` or `error` has failed. Ignore unknown event types and SSE comment lines;
providers may send comments as keep-alives. `WriteLLMStreamEvent` and
`ReadLLMStream` implement this framing.

### Validation and limits

`LLMRequest.Validate` checks the structural rules above and returns
`invalid_request`; providers call it before contacting the vendor, then check the
model and its capabilities. Every provider accepts request bodies up to 16 MiB,
1,024 messages, and 128 tools. Vendor failures map to the common error codes:
429 `rate_limited` for quota, 503 `unavailable` for vendor authentication or
availability failures, and 400 `invalid_request` when the conversation exceeds
the model's context window. Upstream error bodies and credentials are never
returned.
