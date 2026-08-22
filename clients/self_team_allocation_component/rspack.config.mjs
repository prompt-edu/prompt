import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'self_team_allocation_component',
  port: 3009,
  configUrl: import.meta.url,
})
