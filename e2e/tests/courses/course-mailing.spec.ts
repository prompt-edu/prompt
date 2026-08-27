import { SEEDED_COURSES } from '../../src/data/constants'
import { clearMailpit, getMailpitMessages } from '../../src/data/mailpit'
import { expect, test } from '../../src/fixtures/auth'
import { CourseMailingPage } from '../../src/pages/CourseMailingPage'

test.describe('course mailing campaigns', () => {
  test.describe('as an admin', () => {
    test.use({ role: 'admin' })

    test.beforeEach(async () => {
      await clearMailpit()
    })

    test('create, save, and send a campaign; the mail is delivered', async ({ page }) => {
      const mailing = new CourseMailingPage(page, SEEDED_COURSES.fullCourse.id)
      await mailing.goto()
      await mailing.expectLoaded()

      await mailing.startNewCampaign()

      // Unique subject so the mailpit assertion is isolated from other messages.
      const marker = `E2E Campaign ${Date.now()}`
      await mailing.fillDetails({
        name: marker,
        phaseName: 'Assessment',
        // "All participants" is resilient to other specs mutating pass_status.
        status: 'All participants',
        subject: marker,
        body: 'Hello {{firstName}}, this is an automated e2e message.',
        replyToEmail: 'reply-e2e@example.com',
      })

      await mailing.save()
      await mailing.showRecipients()
      await mailing.sendWithConfirmation()

      // Back on the overview once the send is accepted.
      await expect(page).toHaveURL(/\/mailing$/)

      // The rendered mail lands in mailpit.
      await expect
        .poll(
          async () => (await getMailpitMessages()).filter((m) => m.Subject === marker).length,
          { timeout: 20_000 },
        )
        .toBeGreaterThan(0)
    })
  })

  test.describe('as a student', () => {
    test.use({ role: 'student' })

    test('the mailing page is not accessible', async ({ page }) => {
      const mailing = new CourseMailingPage(page, SEEDED_COURSES.fullCourse.id)
      await mailing.goto()
      await expect(mailing.newMailButton).toHaveCount(0)
    })
  })
})
