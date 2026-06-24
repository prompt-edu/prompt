# Coding Style (all code)

Cross-cutting conventions. Language-specific rules in `../go/` and `../react-typescript/` extend
these and take precedence on conflict (specific overrides general).

- **Reuse first.** Prefer existing shared libraries, components, and helpers over writing custom
  code. Client: `@tumaet/prompt-ui-components`, `@tumaet/prompt-shared-state`, and the `@/` shared
  library. Server: `prompt-sdk`. Check the catalogs in the stack rules before implementing.
- **Match the surrounding code** — naming, file layout, comment density, and idioms of the module
  you are editing.
- **American English** in identifiers, comments, and docs.
- Keep changes minimal and focused; don't leave dead code or commented-out blocks.
- Use American English; no emojis in code or commit/PR text unless already present in a template.
