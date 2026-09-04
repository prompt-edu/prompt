import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { Role } from '../../src/data/roles'
import { BASE_URL, INFRASTRUCTURE_SETUP_API } from '../../src/env'

// All infrastructure setup API calls go through the client-core nginx proxy on
// the browser origin (same path prefix as prod Traefik), NOT the core API.
export function infrastructureSetupUrl(phaseId: string, path: string): string {
  return `${BASE_URL}${INFRASTRUCTURE_SETUP_API}/course_phase/${phaseId}/${path}`
}

// 192.0.2.0/24 is TEST-NET-1 (RFC 5737): guaranteed unroutable, so the connection
// attempt is blackholed rather than refused and CreateResource blocks instead of
// failing fast. That keeps the instances non-terminal for as long as the test needs,
// which is what makes the 409 on a second trigger deterministic rather than a race
// against the worker.
export const UNRESPONSIVE_PROVIDER_URL = 'http://192.0.2.1:9999'

export interface ApiResponse {
  status: number
  body: string
}

async function call(
  role: Role,
  method: 'get' | 'put' | 'post' | 'delete',
  phaseId: string,
  path: string,
  data?: Record<string, unknown>,
): Promise<ApiResponse> {
  const api: APIRequestContext = await apiContextFor(role)
  try {
    const res = await api[method](infrastructureSetupUrl(phaseId, path), data ? { data } : {})
    return { status: res.status(), body: await res.text() }
  } finally {
    await api.dispose()
  }
}

async function expectOk(response: ApiResponse, action: string): Promise<ApiResponse> {
  if (response.status < 200 || response.status >= 300) {
    throw new Error(`${action} failed: ${response.status} ${response.body}`)
  }
  return response
}

export async function setSemesterTag(
  phaseId: string,
  semesterTag: string,
  role: Role = 'lecturer',
): Promise<void> {
  await expectOk(await call(role, 'put', phaseId, 'setup-config', { semesterTag }), 'set semester tag')
}

export async function upsertGitlabProvider(
  phaseId: string,
  baseUrl: string,
  role: Role = 'lecturer',
): Promise<void> {
  await expectOk(
    await call(role, 'put', phaseId, 'provider-configs', {
      providerType: 'gitlab',
      credentials: { base_url: baseUrl, private_token: 'e2e-token' },
    }),
    'upsert gitlab provider',
  )
}

export async function createStudentResourceConfig(
  phaseId: string,
  nameTemplate: string,
  role: Role = 'lecturer',
): Promise<void> {
  await expectOk(
    await call(role, 'post', phaseId, 'resource-configs', {
      providerType: 'gitlab',
      resourceType: 'group',
      scope: 'per_student',
      nameTemplate,
      permissionMapping: { student: 'developer' },
      resourceExtraConfig: {},
    }),
    'create resource config',
  )
}

export function createResourceConfigRaw(
  phaseId: string,
  body: Record<string, unknown>,
  role: Role = 'lecturer',
): Promise<ApiResponse> {
  return call(role, 'post', phaseId, 'resource-configs', body)
}

export function triggerExecution(phaseId: string, role: Role = 'lecturer'): Promise<ApiResponse> {
  return call(role, 'post', phaseId, 'execute')
}

export function listProviderConfigs(phaseId: string, role: Role = 'lecturer'): Promise<ApiResponse> {
  return call(role, 'get', phaseId, 'provider-configs')
}

export function listInstances(phaseId: string, role: Role = 'lecturer'): Promise<ApiResponse> {
  return call(role, 'get', phaseId, 'instances')
}

export function getPhaseConfig(phaseId: string, role: Role = 'lecturer'): Promise<ApiResponse> {
  return call(role, 'get', phaseId, 'config')
}

// Idempotent reset, run before and after the journey so a retry (or a rerun against
// a stack that is still up) starts from the same state. Removing the provider
// cascades to its resource configs and their instances, which also clears any
// non-terminal instance that would otherwise make the next trigger a conflict.
export async function resetInfrastructureSetupPhase(
  phaseId: string,
  role: Role = 'lecturer',
): Promise<void> {
  await call(role, 'delete', phaseId, 'provider-configs/gitlab')
  await call(role, 'put', phaseId, 'setup-config', { semesterTag: '' })
}
