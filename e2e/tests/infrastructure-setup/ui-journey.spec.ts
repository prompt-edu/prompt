import { expect } from '@playwright/test'
import { test } from '../../src/fixtures/auth'
import {
  FULL_COURSE_STUDENT,
  INFRASTRUCTURE_SETUP_PHASE_ID,
  SEEDED_COURSES,
  SEEDED_STUDENT,
} from '../../src/data/constants'
import { InfrastructureSetupPage } from '../../src/pages/InfrastructureSetupPage'
import { KEYCLOAK_PROVIDER, resetInfrastructureSetupPhase } from './helpers'

const COURSE_ID = SEEDED_COURSES.fullCourse.id
const PHASE_ID = INFRASTRUCTURE_SETUP_PHASE_ID

// The phase driven through its pages, against the stack's own Keycloak: its two
// participants are Stan, who has a Keycloak account, and Niclas, who does not. One
// run therefore ends with a group Stan was added to and one Niclas was not, which is
// what the participants table and Stan's own page have to tell apart.
test.describe.serial('infrastructure setup: ui journey', () => {
  test.beforeAll(async () => {
    await resetInfrastructureSetupPhase(PHASE_ID)
  })

  test.afterAll(async () => {
    await resetInfrastructureSetupPhase(PHASE_ID)
  })

  test.describe('lecturer', () => {
    test.use({ role: 'course-lecturer' })

    test('configures the phase, provisions it and sees who got access', async ({ page }) => {
      test.setTimeout(120_000)
      const phase = new InfrastructureSetupPage(page)

      // An unconfigured phase says what is missing and cannot be started.
      await phase.goto(COURSE_ID, PHASE_ID, '/provisioning')
      await expect(page.getByText('Not ready to provision')).toBeVisible({ timeout: 15_000 })
      await expect(page.getByText('No provider is configured yet.')).toBeVisible()
      await expect(phase.provisionButton()).toBeDisabled()

      await page.getByRole('link', { name: 'Add a provider' }).click()
      await expect(page).toHaveURL(/\/configuration#providers$/)

      // Every save reports back through a toast.
      await phase.saveSemesterTag('e2eui')
      await phase.expectToast('Semester tag saved')

      await phase.addKeycloakProvider(KEYCLOAK_PROVIDER)
      await phase.expectToast('Provider added')

      await phase.addPerStudentResource('{{semesterTag}}-{{studentLogin}}')
      await phase.expectToast('Resource added')

      // The dry run resolves both participants and says what a click would do.
      await phase.goto(COURSE_ID, PHASE_ID, '/provisioning')
      await expect(page.getByText('Ready to provision', { exact: true })).toBeVisible({
        timeout: 15_000,
      })
      await expect(page.getByText('2 students')).toBeVisible()
      await expect(page.getByText('The next run creates 2.')).toBeVisible()

      await phase.provisionButton().click()
      await phase.expectToast('Provisioning started')

      await expect(phase.statusFilter('1 created')).toBeVisible({ timeout: 60_000 })
      await expect(phase.statusFilter('1 partial')).toBeVisible()
      // Only Niclas's partial group is left to retry.
      await expect(page.getByText(/The next run retries 1/)).toBeVisible({ timeout: 15_000 })

      await phase.goto(COURSE_ID, PHASE_ID, '/participants')
      await expect(page.getByRole('columnheader', { name: 'Keycloak group' })).toBeVisible({
        timeout: 15_000,
      })
      await expect(page.getByText('1 of 2 participants can reach every resource.')).toBeVisible()
      await expect(
        page.getByRole('row').filter({ hasText: FULL_COURSE_STUDENT.email }).getByText('Ready'),
      ).toBeVisible()
      await expect(
        page.getByRole('row').filter({ hasText: SEEDED_STUDENT.email }).getByText('Not added yet'),
      ).toBeVisible()
    })
  })

  test.describe('student', () => {
    test.use({ role: 'student' })

    test('sees the group provisioned for them', async ({ page }) => {
      const phase = new InfrastructureSetupPage(page)
      await phase.goto(COURSE_ID, PHASE_ID)
      await phase.expectLoaded()

      await expect(page.getByText('You are not a student of this course.')).toBeHidden()
      await expect(page.getByText('Keycloak group')).toBeVisible()
      await expect(page.getByText('Ready', { exact: true })).toBeVisible()
      // A Keycloak group only has an admin console page, which a student cannot open.
      await expect(page.getByRole('link', { name: 'Open' })).toBeHidden()
    })
  })
})
