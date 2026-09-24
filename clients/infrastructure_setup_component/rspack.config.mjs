import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'infrastructure_setup_component',
  port: 3012,
  configUrl: import.meta.url,
})
