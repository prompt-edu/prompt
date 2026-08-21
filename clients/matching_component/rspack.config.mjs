import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'matching_component',
  port: 3003,
  configUrl: import.meta.url,
})
