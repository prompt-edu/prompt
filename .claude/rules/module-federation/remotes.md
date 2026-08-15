---
paths:
  - "**/rspack.config.mjs"
---

# Module Federation

Micro-frontends are composed via Module Federation (rspack's `ModuleFederationPlugin`, imported as
`rspack.container.ModuleFederationPlugin`): `clients/core` is the host; each `<name>_component` is a
remote. For the step-by-step wiring use the **`module-federation-remote`** skill.

## Expose (remote component)

```js
new ModuleFederationPlugin({
  name: COMPONENT_NAME,             // must equal the core remotes key
  filename: 'remoteEntry.js',
  exposes: {
    './routes': './routes',         // RouteObject[] default export
    './sidebar': './sidebar',       // SidebarMenuItemProps default export
    './provide': './src/provide',   // optional, for cross-phase components
  },
})
```

A component exposes whole directories (`routes/`, `sidebar/`) that sit next to `src/`, not modules
inside `src/`. It does not expose `./App`. The top of each component's `rspack.config.mjs` sets
`COMPONENT_NAME` and `COMPONENT_DEV_PORT`.

## Register (core host)

```js
const yourComponentURL = IS_DEV ? `http://localhost:3011` : `/your-component`

remotes: {
  your_component: `your_component@${yourComponentURL}/remoteEntry.js?${Date.now()}`,
}
```

Remote URLs are resolved in `clients/core/rspack.config.mjs` from `IS_DEV`, not from environment
variables: the dev port in development, the reverse-proxy path in production. The `?${Date.now()}`
cache-buster forces a reload after redeploy.

Core consumes a remote lazily, one file per remote under
`src/managementConsole/PhaseMapping/ExternalRoutes/` (and `ExternalSidebars/`):

```typescript
export const YourRoutes = React.lazy(() => import('your_component/routes').then(/* … */))
```

The `*_component/routes`, `*_component/sidebar`, and `*_component/provide` module shapes are typed
once in `clients/core/src/declaration.d.ts`.

## Rules

- `react`, `react-dom`, `react-router-dom`, `@tanstack/react-query`, and
  `@tumaet/prompt-shared-state` must be `singleton: true` on both host and remote.
- The federation `name`, the core `remotes` key, and the import specifier must all match exactly.
