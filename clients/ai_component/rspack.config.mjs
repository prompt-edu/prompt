import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'ai_component',
  port: 3013,
  configUrl: import.meta.url,
})
