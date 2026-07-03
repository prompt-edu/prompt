// Known values from the seed dump (e2e/seed/e2e_seed.sql). When the seed
// changes, update these. IDs survive the migrate-up upgrade applied on startup.

export const SEEDED_COURSES = {
  iPraktikum: {
    id: 'd7307be2-d3dc-496e-86f0-643bff6cc1c8',
    name: 'iPraktikum',
    semesterTag: 'ios2425',
  },
  iPraktikumTest: {
    id: 'e12ffe63-448d-4469-a840-1699e9b328d1',
    name: 'iPraktikum-Test',
    semesterTag: 'ios2425',
  },
  testCourse: {
    id: 'be780b32-a678-4b79-ae1c-80071771d254',
    name: 'TestCourse',
    semesterTag: 'ios2425',
  },
  // A course with a full phase graph, participations, and course-scoped roles.
  // Name is hyphen-free so the split_part('-') role parsing in the course-list
  // query resolves it correctly.
  fullCourse: {
    id: 'c0000001-0000-0000-0000-000000000001',
    name: 'iPraktikumFull',
    semesterTag: 'ios2425',
  },
}

// Total non-template courses present in the seed.
export const SEEDED_COURSE_COUNT = 4

// An open Application phase on iPraktikum (applicationEndDate in the far future),
// so the application file-upload endpoints accept uploads.
export const OPEN_APPLICATION_PHASE_ID = 'aaaa1111-0000-0000-0000-0000000000a1'

export const SEEDED_STUDENT = {
  id: '3869f209-9a21-4595-ae0e-bc6d6a3e2d63',
  firstName: 'Niclas',
  lastName: 'Heun',
  email: 'niclas.heun@tum.de',
  matriculationNumber: '03711126',
}

// The linear phase graph on SEEDED_COURSES.fullCourse, in order.
export const FULL_COURSE_PHASES = {
  application: { id: 'd0000001-0000-0000-0000-000000000001', type: 'Application' },
  interview: { id: 'd0000002-0000-0000-0000-000000000002', type: 'Interview' },
  matching: { id: 'd0000003-0000-0000-0000-000000000003', type: 'Matching' },
  teamAllocation: { id: 'd0000004-0000-0000-0000-000000000004', type: 'Team Allocation' },
  assessment: { id: 'd0000005-0000-0000-0000-000000000005', type: 'Assessment' },
}

// The student mapping to the Keycloak `student` role user; participates in every
// phase of fullCourse. Course access is DB-derived (matriculation + university login),
// not a Keycloak role.
export const FULL_COURSE_STUDENT = {
  id: 'e0000005-0000-0000-0000-000000000005',
  courseParticipationId: 'a0000001-0000-0000-0000-000000000001',
  matriculationNumber: '00000005',
  universityLogin: 'no42tum',
}

// Course-scoped Keycloak roles for fullCourse, assigned in e2e/keycloak/realm.json.
// `lecturer` + `course-lecturer` hold the Lecturer role; `course-editor` holds Editor.
export const FULL_COURSE_ROLES = {
  lecturer: 'ios2425-iPraktikumFull-Lecturer',
  editor: 'ios2425-iPraktikumFull-Editor',
}
