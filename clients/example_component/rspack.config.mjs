import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'example_component',
  port: 3001,
  configUrl: import.meta.url,
})
