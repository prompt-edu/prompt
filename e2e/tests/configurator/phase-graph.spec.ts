import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CourseConfiguratorPage } from '../../src/pages/CourseConfiguratorPage'
import {
  cleanupCourses,
  createCourseViaApi,
  getCourseById,
  getCoursePhaseGraph,
  uniqueSuffix,
} from '../courses/helpers'

test.use({ role: 'admin' })

test.describe('phase-graph editing', () => {
  let courseId: string

  test.beforeAll(async ({}, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    const admin = await apiContextFor('admin')
    try {
      courseId = await createCourseViaApi(admin, {
        name: `E2EGraph${suffix}`,
        semesterTag: `e2egraph${suffix}`,
      })
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('admin adds, connects, and removes phases in the DAG editor', async ({ page }) => {
    const configurator = new CourseConfiguratorPage(page, courseId)
    await configurator.goto()

    // Application is the only initial-phase type, so it must be added first.
    const applicationId = await configurator.addPhase('Application')
    const matchingId = await configurator.addPhase('Matching')
    await configurator.connect(applicationId, matchingId)
    await configurator.save()

    // Assert the persisted graph semantically (real uuids, not temp ids).
    const api = await apiContextFor('admin')
    try {
      const course = await getCourseById(api, courseId)
      const app = course.coursePhases.find((p) => p.coursePhaseType === 'Application')
      const matching = course.coursePhases.find((p) => p.coursePhaseType === 'Matching')
      expect(app, 'Application phase persisted').toBeTruthy()
      expect(matching, 'Matching phase persisted').toBeTruthy()

      const graph = await getCoursePhaseGraph(api, courseId)
      expect(graph).toHaveLength(1)
      expect(graph[0].fromCoursePhaseID).toBe(app!.id)
      expect(graph[0].toCoursePhaseID).toBe(matching!.id)
    } finally {
      await api.dispose()
    }

    // Remove the non-initial phase; the phase and its edge disappear.
    await configurator.removePhase('Matching')
    await configurator.save()

    const api2 = await apiContextFor('admin')
    try {
      const course = await getCourseById(api2, courseId)
      expect(course.coursePhases.some((p) => p.coursePhaseType === 'Matching')).toBeFalsy()
      const graph = await getCoursePhaseGraph(api2, courseId)
      expect(graph).toHaveLength(0)
    } finally {
      await api2.dispose()
    }
  })
})
