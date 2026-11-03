import { afterEach, describe, expect, it, vi } from 'vitest'

const env = window.env as { AI_HOST?: string }
const originalHost = env.AI_HOST

const loadAIInstance = async (host: string) => {
  env.AI_HOST = host
  vi.resetModules()
  return (await import('./client')).aiAxiosInstance
}

afterEach(() => {
  env.AI_HOST = originalHost
})

describe('aiAxiosInstance', () => {
  it('adds https to the bare host production sets', async () => {
    const instance = await loadAIInstance('ai.test')
    expect(instance.getUri({ url: '/ai/api/info' })).toBe('https://ai.test/ai/api/info')
  })

  it('keeps a host that already has a scheme', async () => {
    const instance = await loadAIInstance('http://localhost:8092')
    expect(instance.getUri({ url: '/ai/api/info' })).toBe('http://localhost:8092/ai/api/info')
  })
})
