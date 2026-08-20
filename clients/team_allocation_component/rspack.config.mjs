import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'team_allocation_component',
  port: 3008,
  configUrl: import.meta.url,
})
