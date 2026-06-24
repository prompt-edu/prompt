# Codex Overrides — PROMPT 2.0

> Codex reads this file **instead of** `AGENTS.md` at the project root. The base guidance still
> applies — read `AGENTS.md` for project overview, structure, commands, and tooling.

Codex auto-discovers skills from `.agents/skills/` but does **not** load `.claude/rules/`. Before
editing files that match a pattern below, read the corresponding rule file(s) and follow them
(language-specific rules override the general `.claude/rules/common/*` rules on conflict):

| When editing | Read |
| --- | --- |
| `servers/**/*.go` | `.claude/rules/go/*.md` (coding-style, patterns, auth-routing, sqlc, prompt-sdk) |
| `clients/**/*.{ts,tsx}` | `.claude/rules/react-typescript/*.md` (coding-style, state-management, shared-libraries) |
| `**/db/migration/**`, `**/db/query/**`, `**/sqlc.yaml` | `.claude/rules/database/migrations.md` |
| `**/webpack.config.mjs` | `.claude/rules/module-federation/remotes.md` |
| `docker-compose*.yml` | `.claude/rules/docker/compose.md` |
| any file | `.claude/rules/common/*.md` (always applies) |

Repeatable procedures are in `.agents/skills/` (e.g. `new-course-phase`, `sqlc-migration`,
`module-federation-remote`, `add-shared-ui-component`, `keycloak-local-setup`).

> Maintenance: this is a hand-maintained pointer because Codex's root override replaces (not merges
> with) `AGENTS.md`. Keep the table in sync with `.claude/rules/`. See `AI_TOOLING_PLAN.md`.
