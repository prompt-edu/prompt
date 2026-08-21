---
paths:
  - "**/rspack.config.mjs"
---

# Module Federation

Micro-frontends are composed via Module Federation (rspack's `ModuleFederationPlugin`, imported as
`rspack.container.ModuleFederationPlugin`): `clients/core` is the host; each `<name>_component` is a
remote. For the step-by-step wiring use the **`module-federation-remote`** skill.

## Expose (remote component)

Every in-repo remote gets its whole rspack config from the shared factory, so
`clients/<name>_component/rspack.config.mjs` is only the two values that differ:

```js
import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'your_component',           // must equal the core remotes key
  port: 3011,                       // dev server port, unique per remote
  configUrl: import.meta.url,       // resolves the component's own directory
})
```

The factory (`clients/shared/rspack/createRspackConfig.mjs`) owns `filename: 'remoteEntry.js'`, the
`exposes` map (`./routes`, `./sidebar`, `./provide`), the loaders, the output settings, and the share
scope. Do not fork it per component; add an option instead. A remote that needs extra resolution
(assessment's `@hookform/resolvers`) passes `resolveAlias: (componentDir) => ({ … })`.

A component exposes whole directories (`routes/`, `sidebar/`) that sit next to `src/`, not modules
inside `src/`. It does not expose `./App`.

## Share scope

`clients/shared/rspack/federatedDependencies.mjs` is the single source of the singleton share scope
(`react`, `react-dom`, `react-router-dom`, `@tanstack/react-query`, `@tumaet/prompt-shared-state`).
Both the host and every remote import it, which is what keeps them from drifting apart. Changing the
set changes it for host and remotes at once, which is the only safe way to change it.

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
