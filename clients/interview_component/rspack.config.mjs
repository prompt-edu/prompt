import { createRspackConfig } from '../shared/rspack/createRspackConfig.mjs'

export default createRspackConfig({
  name: 'interview_component',
  port: 3002,
  configUrl: import.meta.url,
})
