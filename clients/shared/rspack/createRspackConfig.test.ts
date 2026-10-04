import { describe, expect, it } from 'vitest'
import { createRspackConfig } from './createRspackConfig.mjs'

const federationOptions = (consumesCore: boolean | undefined, NODE_ENV: string) => {
  const config = createRspackConfig({
    name: 'interview_component',
    port: 3002,
    configUrl: import.meta.url,
    consumesCore,
  })({ NODE_ENV })
  const plugin = config.plugins.find((p) => p.constructor.name === 'ModuleFederationPlugin')
  return plugin._options
}

describe('createRspackConfig', () => {
  it('does not consume core by default', () => {
    const options = federationOptions(undefined, 'development')
    expect(options.remotes).toBeUndefined()
    expect(options.shareStrategy).toBeUndefined()
  })

  it('loads core only when a core module is first imported', () => {
    expect(federationOptions(true, 'development').shareStrategy).toBe('loaded-first')
  })

  it('consumes core from the dev server in development', () => {
    expect(federationOptions(true, 'development').remotes.core).toMatch(
      /^core@http:\/\/localhost:3000\/remoteEntry\.js\?\d+$/,
    )
  })

  it('consumes core from the site root in production', () => {
    expect(federationOptions(true, 'production').remotes.core).toMatch(
      /^core@\/remoteEntry\.js\?\d+$/,
    )
  })
})
