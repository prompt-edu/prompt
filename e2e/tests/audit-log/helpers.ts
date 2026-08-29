import { APIRequestContext, expect } from '@playwright/test'

// The audit table is created by a migration that runs after the seed dump, so
// audit rows cannot be seeded. Driving the API to produce them also exercises
// the real capture path.

export interface AuditEntry {
  id: string
  createdAt: string
  action: string
  outcome: string
}

export interface AuditLogPage {
  entries: AuditEntry[]
  nextCursor: { createdAt: string; id: string } | null
}

// Each archive toggle is one captured mutation on the course, so it lands in
// that course's audit log. An admin's toggle succeeds; a student's is denied
// with a 403 and captured as such, which is what makes the two pages of the
// generated log tell each other apart.
async function toggleArchive(
  api: APIRequestContext,
  courseId: string,
  count: number,
  expectedStatus: number,
): Promise<void> {
  for (let i = 0; i < count; i++) {
    const res = await api.put(`/api/courses/${courseId}/archive`, {
      data: { archived: i % 2 === 0 },
    })
    expect(res.status()).toBe(expectedStatus)
  }
}

export async function getCourseAuditLog(
  api: APIRequestContext,
  courseId: string,
  query = '',
): Promise<AuditLogPage> {
  const res = await api.get(`/api/courses/${courseId}/audit-log${query}`)
  if (!res.ok()) {
    throw new Error(`GET audit-log failed: ${res.status()} ${await res.text()}`)
  }
  return (await res.json()) as AuditLogPage
}

// Capture is delivered from a background goroutine, so the entries of one batch
// have to be visible before the next batch is issued, or the two would
// interleave and the oldest page would not be all-denied.
async function waitForEntryCount(
  api: APIRequestContext,
  courseId: string,
  count: number,
  timeoutMs = 30_000,
): Promise<void> {
  const deadline = Date.now() + timeoutMs
  for (;;) {
    const page = await getCourseAuditLog(api, courseId, '?limit=200')
    if (page.entries.length >= count) return
    if (Date.now() > deadline) {
      throw new Error(
        `audit log for ${courseId} stalled at ${page.entries.length}/${count} entries`,
      )
    }
    await new Promise((resolve) => setTimeout(resolve, 250))
  }
}

// Produces `deniedCount` oldest denied entries followed by `successCount` newer
// successful ones, so the newest server page holds only successes and the
// oldest page only denials.
export async function generateCourseAuditEntries(
  admin: APIRequestContext,
  student: APIRequestContext,
  courseId: string,
  deniedCount: number,
  successCount: number,
): Promise<void> {
  await toggleArchive(student, courseId, deniedCount, 403)
  await waitForEntryCount(admin, courseId, deniedCount)

  await toggleArchive(admin, courseId, successCount, 200)
  await waitForEntryCount(admin, courseId, deniedCount + successCount)
}
