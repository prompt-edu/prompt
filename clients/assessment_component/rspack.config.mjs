import path from 'node:path'
import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'assessment_component',
  port: 3007,
  configUrl: import.meta.url,
  resolveAlias: (componentDir) => ({
    '@hookform/resolvers': path.resolve(componentDir, '../node_modules/@hookform/resolvers'),
    '@hookform/resolvers/zod': path.resolve(
      componentDir,
      '../node_modules/@hookform/resolvers/zod',
    ),
  }),
})
