import { test } from '../../src/fixtures/auth'
import { CertificatePage } from '../../src/pages/CertificatePage'
import {
  CERTIFICATE_FIXTURE_PHASES,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
} from '../../src/data/constants'
import { E2E_TEMPLATE, resetCertificatePhase } from './helpers'

const PHASE_ID = CERTIFICATE_FIXTURE_PHASES.lecturer
const STUDENT_NAME = `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer role.
test.use({ role: 'lecturer' })

test.describe('certificate: lecturer journey', () => {
  test.afterAll(async () => {
    await resetCertificatePhase(PHASE_ID)
  })

  test('a lecturer configures the template, previews it, and releases it', async ({ page }) => {
    const phase = new CertificatePage(page)

    // Configure the certificate template.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()
    await phase.pasteTemplate(E2E_TEMPLATE)
    await phase.saveTemplate()

    // Trigger generation: preview compiles the template (no error), and
    // Release Now makes the certificate available to students.
    await phase.testCertificate()
    await phase.releaseNow()

    // The participants table shows the seeded student and their download status.
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/participants')
    await phase.expectParticipantsLoaded()
    await phase.expectParticipantRow(STUDENT_NAME, 'Not downloaded')
  })
})
