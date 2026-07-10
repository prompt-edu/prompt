import { test } from '../../src/fixtures/auth'
import { CertificatePage } from '../../src/pages/CertificatePage'
import { CERTIFICATE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// certificate_component REMOTE, loaded by the core shell from
// /certificate/remoteEntry.js through the e2e nginx proxy. If the remote fails
// to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('certificate: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new CertificatePage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, CERTIFICATE_PHASES.graphTail)
    await phase.expectOverviewLoaded()
  })
})
