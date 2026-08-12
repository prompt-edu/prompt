---
name: module-federation-remote
description: Register or expose a Module Federation remote in PROMPT 2.0 — wire a micro-frontend's exposes and the core shell's remotes with the cache-busting pattern. Use when connecting a component to core, exposing a new module, or debugging a remote that won't load.
---

PROMPT 2.0 composes micro-frontends with Module Federation. The core shell (`clients/core`)
is the host; each `<name>_component` is a remote. All bundler config is in `rspack.config.mjs` files.

## Expose from the remote component

In `clients/<name>_component/rspack.config.mjs`, the `ModuleFederationPlugin`
(`const { ModuleFederationPlugin } = rspack.container`):

```js
new ModuleFederationPlugin({
  name: COMPONENT_NAME,             // '<name>_component'; must equal the core remotes key
  filename: 'remoteEntry.js',
  exposes: {
    './routes': './routes',         // default export: RouteObject[]
    './sidebar': './sidebar',       // default export: SidebarMenuItemProps
    './provide': './src/provide',   // optional; components other phases may render
  },
  shared: {
    react: { singleton: true, requiredVersion: deps.react },
    'react-dom': { singleton: true, requiredVersion: deps['react-dom'] },
    'react-router-dom': { singleton: true, requiredVersion: deps['react-router-dom'] },
    '@tanstack/react-query': { singleton: true, requiredVersion: deps['@tanstack/react-query'] },
    '@tumaet/prompt-shared-state': {
      singleton: true,
      requiredVersion: deps['@tumaet/prompt-shared-state'],
    },
  },
})
```

`routes/` and `sidebar/` are directories next to `src/`, each with an `index.tsx` default export.
There is no `./App` expose — core mounts a phase through its routes, not through a root component.

## Register in core (host)

In `clients/core/rspack.config.mjs`, resolve the URL next to the other `*URL` constants and add the
remote using the cache-busting query so a redeploy forces a reload:

```js
const yourComponentURL = IS_DEV ? `http://localhost:<COMPONENT_DEV_PORT>` : `/your-component`

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
