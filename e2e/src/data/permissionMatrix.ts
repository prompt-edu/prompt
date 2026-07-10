// Data-driven role x surface authorization matrix. Adding a surface is one entry
// in SURFACES; the browser and API specs both derive their cases from it.
//
// Authorization model (servers/core/permissionValidation): denials return 403,
// missing/invalid token returns 401, and there is no implicit admin bypass -
// admin only passes where PromptAdmin is an explicitly allowed role. The frontend
// PermissionRestriction renders the shared UnauthorizedPage ("Access Denied") when
// the current role lacks the route's permissions for the current course.

import { Role } from './roles'
import {
  SEEDED_COURSES,
  SELF_TEAM_ALLOCATION_PHASE_ID,
  FULL_COURSE_PHASES,
} from './constants'

// The five task roles (student2 is an extra journey user, not part of the matrix).
export const MATRIX_ROLES: Role[] = [
  'admin',
  'lecturer',
  'course-lecturer',
  'course-editor',
  'student',
]

// Course that holds every course-scoped role in the seed, so allowed cases resolve.
export const PRIMARY_COURSE = SEEDED_COURSES.fullCourse

// fullCourse's Application phase (mailing) and an iPraktikum phase the student is
// enrolled in (student self-participation).
const MAILING_PHASE_ID = FULL_COURSE_PHASES.application.id
const STUDENT_PHASE_ID = SELF_TEAM_ALLOCATION_PHASE_ID

export interface Surface {
  name: string
  browser?: {
    path: (courseId: string) => string
    heading: string // the page <h1> rendered when authorized
    allowed: Role[]
  }
  api?: {
    method?: 'GET' | 'PUT' // default GET
    path: (courseId: string) => string
    allowed: Role[]
    allowedStatus?: number // default: any 2xx; mailing probe -> 400
    invalidBody?: string // malformed body for the allowed-side probe (mailing)
    blockedReason?: string // appended to blocked test titles for clarity
  }
}

export const SURFACES: Surface[] = [
  {
    name: 'courses list',
    browser: {
      path: () => '/management/courses',
      heading: 'Courses',
      allowed: ['admin', 'lecturer', 'course-lecturer', 'course-editor', 'student'],
    },
    api: {
      path: () => '/api/courses/',
      allowed: ['admin', 'lecturer', 'course-lecturer', 'course-editor', 'student'],
    },
  },
  {
    name: 'student list',
    browser: {
      path: () => '/management/students',
      heading: 'Students',
      allowed: ['admin', 'lecturer'],
    },
    api: {
      path: () => '/api/students/',
      allowed: ['admin', 'lecturer'],
    },
  },
  {
    name: 'course settings',
    browser: {
      path: (courseId) => `/management/course/${courseId}/settings`,
      heading: 'Settings',
      allowed: ['admin', 'lecturer', 'course-lecturer'],
    },
  },
  {
    // API read of course data - editors may read it even though they cannot open
    // the Settings page (page guard is Admin+Lecturer, API read also allows Editor).
    name: 'course detail',
    api: {
      path: (courseId) => `/api/courses/${courseId}`,
      allowed: ['admin', 'lecturer', 'course-lecturer', 'course-editor'],
    },
  },
  {
    name: 'phase config',
    browser: {
      path: (courseId) => `/management/course/${courseId}/configurator`,
      heading: 'Course Configurator',
      allowed: ['admin', 'lecturer', 'course-lecturer', 'course-editor'],
    },
    api: {
      path: (courseId) => `/api/courses/${courseId}/phase_graph`,
      allowed: ['admin', 'lecturer', 'course-lecturer', 'course-editor'],
    },
  },
  {
    name: 'mailing',
    browser: {
      path: (courseId) => `/management/course/${courseId}/${MAILING_PHASE_ID}/mailing`,
      heading: 'Application Mailing Settings',
      allowed: ['admin', 'lecturer', 'course-lecturer'],
    },
    api: {
      // A malformed body proves auth passed (400 after BindJSON) without sending
      // mail; blocked roles never reach BindJSON (403 in the middleware first).
      method: 'PUT',
      path: () => `/api/mailing/${MAILING_PHASE_ID}`,
      allowed: ['admin', 'lecturer', 'course-lecturer'],
      allowedStatus: 400,
      invalidBody: 'not-json',
    },
  },
  {
    name: 'student self-participation',
    api: {
      path: () => `/api/course_phases/${STUDENT_PHASE_ID}/participations/self`,
      allowed: ['student'],
      blockedReason: 'no implicit admin bypass',
    },
  },
]

// A role scoped to PRIMARY_COURSE must be denied on courses it has no role in.
// Only surfaces the role is *allowed* on for PRIMARY_COURSE are listed, so each
// case is a genuine isolation signal rather than a blanket denial.
//
// Caveat: keep cross-course surfaces to DIRECT guarded routes (course settings /
// phase config). Phase routes (`/:courseId/:phaseId/*`) resolve the phase through
// PhaseRouterMapping first and render "Phase not found" for an inaccessible course
// before PermissionRestriction runs - so a phase/mailing cross-course row would
// need a different assertion.
export interface CrossCourseCase {
  role: Role
  course: { id: string; name: string }
  surfaces: string[]
}

export const CROSS_COURSE: CrossCourseCase[] = [
  { role: 'course-editor', course: SEEDED_COURSES.iPraktikum, surfaces: ['phase config'] },
  { role: 'course-editor', course: SEEDED_COURSES.testCourse, surfaces: ['course detail', 'phase config'] },
  { role: 'course-lecturer', course: SEEDED_COURSES.testCourse, surfaces: ['course settings', 'course detail'] },
]
