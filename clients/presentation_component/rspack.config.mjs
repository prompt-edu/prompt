import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'presentation_component',
  port: 3011,
  configUrl: import.meta.url,
})
