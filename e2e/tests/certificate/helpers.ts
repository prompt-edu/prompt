import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { Role } from '../../src/data/roles'
import { BASE_URL, CERTIFICATE_API } from '../../src/env'

// All certificate API calls go through the client-core nginx proxy on the
// browser origin (same path prefix as prod Traefik), NOT the core API.
export function certificateUrl(phaseId: string, path: string): string {
  return `${BASE_URL}${CERTIFICATE_API}/course_phase/${phaseId}/${path}`
}

// A minimal, valid Typst template the bundled compiler (in the certificate
// image) can render. It reads the data.json the server writes and references
// only studentName + courseName, so it compiles for phases without team data.
export const E2E_TEMPLATE = [
  '#let data = json("data.json")',
  '#set page(paper: "a4")',
  '= Certificate of Completion',
  'This certifies that #data.studentName has completed #data.courseName.',
].join('\n')

async function put(
  role: Role,
  phaseId: string,
  path: string,
  data: Record<string, unknown>,
): Promise<void> {
  const api: APIRequestContext = await apiContextFor(role)
  try {
    const res = await api.put(certificateUrl(phaseId, path), { data })
    if (!res.ok()) {
      throw new Error(`PUT ${path} failed: ${res.status()} ${await res.text()}`)
    }
  } finally {
    await api.dispose()
  }
}

export async function putTemplate(
  phaseId: string,
  templateContent: string,
  role: Role = 'lecturer',
): Promise<void> {
  await put(role, phaseId, 'config', { templateContent })
}

export async function setReleaseDate(
  phaseId: string,
  releaseDate: string | null,
  role: Role = 'lecturer',
): Promise<void> {
  await put(role, phaseId, 'config/release-date', { releaseDate })
}

// Idempotent teardown so CI retries (which reuse the ephemeral DB) stay
// deterministic. Certificate downloads cannot be un-recorded, but clearing the
// release date restores the phase's release gating.
export async function resetCertificatePhase(phaseId: string, role: Role = 'admin'): Promise<void> {
  await setReleaseDate(phaseId, null, role)
}
