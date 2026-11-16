import { describe, expect, it } from 'vitest'
import { devPort, devPortOffset, shiftLocalhostPorts } from './devPorts.mjs'

describe('devPorts', () => {
  it('keeps the stack 0 ports without PROMPT_CLIENT_PORT', () => {
    expect(devPortOffset({})).toBe(0)
    expect(devPort(3007, {})).toBe(3007)
  })

  it('shifts every port by the distance of the base port from 3000', () => {
    const env = { PROMPT_CLIENT_PORT: '3200' }
    expect(devPortOffset(env)).toBe(200)
    expect(devPort(3000, env)).toBe(3200)
    expect(devPort(3012, env)).toBe(3212)
  })

  it('rejects a base port that is not a number', () => {
    expect(() => devPortOffset({ PROMPT_CLIENT_PORT: 'abc' })).toThrow(/PROMPT_CLIENT_PORT/)
  })

  it('shifts the localhost ports of env.js', () => {
    const envJs = "CORE_HOST: 'http://localhost:8080',\nKEYCLOAK_HOST: 'http://localhost:8081',"
    expect(shiftLocalhostPorts(envJs, 100)).toBe(
      "CORE_HOST: 'http://localhost:8180',\nKEYCLOAK_HOST: 'http://localhost:8181',",
    )
  })
})
