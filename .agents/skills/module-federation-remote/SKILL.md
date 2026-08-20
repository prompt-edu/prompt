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

In `clients/core/rspack.config.mjs`, resolve the URL next to the other `*URL` constants and add the
remote using the cache-busting query so a redeploy forces a reload:

```js
const yourComponentURL = IS_DEV ? `http://localhost:3011` : `/your-component`

remotes: {
  your_component: `your_component@${yourComponentURL}/remoteEntry.js?${Date.now()}`,
}
```

The URL is derived from `IS_DEV`, not from an environment variable, so nothing needs to be added to
`.env.template` or `.env.dev.template`. In production the path is served by the reverse proxy.

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
