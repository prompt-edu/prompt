import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import { REMOTES, resolveRemotes } from './remotes.config.mjs'

const routerSource = readFileSync(
  new URL('./src/managementConsole/PhaseMapping/PhaseRouterMapping.tsx', import.meta.url),
  'utf8',
)
const routerBlock = routerSource.match(/const PhaseRouter\b[^=]*= \{([\s\S]*?)\n\}/)?.[1] ?? ''
const routerKeys = [...routerBlock.matchAll(/^\s+'?([^':\n]+)'?:\s*\w+,$/gm)].map((m) => m[1])

const PHASE_TYPES_RENDERED_BY_CORE = ['Application']

describe('REMOTES', () => {
  it('has a remote for every phase type PhaseRouterMapping renders from one', () => {
    const remotePhaseTypes = routerKeys.filter((key) => !PHASE_TYPES_RENDERED_BY_CORE.includes(key))

    expect(remotePhaseTypes.length).toBeGreaterThan(0)
    expect(REMOTES.map((remote) => remote.phaseTypeName).sort()).toEqual(remotePhaseTypes.sort())
  })

  it('gives every remote its own name, dev port, and production path', () => {
    for (const field of ['name', 'devPort', 'prodPath'] as const) {
      const values = REMOTES.map((remote) => remote[field])
      expect(new Set(values).size, field).toBe(values.length)
    }
  })

  it('resolves the dev server in development and the proxy path in production', () => {
    const interview = (isDev: boolean) =>
      resolveRemotes(isDev).find((remote) => remote.name === 'interview_component')

    expect(interview(true)).toEqual({
      name: 'interview_component',
      phaseTypeName: 'Interview',
      url: 'http://localhost:3002',
    })
    expect(interview(false)?.url).toBe('/interview')
  })
})
