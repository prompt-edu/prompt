import { test } from '../../src/fixtures/auth'
import { AuditLogPage } from '../../src/pages/AuditLogPage'
import { SEEDED_COURSES } from '../../src/data/constants'

test.describe('audit log: platform-wide (admin)', () => {
  test.use({ role: 'admin' })

  test('admin can open the global audit log', async ({ page }) => {
    const auditLog = new AuditLogPage(page)
    await auditLog.gotoGlobal()
    await auditLog.expectLoaded()
  })
})

test.describe('audit log: course-scoped (lecturer)', () => {
  test.use({ role: 'course-lecturer' })

  test('course lecturer can open their course audit log', async ({ page }) => {
    const auditLog = new AuditLogPage(page)
    await auditLog.gotoCourse(SEEDED_COURSES.fullCourse.id)
    await auditLog.expectLoaded()
  })
})

test.describe('audit log: access control', () => {
  test.use({ role: 'student' })

  test('a student is blocked from the global audit log', async ({ page }) => {
    const auditLog = new AuditLogPage(page)
    await auditLog.gotoGlobal()
    await auditLog.expectUnauthorized()
  })
})
