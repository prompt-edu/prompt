import { APIRequestContext, expect } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { test } from '../../src/fixtures/auth'
import { AuditLogPage } from '../../src/pages/AuditLogPage'
import { cleanupCourses, createCourseViaApi, uniqueSuffix } from '../courses/helpers'
import { generateCourseAuditEntries } from './helpers'

// One server page holds AUDIT_PAGE_SIZE entries, so 50 successes plus 5 older
// denials give a full newest page and a second page of exactly the denials.
const PAGE_SIZE = 50
const DENIED_COUNT = 5

test.describe('audit log: server-side paging', () => {
  test.use({ role: 'admin' })

  let admin: APIRequestContext
  let courseId: string | undefined

  test.beforeAll(async ({}, testInfo) => {
    admin = await apiContextFor('admin')
    const student = await apiContextFor('student')
    try {
      courseId = await createCourseViaApi(admin, {
        name: `AuditPaging${uniqueSuffix(testInfo.workerIndex)}`,
        semesterTag: 'ios2425',
        shortDescription: 'e2e audit log paging fixture',
      })
      await generateCourseAuditEntries(admin, student, courseId, DENIED_COUNT, PAGE_SIZE)
    } finally {
      await student.dispose()
    }
  })

  test.afterAll(async () => {
    await cleanupCourses(courseId)
    await admin?.dispose()
  })

  test('Newer/Older move through the log and change the rendered rows', async ({ page }) => {
    const auditLog = new AuditLogPage(page)
    await auditLog.gotoCourse(courseId!)
    await auditLog.expectLoaded()

    // Exactly one server page is rendered, and the table's own page size is a
    // constant its control can display (it used to be derived from the row
    // count, which left the control blank).
    await expect(auditLog.rows).toHaveCount(PAGE_SIZE)
    await expect(auditLog.rowsPerPage).toHaveText(String(PAGE_SIZE))
    await expect(auditLog.deniedRows).toHaveCount(0)
    await expect(auditLog.newer).toBeDisabled()
    await expect(auditLog.older).toBeEnabled()

    await auditLog.older.click()
    await expect(auditLog.rows).toHaveCount(DENIED_COUNT)
    await expect(auditLog.deniedRows).toHaveCount(DENIED_COUNT)
    await expect(auditLog.newer).toBeEnabled()
    await expect(auditLog.older).toBeDisabled()

    await auditLog.newer.click()
    await expect(auditLog.rows).toHaveCount(PAGE_SIZE)
    await expect(auditLog.deniedRows).toHaveCount(0)
  })

  test('changing a filter returns to the newest page', async ({ page }) => {
    const auditLog = new AuditLogPage(page)
    await auditLog.gotoCourse(courseId!)
    await auditLog.expectLoaded()
    await expect(auditLog.rows).toHaveCount(PAGE_SIZE)

    await auditLog.older.click()
    await expect(auditLog.rows).toHaveCount(DENIED_COUNT)

    await auditLog.selectOutcome('Success')
    await expect(auditLog.rows).toHaveCount(PAGE_SIZE)
    await expect(auditLog.deniedRows).toHaveCount(0)
    // The filtered log is a single page, so no navigation is offered at all. A
    // cursor left over from the previous page would still offer "Newer".
    await expect(auditLog.newer).toBeHidden()
  })
})
