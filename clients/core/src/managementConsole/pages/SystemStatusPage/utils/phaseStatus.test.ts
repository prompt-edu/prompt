import { describe, expect, it } from 'vitest'
import type { ClientRemote } from '../interfaces/clientInfo'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'
import type { ServiceInfo } from '../interfaces/serviceCapabilities'
import {
  buildPhaseStatusEntries,
  clientStatus,
  type ProbeResult,
  phaseCardStatus,
  serverStatus,
} from './phaseStatus'

const phaseType = (name: string, baseUrl: string): CoursePhaseType => ({
  id: `id-${name}`,
  name,
  initialPhase: false,
  baseUrl,
  description: '',
})

const remote = (name: string, phaseTypeName: string): ClientRemote => ({
  name,
  phaseTypeName,
  url: `/${name}`,
})

const probe = <T>(overrides: Partial<ProbeResult<T>>): ProbeResult<T> => ({
  data: undefined,
  isPending: false,
  isError: false,
  ...overrides,
})

const serviceInfo = (healthy: boolean): ServiceInfo => ({
  serviceName: 'interview',
  healthy,
  capabilities: {},
})

describe('buildPhaseStatusEntries', () => {
  const interview = phaseType('Interview', 'http://localhost:8087/interview/api')
  const matching = phaseType('Matching', 'core')
  const application = phaseType('Application', 'core')
  const remotes = [
    remote('interview_component', 'Interview'),
    remote('matching_component', 'Matching'),
    remote('example_component', 'example_component'),
  ]

  it('pairs a microservice phase type with its remote', () => {
    expect(buildPhaseStatusEntries([interview], remotes)).toEqual([
      { phaseType: interview, hasServerInfo: true, remote: remotes[0] },
    ])
  })

  it('keeps a phase type without an /info endpoint that loads a remote', () => {
    expect(buildPhaseStatusEntries([matching], remotes)).toEqual([
      { phaseType: matching, hasServerInfo: false, remote: remotes[1] },
    ])
  })

  it('keeps a microservice phase type whose remote core does not know', () => {
    expect(buildPhaseStatusEntries([interview], [])).toEqual([
      { phaseType: interview, hasServerInfo: true, remote: undefined },
    ])
  })

  it('sorts the cards by phase type name', () => {
    const entries = buildPhaseStatusEntries([matching, interview], remotes)

    expect(entries.map((entry) => entry.phaseType.name)).toEqual(['Interview', 'Matching'])
  })

  it('drops a phase type with neither an /info endpoint nor a remote, and an unpaired remote', () => {
    const entries = buildPhaseStatusEntries([interview, matching, application], remotes)

    expect(entries.map((entry) => entry.phaseType.name)).toEqual(['Interview', 'Matching'])
    expect(entries.map((entry) => entry.remote?.name)).not.toContain('example_component')
  })
})

describe('serverStatus', () => {
  it('is offline when the probe failed or answered nothing', () => {
    expect(serverStatus(probe({ isError: true }))).toBe('Offline')
    expect(serverStatus(probe({}))).toBe('Offline')
  })

  it('is degraded when the service reports itself unhealthy', () => {
    expect(serverStatus(probe({ data: serviceInfo(false) }))).toBe('OnlineUnhealthy')
  })

  it('is online when the service reports itself healthy', () => {
    expect(serverStatus(probe({ data: serviceInfo(true) }))).toBe('Online')
  })
})

describe('clientStatus', () => {
  it('is online once the remote answered, with or without a version', () => {
    expect(clientStatus(probe({ data: { buildVersion: 'v2.18.1' } }))).toBe('Online')
    expect(clientStatus(probe({ data: {} }))).toBe('Online')
  })

  it('is offline when the remote failed to load', () => {
    expect(clientStatus(probe({ isError: true }))).toBe('Offline')
  })
})

describe('phaseCardStatus', () => {
  const up = probe<ServiceInfo>({ data: serviceInfo(true) })
  const unhealthy = probe<ServiceInfo>({ data: serviceInfo(false) })
  const down = probe<ServiceInfo>({ isError: true })
  const clientUp = probe({ data: {} })
  const clientDown = probe({ isError: true })

  it('is online when every present part is up', () => {
    expect(phaseCardStatus(up, clientUp)).toBe('Online')
    expect(phaseCardStatus(undefined, clientUp)).toBe('Online')
    expect(phaseCardStatus(up, undefined)).toBe('Online')
  })

  it('is degraded when only the server or only the client is down', () => {
    expect(phaseCardStatus(down, clientUp)).toBe('OnlineUnhealthy')
    expect(phaseCardStatus(up, clientDown)).toBe('OnlineUnhealthy')
  })

  it('is degraded when the server reports itself unhealthy', () => {
    expect(phaseCardStatus(unhealthy, clientUp)).toBe('OnlineUnhealthy')
  })

  it('is offline when every present part is down', () => {
    expect(phaseCardStatus(down, clientDown)).toBe('Offline')
    expect(phaseCardStatus(undefined, clientDown)).toBe('Offline')
    expect(phaseCardStatus(down, undefined)).toBe('Offline')
  })

  it('has no status while a probe is still pending', () => {
    expect(phaseCardStatus(probe({ isPending: true }), clientUp)).toBeUndefined()
    expect(phaseCardStatus(up, probe({ isPending: true }))).toBeUndefined()
  })
})
