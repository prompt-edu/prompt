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

// Self Team Allocation phase on iPraktikum (follows the Application phase).
// The Keycloak users `student` / `student2` are enrolled participants.
export const SELF_TEAM_ALLOCATION_PHASE_ID = 'aaaa2222-0000-0000-0000-0000000000a2'

// Self Team Allocation phase on TestCourse with NO participants: requests by
// the e2e students must be rejected (negative auth fixture).
export const SELF_TEAM_ALLOCATION_FOREIGN_PHASE_ID = 'aaaa3333-0000-0000-0000-0000000000a3'

// Standalone Self Team Allocation phase owned by the lecturer-overview spec
// (no graph edge, navigate by URL): teams formed there never collide with the
// student journey's phase when spec files run in parallel workers.
export const SELF_TEAM_ALLOCATION_OVERVIEW_PHASE_ID = 'aaaa4444-0000-0000-0000-0000000000a4'

// Students matching the Keycloak users (see e2e/keycloak/realm.json attributes).
export const SEEDED_PHASE_STUDENTS = {
  student: { firstName: 'Stan', lastName: 'Stan' },
  student2: { firstName: 'Selma', lastName: 'Second' },
}

// The linear phase graph on SEEDED_COURSES.fullCourse, in order.
export const FULL_COURSE_PHASES = {
  application: { id: 'd0000001-0000-0000-0000-000000000001', type: 'Application' },
  interview: { id: 'd0000002-0000-0000-0000-000000000002', type: 'Interview' },
  matching: { id: 'd0000003-0000-0000-0000-000000000003', type: 'Matching' },
  teamAllocation: { id: 'd0000004-0000-0000-0000-000000000004', type: 'Team Allocation' },
  assessment: { id: 'd0000005-0000-0000-0000-000000000005', type: 'Assessment' },
}

// Standalone Assessment phases on fullCourse (no graph edges, so they are not
// in the course sidebar — navigate by URL). One phase per spec file, so release
// state and schema locking never leak between parallel Playwright files:
// the graph-tail assessment phase above is used by the lecturer journey + smoke,
// `visibility` by the student-visibility spec (Stan + Selma participate), and
// `selfEvaluation` by the self-evaluation spec (Stan participates).
export const ASSESSMENT_FIXTURE_PHASES = {
  visibility: 'd0000006-0000-0000-0000-000000000006',
  selfEvaluation: 'd0000007-0000-0000-0000-000000000007',
}

// Assessment phase on TestCourse with NO participants: requests by the e2e
// students must be rejected (negative auth fixture).
export const ASSESSMENT_FOREIGN_PHASE_ID = 'd0000008-0000-0000-0000-000000000008'

// Certificate phase appended to the tail of the iPraktikumFull graph (after
// Assessment). Left unconfigured, so it hosts the module-federation smoke test
// and the side-effect-free API-auth reads without colliding with the journeys.
export const CERTIFICATE_PHASES = {
  graphTail: 'd0000009-0000-0000-0000-000000000009',
}

// Standalone Certificate phases on fullCourse (no graph edges, navigate by
// URL). One phase per mutating spec file so template/release-date state and
// download recording never leak between parallel Playwright files:
// `lecturer` hosts the lecturer journey (template + release + participants),
// `student` the student self-download journey (Stan participates in both).
export const CERTIFICATE_FIXTURE_PHASES = {
  lecturer: 'd000000a-0000-0000-0000-00000000000a',
  student: 'd000000b-0000-0000-0000-00000000000b',
}

// Certificate phase on TestCourse with NO participants: requests by the e2e
// students must be rejected (negative auth fixture).
export const CERTIFICATE_FOREIGN_PHASE_ID = 'd000000c-0000-0000-0000-00000000000c'

// The student mapping to the Keycloak `student` role user (Stan); participates in
// every phase of fullCourse. Course access is DB-derived (matriculation + university
// login), not a Keycloak role. Same student row as SEEDED_PHASE_STUDENTS.student.
export const FULL_COURSE_STUDENT = {
  id: 'e0000005-0000-0000-0000-000000000005',
  courseParticipationId: 'a0000001-0000-0000-0000-000000000001',
  matriculationNumber: '00000005',
  universityLogin: 'no42tum',
}

// The student mapping to the Keycloak `student2` user (Selma); enrolled in
// fullCourse and participating in the assessment visibility fixture phase.
export const FULL_COURSE_STUDENT2 = {
  id: 'a5000007-0000-4000-8000-000000000007',
  courseParticipationId: 'ca000008-0000-4000-8000-000000000008',
  matriculationNumber: '00000007',
  universityLogin: 'st70two',
}

// Course-scoped Keycloak roles for fullCourse, assigned in e2e/keycloak/realm.json.
// `lecturer` + `course-lecturer` hold the Lecturer role; `course-editor` holds Editor.
export const FULL_COURSE_ROLES = {
  lecturer: 'ios2425-iPraktikumFull-Lecturer',
  editor: 'ios2425-iPraktikumFull-Editor',
}
