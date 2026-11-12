import rspack from '@rspack/core'
import { describe, expect, it } from 'vitest'
import { createRspackConfig } from './createRspackConfig.mjs'

const federationOptions = (consumesCore: boolean | undefined, NODE_ENV: string) => {
  const config = createRspackConfig({
    name: 'interview_component',
    port: 3002,
    configUrl: import.meta.url,
    consumesCore,
  })({ NODE_ENV })
  const plugin = config.plugins.find((p) => p instanceof rspack.container.ModuleFederationPlugin)
  return plugin._options
}

describe('createRspackConfig', () => {
  it('does not consume core by default', () => {
    const options = federationOptions(undefined, 'development')
    expect(options.remotes).toBeUndefined()
    expect(options.shareStrategy).toBeUndefined()
  })

  it('uses the loaded-first share strategy', () => {
    expect(federationOptions(true, 'development').shareStrategy).toBe('loaded-first')
  })

  it('consumes core from the dev server in development', () => {
    expect(federationOptions(true, 'development').remotes.core).toMatch(
      /^core@http:\/\/localhost:3000\/remoteEntry\.js\?\d+$/,
    )
  })

  it('follows the stack base port in development', () => {
    process.env.PROMPT_CLIENT_PORT = '3100'
    try {
      const config = createRspackConfig({
        name: 'interview_component',
        port: 3002,
        configUrl: import.meta.url,
        consumesCore: true,
      })({ NODE_ENV: 'development' })
      const plugin = config.plugins.find(
        (p) => p instanceof rspack.container.ModuleFederationPlugin,
      )
      expect(config.devServer.port).toBe(3102)
      expect(plugin._options.remotes.core).toMatch(/^core@http:\/\/localhost:3100\/remoteEntry\.js/)
    } finally {
      delete process.env.PROMPT_CLIENT_PORT
    }
  })

  it('consumes core from the site root in production', () => {
    expect(federationOptions(true, 'production').remotes.core).toMatch(
      /^core@\/remoteEntry\.js\?\d+$/,
    )
  })
})
