import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import { CoursesPage } from '../../src/pages/CoursesPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// assessment_component REMOTE, loaded by the core shell from
// /assessment/remoteEntry.js through the e2e nginx proxy. If the remote fails
// to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('assessment: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.assessment.id)
    await phase.expectOverviewLoaded()
  })

  // Regression guard for #2086: the remote used to ship its own Tailwind build,
  // which style-loader appended after core's. Its .px-3 then beat core's .pl-9
  // on the shared Input, and the course search icon overlapped the placeholder.
  test('the loaded remote leaves the shell utility classes alone', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.assessment.id)
    await phase.expectOverviewLoaded()

    await page.evaluate(() => {
      Object.assign(window, { __e2eSameDocument: true })
    })

    const courses = new CoursesPage(page)
    await courses.gotoFromSidebar()
    await courses.expectLoaded()

    // A full reload would drop anything the remote injected and turn the
    // assertions below green for the wrong reason.
    expect(await page.evaluate(() => '__e2eSameDocument' in window)).toBe(true)

    const search = courses.searchInput()
    // Plain pl-9, not pl-9!: the important modifier was the workaround.
    await expect(search).toHaveClass(/(^|\s)pl-9(\s|$)/)
    await expect(search).toHaveCSS('padding-left', '36px')
  })
})
