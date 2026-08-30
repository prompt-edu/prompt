import type { APIRequestContext } from '@playwright/test'

// The server allows a deletion or export run up to 30 minutes, but on this stack
// it is normally seconds. Three minutes keeps failure diagnostics useful without
// eating the shard's 20-minute global budget.
export const PRIVACY_WAIT_MS = 3 * 60 * 1000

// Each spec that waits on one of these runs needs more than Playwright's 60s
// default, or the test is preempted before the helper's own deadline.
export const PRIVACY_TEST_TIMEOUT_MS = 4 * 60 * 1000

export interface DeletionRequest {
  id: string
  status: string
  subrequests?: { source_name: string; status: string }[]
}

export interface PrivacyExport {
  id: string
  status: string
  documents: { id: string; source_name: string; status: string }[]
}

export type LatestDeletion = { status: 'ready' } | { status: 'exists'; request: DeletionRequest }

export type LatestExport =
  | { status: 'ready' }
  | { status: 'rate_limited'; retry_after: string }
  | { status: 'exists'; export: PrivacyExport }

async function readJson<T>(api: APIRequestContext, path: string): Promise<T | undefined> {
  const res = await api.get(path)
  if (res.status() === 204) return undefined
  if (!res.ok()) throw new Error(`GET ${path} failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as T
}

export async function latestDeletion(api: APIRequestContext): Promise<LatestDeletion> {
  return (await readJson<LatestDeletion>(api, '/api/privacy/data-deletion')) ?? { status: 'ready' }
}

export async function latestExport(api: APIRequestContext): Promise<LatestExport> {
  return (await readJson<LatestExport>(api, '/api/privacy/data-export')) ?? { status: 'ready' }
}

export async function getDeletion(api: APIRequestContext, id: string): Promise<DeletionRequest> {
  return (await readJson<DeletionRequest>(
    api,
    `/api/privacy/data-deletion/${id}`,
  )) as DeletionRequest
}

export async function getExport(api: APIRequestContext, id: string): Promise<PrivacyExport> {
  return (await readJson<PrivacyExport>(api, `/api/privacy/data-export/${id}`)) as PrivacyExport
}

interface WaitOptions<T> {
  describe: string
  poll: () => Promise<T>
  done: (value: T) => boolean
  // A run that ends in failure is a defect, not something to wait out: report it
  // immediately with the payload rather than after the full deadline.
  failed?: (value: T) => boolean
  timeoutMs?: number
}

export async function waitFor<T>({
  describe,
  poll,
  done,
  failed,
  timeoutMs = PRIVACY_WAIT_MS,
}: WaitOptions<T>): Promise<T> {
  const deadline = Date.now() + timeoutMs
  for (;;) {
    const last = await poll()
    if (failed?.(last)) {
      throw new Error(`${describe} failed: ${JSON.stringify(last)}`)
    }
    if (done(last)) return last
    if (Date.now() > deadline) {
      throw new Error(`${describe} did not settle within ${timeoutMs}ms: ${JSON.stringify(last)}`)
    }
    await new Promise((resolve) => setTimeout(resolve, 500))
  }
}

const EXPORT_TERMINAL = ['complete', 'no_data', 'failed', 'archived']
const DELETION_TERMINAL = ['succeeded', 'failed', 'rejected']

export function isExportTerminal(status: string): boolean {
  return EXPORT_TERMINAL.includes(status)
}

export function isDeletionTerminal(status: string): boolean {
  return DELETION_TERMINAL.includes(status)
}

// The owner-scoped GET /data-deletion/:id rejects anyone but the requester, so an
// admin watching a run reads it from the console listing instead.
export async function adminDeletion(
  admin: APIRequestContext,
  id: string,
): Promise<DeletionRequest> {
  const res = await admin.get('/api/privacy/admin/data-deletions')
  if (!res.ok()) {
    throw new Error(`GET admin/data-deletions failed: ${res.status()} ${await res.text()}`)
  }
  const requests = (await res.json()) as DeletionRequest[]
  const found = requests.find((request) => request.id === id)
  if (!found) throw new Error(`deletion request ${id} is not listed in the admin console`)
  return found
}

// Leaves no export behind and frees the subject's 30-day rate limit, which is
// otherwise one export per user for the lifetime of the stack.
export async function archiveExports(admin: APIRequestContext, userEmail: string): Promise<void> {
  const res = await admin.get('/api/privacy/admin/data-exports')
  if (!res.ok()) return
  const exports = (await res.json()) as { id: string; student_email?: string; status: string }[]
  for (const record of exports) {
    if (record.student_email === userEmail) {
      await admin.delete(`/api/privacy/admin/data-exports/${record.id}?reset_rate_limit=true`)
    }
  }
}
