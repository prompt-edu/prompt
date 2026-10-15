---
name: module-federation-remote
description: Register or expose a Module Federation remote in PROMPT 2.0 — wire a micro-frontend's exposes and the core shell's remotes with the cache-busting pattern. Use when connecting a component to core, exposing a new module, or debugging a remote that won't load.
---

PROMPT 2.0 composes micro-frontends with Module Federation. The core shell (`clients/core`)
is the host; each `<name>_component` is a remote. All bundler config is in `rspack.config.mjs` files.

## Expose from the remote component

In-repo remotes do not write their own `ModuleFederationPlugin`. The whole config comes from
`clients/shared/rspack/createRspackConfig.mjs`, so `clients/<name>_component/rspack.config.mjs` is:

```js
import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'your_component',           // '<name>_component'; must equal the core remotes key
  port: 3011,                       // dev server port, unique per remote
  configUrl: import.meta.url,       // resolves the component's own directory
})
```

The factory sets `filename: 'remoteEntry.js'`, the `exposes` map (`./routes` → `RouteObject[]`,
`./sidebar` → `SidebarMenuItemProps`, `./provide` → components other phases may render), the
loaders, the output settings, and the share scope. A remote needing extra module resolution
(assessment's `@hookform/resolvers`) passes `resolveAlias: (componentDir) => ({ … })`; anything else
that differs belongs in the factory as a new option, not in a forked config.

The singleton share scope lives in `clients/shared/rspack/federatedDependencies.mjs` and is imported
by the host and every remote, so host and remotes cannot drift apart.

`routes/` and `sidebar/` are directories next to `src/`, each with an `index.tsx` default export.
There is no `./App` expose — core mounts a phase through its routes, not through a root component.

## Standalone dev page

Each remote also builds as a standalone page that only renders a notice, since a phase is meant to
run inside core. That is `src/bootstrap.tsx`, two lines using
`clients/shared/runtime/mountRemote.tsx` and `clients/shared/runtime/StandaloneNotice.tsx`. The root
element id must match the `<div id="…">` in the component's `public/template.html`.

## Register in core (host)

Add the remote to `clients/core/remotes.config.mjs`:

```js
// clients/core/remotes.config.mjs
export const REMOTES = [
  {
    name: 'your_component', // the federation name, also the core remotes key
    phaseTypeName: 'Your Phase', // the course phase type it renders: its PhaseRouterMapping key
    devPort: 3011,
    prodPath: '/your-component',
  },
]
```

`clients/core/rspack.config.mjs` builds the federation `remotes` from this list, resolving each entry
to the dev port in development and the reverse-proxy path in production, with a cache-busting query
so a redeploy forces a reload. The URL is derived from `IS_DEV`, not from an environment variable, so
nothing needs to be added to `.env.template` or `.env.dev.template`.

The same list reaches the app as `__PROMPT_REMOTES__`, and the admin System Status page probes each
remote's `remoteEntry.js` and `mf-manifest.json`. `clients/core/remotes.config.test.ts` fails when
the list and the `PhaseRouterMapping` keys drift apart.

## Load dynamically

Add one file per remote under `clients/core/src/managementConsole/PhaseMapping/ExternalRoutes/`
(and `ExternalSidebars/`), following the existing files — they lazy-load the remote and fall back to
`LoadingError` when it cannot be reached:

```typescript
export const YourRoutes = React.lazy(() =>
  import('your_component/routes')
    .then((module): { default: React.FC } => ({
      default: () => <ExternalRoutes routes={module.default || []} />,
    }))
    .catch((): { default: React.FC } => ({
      default: () => <LoadingError phaseTitle={'Your Phase'} />,
    })),
)
```

`clients/core/src/declaration.d.ts` already types `*_component/routes`, `*_component/sidebar`, and
`*_component/provide`, so no per-remote declaration is needed.

## Verify / debug

- `name` in the remote MUST match the key used in core's `remotes` and the import specifier.
- Keep the `shared` entries above `singleton: true` on both sides — version mismatches are the usual
  cause of runtime federation errors.
- Start the remote on its dev port and confirm `…/remoteEntry.js` is reachable; then load core. A
  remote that fails to load surfaces as the `LoadingError` page rather than a crash.
