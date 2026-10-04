# AI service

The AI service is the only component that talks to the model provider (Logos). Phase servers and
clients call it with the user's Keycloak token, so a call on behalf of a user and a direct call are
treated the same. Every call is recorded in this service's own database before the provider sees
it. Design and decision records: `docs/contributor/architecture/ai-integration.md`.

All course data routes are below:

```text
/ai/api/course_phase/:coursePhaseID
```

| Route | Roles | Purpose |
| --- | --- | --- |
| `POST /v1/chat/completions` | PromptAdmin, CourseLecturer, CourseEditor | OpenAI-compatible pass-through, streamed or not |
| `GET /v1/models` | PromptAdmin, CourseLecturer, CourseEditor | Models both allowed here and available to the phase's key |
| `GET`, `PUT`, `DELETE /key` | PromptAdmin, CourseLecturer | The phase's Logos key; `GET` never returns the key itself |
| `POST /calls/:callID/events` | PromptAdmin, CourseLecturer, CourseEditor | Oversight events (`shown`, `accepted`, `edited`, `rejected`) on the caller's own calls |
| `GET /calls`, `GET /calls/:callID` | PromptAdmin | Audit metadata, and one call's content (recorded as `content_viewed`) |

`PromptLecturer` is never allowed: the SDK lets it pass for every course phase, so it could spend
every phase's key. The service also implements the SDK privacy export and deletion contracts and
`phase.deletion` (`DELETE /ai/api/course_phase/:coursePhaseID`, which deletes only the key).

A stock OpenAI SDK works with `baseURL = <host>/ai/api/course_phase/<id>/v1` and the Keycloak token
as API key. Optional request headers, recorded as given: `X-Prompt-Feature`, `X-Prompt-Template`,
`X-Prompt-Template-Version` and `X-Prompt-Subjects` (comma-separated course participation ids). The
response carries `X-Prompt-AI-Call-ID`.

Features and their content retention live in `feature/registry.go`, keyed
`<phase type slug>.<feature>`. A call without a feature header is `adhoc`.

## Configuration

| Variable | Purpose |
| --- | --- |
| `AI_PROVIDER_BASE_URL` | OpenAI-compatible base URL of Logos, including `/v1` |
| `AI_ALLOWED_MODELS` | Comma-separated models Logos hosts locally; nothing else is forwarded |
| `AI_ENCRYPTION_KEY` | Base64-encoded 32-byte AES key for phase keys and stored content |
| `AI_AUDIT_METADATA_RETENTION_DAYS` | Call metadata retention (default 1825), longer than any content retention |
| `DB_HOST_AI`, `DB_PORT_AI` | The service's own database (local default port 5442) |

Run it with `make server-ai` (port 8092), test it with `make test-ai`. The Go tests start Postgres
and `aimock` through testcontainers, with the fixtures shared with e2e under `e2e/fixtures/aimock`.
