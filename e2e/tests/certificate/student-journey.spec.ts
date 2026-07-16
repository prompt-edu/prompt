import { expect, test } from '../../src/fixtures/auth'
import { CertificatePage } from '../../src/pages/CertificatePage'
import { CERTIFICATE_FIXTURE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { E2E_TEMPLATE, putTemplate, resetCertificatePhase, setReleaseDate } from './helpers'

const PHASE_ID = CERTIFICATE_FIXTURE_PHASES.student

// The `student` user (Stan) participates in this phase; course access is
// DB-derived from the matriculation/login token claims.
test.use({ role: 'student' })

test.describe('certificate: student journey', () => {
  test.beforeAll(async () => {
    // Lecturer sets up a configured but not-yet-released certificate.
    await putTemplate(PHASE_ID, E2E_TEMPLATE)
    await setReleaseDate(PHASE_ID, new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString())
  })

  test.afterAll(async () => {
    await resetCertificatePhase(PHASE_ID)
  })

  test('a student downloads their certificate once it is released', async ({ page }) => {
    const phase = new CertificatePage(page)

    // Before release the certificate is gated.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()
    await phase.expectNotAvailable(/will be available after/i)

    // The lecturer releases the certificate (release date in the past).
    await setReleaseDate(PHASE_ID, new Date(Date.now() - 60 * 1000).toISOString())

    // The student can now download the generated PDF.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectOverviewLoaded()
    const download = await phase.downloadOwnCertificate()
    expect(download.suggestedFilename()).toBe('certificate.pdf')

    // The download is recorded and surfaced on the overview.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID)
    await phase.expectLastDownloaded()
  })
})
