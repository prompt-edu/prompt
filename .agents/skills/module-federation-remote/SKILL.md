---
name: module-federation-remote
description: Register or expose a Webpack Module Federation remote in PROMPT 2.0 — wire a micro-frontend's exposes and the core shell's remotes with the cache-busting pattern. Use when connecting a component to core, exposing a new module, or debugging a remote that won't load.
---

PROMPT 2.0 composes micro-frontends with Webpack Module Federation. The core shell (`clients/core`)
is the host; each `<name>_component` is a remote. All webpack config is in `webpack.config.mjs` files.

## Expose from the remote component

In `clients/<name>_component/webpack.config.mjs`, the `ModuleFederationPlugin`:

```js
new ModuleFederationPlugin({
  name: '<name>_component',          // must equal COMPONENT_NAME constant
  filename: 'remoteEntry.js',
  exposes: {
    './App': './src/App',
    './sidebar': './src/Sidebar',    // for management-console integration
  },
  shared: { /* react, react-dom, react-router-dom as singletons */ },
})
```

## Register in core (host)

In `clients/core/webpack.config.mjs`, add to `remotes:` using the cache-busting query so a redeploy
forces a reload:

```js
<name>_component: `<name>_component@${<name>URL}/remoteEntry.js?${Date.now()}`,
```

Define the matching `<name>URL` from an env var alongside the other `*URL` resolutions in that file,
and add the URL to `.env.template` / `.env.dev.template`.

## Load dynamically

```typescript
const Component = React.lazy(() => import('<name>_component/App'))
```

## Verify / debug

- `name` in the remote MUST match the key used in core's `remotes` and the import specifier.
- Keep `react`, `react-dom`, `react-router-dom` as `singleton: true` on both sides — version
  mismatches are the usual cause of runtime federation errors.
- Start the remote on its dev port and confirm `…/remoteEntry.js` is reachable; then load core.
