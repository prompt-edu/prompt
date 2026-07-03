---
paths:
  - "**/webpack.config.mjs"
---

# Module Federation

Micro-frontends are composed via Webpack Module Federation: `clients/core` is the host; each
`<name>_component` is a remote. For the step-by-step wiring use the **`module-federation-remote`** skill.

## Expose (remote component)

```js
new ModuleFederationPlugin({
  name: 'your_component',           // must equal COMPONENT_NAME and the core remotes key
  filename: 'remoteEntry.js',
  exposes: { './App': './src/App', './sidebar': './src/Sidebar' },
})
```

The top of each component's `webpack.config.mjs` sets `COMPONENT_NAME` and `COMPONENT_DEV_PORT`.

## Register (core host)

```js
remotes: {
  your_component: `your_component@${yourComponentURL}/remoteEntry.js?${Date.now()}`,
}
```

The `?${Date.now()}` cache-buster forces a reload after redeploy. Define the matching `*URL` from an
env var. Load lazily: `React.lazy(() => import('your_component/App'))`.

## Rules

- `react`, `react-dom`, `react-router-dom` must be `singleton: true` on both host and remote.
- The federation `name`, the core `remotes` key, and the import specifier must all match exactly.
