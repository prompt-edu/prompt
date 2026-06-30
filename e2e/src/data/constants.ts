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
}

// Total non-template courses present in the seed.
export const SEEDED_COURSE_COUNT = 3

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
