# Shared client scaffolding

Module Federation scaffolding shared by every phase remote in this monorepo. Not a yarn workspace:
the files are consumed through relative imports so they compile inside each component's own rspack
and TypeScript setup.

- `rspack/federatedDependencies.mjs` - the singleton share scope. The host (`clients/core`) and every
  remote import it, so their share scopes cannot drift apart.
- `rspack/createRspackConfig.mjs` - rspack config factory for a phase remote.
- `runtime/mountRemote.tsx` - React root mount for the standalone dev page.
- `runtime/StandaloneNotice.tsx` - the notice that standalone page renders.

Styling is not part of this scaffolding: `clients/core` builds the single Tailwind stylesheet for
the whole shell and scans every component directory for it. Remotes ship no Tailwind build of their
own. See the styling section of `.claude/rules/module-federation/remotes.md`.

External phases live in their own repositories and cannot import from here; keep
`template-repository/` in sync by hand.
