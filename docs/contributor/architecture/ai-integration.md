---
sidebar_position: 6
---

# 🤖 AI Integration

PROMPT reaches language models through one service, the **AI server** (`servers/ai`). It is the only
component that talks to the model provider, Logos. Phase servers and clients call it with the
user's Keycloak token, and it records every call in its own database before the provider sees it.

Everything AI is behind one switch, `AI_ENABLED`, which is off by default.

## Component Diagram

![Component diagram of the AI integration](./img/ai-integration/ai-integration.svg)

> The AI server and its database run only with `AI_ENABLED=true`. The Assessment client and server
> stand for any phase that uses AI. The diagram is an [Apollon](https://apollon.aet.cit.tum.de)
> model: import `img/ai-integration/ai-integration.json` there to edit it, then export JSON and SVG
> again.

## Decision Records

1. **Packaging.** A separate service, `servers/ai`, with its own PostgreSQL database. Core stays out
   of the model call path, so a slow or stuck model call cannot starve it.
2. **Exposure and auth.** The AI server is public under `/ai/api/course_phase/:coursePhaseID` and
   uses the unchanged prompt-sdk `AuthenticationMiddleware` (Keycloak JWKS plus core role
   resolution). Calls from a phase server on behalf of a user and direct user calls are treated
   the same. Inference and events: PromptAdmin, CourseLecturer, CourseEditor. Keys: PromptAdmin,
   CourseLecturer. Audit reads: PromptAdmin. No students, and never `PromptLecturer`, which the SDK
   lets pass for every phase.
3. **Contract.** OpenAI-compatible pass-through: `POST /v1/chat/completions` and `GET /v1/models`
   under the course phase path. PROMPT context goes into headers (`X-Prompt-Feature`,
   `X-Prompt-Template`, `X-Prompt-Template-Version`, `X-Prompt-Subjects`); the response carries
   `X-Prompt-AI-Call-ID`. Prompt templates, context aggregation and identifier stripping stay in
   the calling service.
4. **Model access.** Logos only. One Logos key per `course_phase_id`, AES-GCM encrypted and
   write-only; no key means no AI for that phase. `AI_ALLOWED_MODELS` restricts requests to models
   Logos hosts locally, because a Logos key alone does not guarantee local processing. No rate
   limiting or budget logic in the AI server: Logos meters per key.
5. **Streaming.** OpenAI SSE end to end, passed through rather than re-encoded
   (`httputil.ReverseProxy`). Clients use `fetch` plus `eventsource-parser`, because `EventSource`
   cannot send a bearer token. Calls stay inside the user's request; there is no background
   execution.
6. **Audit.** Own database, fail-closed: the record is written before the provider is called.
   Content lives in a separate table with per-feature retention from a feature registry in code
   (`<phase type slug>.<feature>`, prefix enforced against the phase type stored with the key).
   Metadata retention is longer than the longest content retention. Nothing goes to core's
   `audit_log`, and phase servers mark their AI routes with `audit.Skip()`.
7. **Human oversight.** Events API for `shown`, `accepted`, `edited` (edit distance) and
   `rejected`, reported by the user who made the call. Audit reads are admin-only, under the course phase routes, and every content read
   is itself recorded. Admin page per course in core.
8. **Privacy.** Core adds the AI server as an extra target to the privacy export, per-student
   deletion and phase deletion fan-outs when AI is enabled. High-risk content under erasure is
   access-restricted until its retention ends (GDPR Art. 17(3)(b)); other content is deleted
   immediately. Students are informed through the privacy policy (AI Act Art. 26(11));
   lecturer-facing output is labeled AI-generated (Art. 50(1)).
9. **Switch and failure behavior.** One GitHub variable, `AI_ENABLED`, off by default. The
   deployment derives the compose profile `ai` from it (AI server and database), and it is passed
   to core and to the phase servers that use AI. Disabled, unreachable or unconfigured all mean
   the AI surfaces are hidden and the workflow is untouched.
10. **Regulatory stance.** Treated as high-risk under the EU AI Act (Annex III 3(b), steering the
    learning process; the Art. 6(3) exemption does not apply because condensing feedback about one
    student is profiling). Logging per Art. 12, retention of at least six months per Art. 19 and
    26(6) for high-risk features.
11. **Testing.** aimock (`ghcr.io/copilotkit/aimock`) as the stub provider in the Go tests
    (testcontainers) and in the `ai` e2e shard, with shared fixtures under `e2e/fixtures/aimock`.
12. **Evaluation study data.** The study gets only an anonymized aggregate (counts, acceptance and
    edit-distance distributions, latency per feature and template version), produced by a
    documented SQL script run by an admin. Raw audit data never leaves production.

## Calling the AI Server

A stock OpenAI SDK works against the AI server: set `baseURL` to
`<host>/ai/api/course_phase/<coursePhaseID>/v1` and pass the user's Keycloak token as the API key.

| Route | Roles | Purpose |
| --- | --- | --- |
| `POST /v1/chat/completions` | PromptAdmin, CourseLecturer, CourseEditor | Streamed or plain completion |
| `GET /v1/models` | PromptAdmin, CourseLecturer, CourseEditor | Models both allowed and available to the phase's key |
| `GET`, `PUT`, `DELETE /key` | PromptAdmin, CourseLecturer | The phase's Logos key; `GET` returns only `configured`, `last4`, `setBy`, `setAt` |
| `POST /calls/:callID/events` | PromptAdmin, CourseLecturer, CourseEditor | Oversight events, on the caller's own calls (any call for PromptAdmin) |
| `GET /calls`, `GET /calls/:callID` | PromptAdmin | Audit metadata, and one call's content |

The gateway forwards the body unchanged, except that the `model` must be in `AI_ALLOWED_MODELS`,
`user` is removed so no identifier reaches the provider, and streams always end with a usage
chunk. A phase without a key answers `409 AI not configured`. If the call cannot be recorded, the
answer is `503` and the provider is never called. Requests over 4 MiB are refused; a call ends
after 10 minutes, or after 2 minutes without a chunk.

`X-Prompt-Subjects` takes at most 100 course participation ids. They are recorded as given, so the
calling service is responsible for naming the right students.

A feature is named `<phase type slug>.<feature>`, for example `assessment.action_item_suggestions`.
The slug is the phase type name in kebab case. A call without `X-Prompt-Feature` is `adhoc`. A new
feature, or a different retention, is a change to `servers/ai/feature/registry.go`.

| Feature | Content retention | High-risk |
| --- | --- | --- |
| `adhoc` | 183 days | yes |
| `assessment.*` | 730 days | yes |

## Audit Records

| Table | Contents | Retention |
| --- | --- | --- |
| `ai_call` | Who (Keycloak `sub` and role, no name), phase, feature, template, models, parameters, hashes of request and response, outcome, status, finish reason, tokens, timings | `AI_AUDIT_METADATA_RETENTION_DAYS` (default 1825) |
| `ai_call_content` | Request and response as passed through, AES-GCM encrypted; a cancelled stream keeps its partial response | per feature |
| `ai_call_subject` | The course participations named in `X-Prompt-Subjects` | per feature |
| `ai_call_content_restriction` | Content kept but restricted after an erasure request | per feature |
| `ai_call_event` | `shown`, `accepted`, `edited`, `rejected`, `content_viewed` | metadata retention |
| `ai_phase_key` | The encrypted Logos key of a phase, with its phase type | until removed or the phase is deleted |

Triggers keep the audit tables append-only: the only update completes a pending call once. A daily
purge, which also runs at startup, deletes content and subjects per feature and then metadata, and
closes calls left pending by a restart. The AI server refuses to start unless the metadata
retention is longer than the longest content retention.

## Privacy

- **Export:** the calls a person made, and the calls they were a subject of, with their events.
  Content is included only while it is retained and not restricted, and only to the call's sole
  subject: content that also concerns other people is withheld (GDPR Art. 15(4)). The other
  party's identity is left out.
- **Erasure:** applies to the calls the person was a subject of. High-risk content is restricted
  until its retention ends, other content and the person's subject link are deleted at once. The
  metadata of calls the person made stays as logging evidence.
- **Phase and course deletion:** the AI server deletes the phase's key at once; its audit records
  stay until their retention ends.

Core treats the AI server like a phase module in these fan-outs: if it is unreachable, the phase
deletion fails and can be retried, and the privacy request reports the AI part as failed.

## Configuration

| Variable | Purpose |
| --- | --- |
| `AI_ENABLED` | The switch. The Makefile and the deployment derive `COMPOSE_PROFILES=ai` from it |
| `AI_PROVIDER_BASE_URL` | OpenAI-compatible base URL of Logos, including `/v1` |
| `AI_ALLOWED_MODELS` | Comma-separated models Logos hosts locally |
| `AI_ENCRYPTION_KEY` | Base64-encoded 32-byte key, `openssl rand -base64 32` |
| `AI_AUDIT_METADATA_RETENTION_DAYS` | Call metadata retention in days |
| `AI_HOST` | Where the core client reaches the AI server; the core host in production |

In production Traefik routes `/ai/api` to the AI server with the API rate limit and without the
compress middleware, which would buffer streamed completions. Locally, set `AI_ENABLED=true` and
`AI_ALLOWED_MODELS` in `.env`: `make db` then starts the AI database, and `make server-ai` the
server on port 8092. The e2e shard `ai` runs the stack with the AI server and aimock as the
provider.

## Open Items

- A data protection impact assessment (GDPR Art. 35).
- With the data protection officer: what happens to AI data while `AI_ENABLED=false` (the
  database is stopped, so no purge, export or erasure reaches it); restricting instead of deleting
  high-risk content during its retention; a pseudonymized study export, only if the aggregate is
  not enough.
