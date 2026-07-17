import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'

// Thin core-API primitives for the course-lifecycle write-path suites. Used from
// beforeAll/afterAll to set up fixtures and clean up (mirrors the per-area
// helpers.ts pattern, e.g. tests/interview/helpers.ts).

export interface CourseSummary {
  id: string
  name: string
  semesterTag: string | null
  template: boolean
}

export interface CoursePhaseSequence {
  id: string
  name: string
  isInitialPhase: boolean
  coursePhaseTypeID: string
  coursePhaseType: string
}

export interface CourseWithPhases {
  id: string
  name: string
  semesterTag: string
  shortDescription: string
  longDescription: string
  coursePhases: CoursePhaseSequence[]
}

export interface PhaseGraphEdge {
  fromCoursePhaseID: string
  toCoursePhaseID: string
}

export interface CreateCourseOptions {
  name: string
  semesterTag: string
  template?: boolean
  courseType?: string
  ects?: number
  shortDescription?: string
}

// Mirrors courseDTO.CreateCourse. The dates are placeholders — the lifecycle
// suites never assert on them. Returns the new course id.
export async function createCourseViaApi(
  api: APIRequestContext,
  opts: CreateCourseOptions,
): Promise<string> {
  const res = await api.post('/api/courses/', {
    data: {
      name: opts.name,
      startDate: '2025-04-01',
      endDate: '2025-09-30',
      semesterTag: opts.semesterTag,
      restrictedData: {},
      studentReadableData: { icon: 'graduation-cap', 'bg-color': 'bg-blue-100' },
      shortDescription: opts.shortDescription ?? 'e2e lifecycle fixture course',
      longDescription: '',
      courseType: opts.courseType ?? 'practical course',
      ects: opts.ects ?? 10,
      template: opts.template ?? false,
    },
  })
  if (!res.ok()) {
    throw new Error(`POST course failed: ${res.status()} ${await res.text()}`)
  }
  return ((await res.json()) as { id: string }).id
}

export async function deleteCourseViaApi(api: APIRequestContext, id: string): Promise<void> {
  const res = await api.delete(`/api/courses/${id}`)
  if (!res.ok() && res.status() !== 404) {
    throw new Error(`DELETE course failed: ${res.status()} ${await res.text()}`)
  }
}

// Best-effort delete as admin, for afterAll teardown: mints its own admin context
// so it works even when the failing test never obtained one.
export async function cleanupCourses(...ids: (string | undefined)[]): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    for (const id of ids) {
      if (id) await deleteCourseViaApi(admin, id)
    }
  } finally {
    await admin.dispose()
  }
}

export async function listCourses(api: APIRequestContext): Promise<CourseSummary[]> {
  const res = await api.get('/api/courses/')
  if (!res.ok()) throw new Error(`GET courses failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as CourseSummary[]
}

export async function listTemplates(api: APIRequestContext): Promise<CourseSummary[]> {
  const res = await api.get('/api/courses/template')
  if (!res.ok()) throw new Error(`GET templates failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as CourseSummary[]
}

export async function getCourseById(
  api: APIRequestContext,
  id: string,
): Promise<CourseWithPhases> {
  const res = await api.get(`/api/courses/${id}`)
  if (!res.ok()) throw new Error(`GET course failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as CourseWithPhases
}

export async function getCoursePhaseGraph(
  api: APIRequestContext,
  id: string,
): Promise<PhaseGraphEdge[]> {
  const res = await api.get(`/api/courses/${id}/phase_graph`)
  if (!res.ok()) throw new Error(`GET phase_graph failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as PhaseGraphEdge[]
}

export interface StaffMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
}

export interface CourseStaff {
  lecturers: StaffMember[]
  editors: StaffMember[]
}

export async function getCourseStaff(
  api: APIRequestContext,
  id: string,
): Promise<CourseStaff> {
  const res = await api.get(`/api/keycloak/${id}/group/staff`)
  if (!res.ok()) throw new Error(`GET staff failed: ${res.status()} ${await res.text()}`)
  return (await res.json()) as CourseStaff
}

// Unique, hyphen-free identifiers per test run. workerIndex keeps parallel local
// files apart; the random suffix survives CI retries against the reused DB. No
// hyphens: the course-list role parser splits roles on '-'.
export function uniqueSuffix(workerIndex: number): string {
  const rand = Math.random().toString(36).slice(2, 8)
  return `w${workerIndex}${rand}`
}
