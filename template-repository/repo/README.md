# PROMPT Course Phase Template

Template for building a **custom course phase** for [PROMPT](https://github.com/prompt-edu/prompt),
the modular course management platform for project-based university teaching. A course phase is:

- `client/` — a React micro-frontend (Rspack + Module Federation) loaded into PROMPT's core UI
- `server/` — a Go (Gin) microservice with its own PostgreSQL database, secured via
  [prompt-sdk](https://github.com/prompt-edu/prompt-sdk)

Your phase runs in its own deployment; the PROMPT instance loads the client at runtime and your
server talks to the PROMPT core over HTTP. Existing external phases built on this model:
[prompt-intro-course](https://github.com/prompt-edu/prompt-intro-course),
[prompt-github-challenge](https://github.com/prompt-edu/prompt-github-challenge).

## Getting started

1. Click **Use this template** on GitHub and clone your new repository.
2. Initialize it with your phase name:

   ```bash
   ./init.sh my_phase 3011 8089
   ```

   This renames every `example` placeholder (component name, Go module, routes, env vars) and
   verifies the build. Delete `init.sh` afterwards and commit.

3. Run it locally against a PROMPT dev stack (see the
   [PROMPT setup guide](https://prompt-edu.github.io/prompt/contributor/setup)):

   ```bash
   cp .env.template .env      # adjust ports/URLs if needed
   docker compose up          # phase client + server + database
   cd client && yarn install && yarn dev   # or run the client with hot reload
   ```

4. Register your phase in the PROMPT instance — one small PR to
   [prompt-edu/prompt](https://github.com/prompt-edu/prompt) adding the Module Federation remote,
   host env var, phase mappings, and course phase type. Step-by-step:
   [How to Create a New Course Phase](https://prompt-edu.github.io/prompt/contributor/new_course_phase)
   (section "Register the phase in core").

## What the template demonstrates

- **Client**: the three Module Federation exposes PROMPT requires (`routes`, `sidebar`,
  `provide`), role-gated routes, shared UI from `@tumaet/prompt-ui-components`, and the React
  Query + Axios pattern for talking to your server. See `client/README.md`.
- **Server**: prompt-sdk auth middleware (routes under `/api/course_phase/:coursePhaseID`),
  the phase config and copy endpoint contracts, sqlc + golang-migrate database access, and the
  service info endpoint. See `server/README.md`.

## Development

```bash
cd client && yarn dev          # micro-frontend with hot reload
cd server && go run main.go    # phase service (expects Postgres + Keycloak, see .env.template)
cd server && go test ./...
```

## License

MIT — same as PROMPT.
