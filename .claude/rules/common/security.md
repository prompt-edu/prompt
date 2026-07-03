# Security (all code)

- **Protect every non-public route.** Backend routes that expose course data must sit under
  `/api/course_phase/:coursePhaseID` and use the `prompt-sdk` auth middleware with the correct
  roles. See `../go/auth-routing.md`.
- **Never commit secrets.** Tokens, passwords, and connection strings come from environment
  variables (`.env` / `.env.dev`, both gitignored). Use `${VAR}` interpolation in committed config
  (e.g. `.mcp.json`), never literal secrets.
- Validate and parse all external input (path params, request bodies) before use.
- Roles: `PromptAdmin`, `PromptLecturer`, `CourseLecturer`, `CourseEditor`, `CourseStudent`.
