import { test } from '../../src/fixtures/auth'
import { StudentsPage } from '../../src/pages/StudentsPage'
import { SEEDED_STUDENT } from '../../src/data/constants'

test.describe('student list access control', () => {
  test.describe('as an admin', () => {
    test.use({ role: 'admin' })

    test('the student list loads and shows seeded students', async ({
      page,
    }) => {
      const students = new StudentsPage(page)
      await students.goto()
      await students.expectLoaded()
      await page
        .getByText(SEEDED_STUDENT.lastName, { exact: false })
        .first()
        .waitFor()
    })
  })

  test.describe('as a student', () => {
    test.use({ role: 'student' })

    test('the student list is blocked', async ({ page }) => {
      const students = new StudentsPage(page)
      await students.goto()
      await students.expectBlocked()
    })
  })
})
