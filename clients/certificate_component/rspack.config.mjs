import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'certificate_component',
  port: 3010,
  configUrl: import.meta.url,
})
