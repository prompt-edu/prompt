import { expect, test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import {
  ASSESSMENT_FIXTURE_PHASES,
  FULL_COURSE_STUDENT,
  FULL_COURSE_STUDENT2,
  SEEDED_COURSES,
  SEEDED_PHASE_STUDENTS,
} from '../../src/data/constants'
import {
  apiAsRole,
  buildCampusOnlineCsv,
  CAMPUS_ONLINE_HEADER,
  campusOnlineHeaderLine,
  createCategory,
  createCompetency,
  createSchema,
  encodeWindows1252,
  getAssessmentCategories,
  gradeCompetency,
  markAssessmentComplete,
  parseCampusOnlineCsv,
  putConfig,
  resetAssessmentPhase,
} from './helpers'

const PHASE_ID = ASSESSMENT_FIXTURE_PHASES.gradeExport
const SCHEMA_NAME = 'E2E Grade Export Rubric'
const COMPETENCY_NAME = 'Grade Export Competency'

const STAN_GRADE = 1.7
const SELMA_GRADE = 2.3
const UNKNOWN_REGISTRATION_NUMBER = '09999999'
const EDITING_NOTE = 'do not touch me'

const GRADE_INDEX = CAMPUS_ONLINE_HEADER.indexOf('GRADE')
const DATE_INDEX = CAMPUS_ONLINE_HEADER.indexOf('DATE_OF_ASSESSMENT')
const NOTES_INDEX = CAMPUS_ONLINE_HEADER.indexOf('EDITING_NOTES')
const REGISTRATION_INDEX = CAMPUS_ONLINE_HEADER.indexOf('REGISTRATION_NUMBER')

const PARTICIPATION_IDS = [
  FULL_COURSE_STUDENT.courseParticipationId,
  FULL_COURSE_STUDENT2.courseParticipationId,
]

const reset = () =>
  resetAssessmentPhase(PHASE_ID, {
    courseParticipationIds: PARTICIPATION_IDS,
    schemaNames: [SCHEMA_NAME],
  })

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer role.
test.use({ role: 'lecturer' })

test.describe('assessment: CampusOnline grade export', () => {
  test.beforeAll(async () => {
    await reset()
    const lecturer = await apiAsRole('lecturer')
    try {
      // Both seeded students are graded and finalized. Only Stan gets a row in
      // the uploaded CSV, so Selma exercises the "graded but absent from the
      // file" warning — the case where a grade would silently never be exported.
      const schema = await createSchema(lecturer, PHASE_ID, SCHEMA_NAME)
      await putConfig(lecturer, PHASE_ID, { assessmentSchemaId: schema.id })
      await createCategory(lecturer, PHASE_ID, schema.id, 'Grade Export Category')
      const category = (await getAssessmentCategories(lecturer, PHASE_ID))[0]
      await createCompetency(lecturer, PHASE_ID, category.id, COMPETENCY_NAME)
      const competency = (await getAssessmentCategories(lecturer, PHASE_ID))[0].competencies[0]

      for (const [participationId, grade] of [
        [FULL_COURSE_STUDENT.courseParticipationId, STAN_GRADE],
        [FULL_COURSE_STUDENT2.courseParticipationId, SELMA_GRADE],
      ] as const) {
        await gradeCompetency(lecturer, PHASE_ID, participationId, competency.id, 'good')
        await markAssessmentComplete(lecturer, PHASE_ID, participationId, grade, 'e2e')
      }
    } finally {
      await lecturer.dispose()
    }
  })

  test.afterAll(async () => {
    await reset()
  })

  test('a lecturer fills PROMPT grades into a CampusOnline CSV and downloads it', async ({
    page,
  }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()

    const csv = buildCampusOnlineCsv([
      {
        // Padded with an extra leading zero: matching must tolerate that.
        REGISTRATION_NUMBER: `0${FULL_COURSE_STUDENT.matriculationNumber}`,
        FAMILY_NAME_OF_STUDENT: SEEDED_PHASE_STUDENTS.student.lastName,
        FIRST_NAME_OF_STUDENT: SEEDED_PHASE_STUDENTS.student.firstName,
        COURSE_TITLE: 'iPraktikum; "Advanced"',
        EDITING_NOTES: EDITING_NOTE,
      },
      {
        REGISTRATION_NUMBER: UNKNOWN_REGISTRATION_NUMBER,
        FAMILY_NAME_OF_STUDENT: 'Nobody',
        FIRST_NAME_OF_STUDENT: 'Nina',
      },
    ])

    await phase.uploadCampusCsv('exam-list.csv', Buffer.from(csv, 'utf-8'))

    await phase.expectGradeFilled(FULL_COURSE_STUDENT.matriculationNumber, STAN_GRADE.toFixed(1))
    await phase.expectSkippedForReason(
      'No matching student in this phase',
      UNKNOWN_REGISTRATION_NUMBER,
    )
    await phase.expectGradedStudentMissingWarning(SEEDED_PHASE_STUDENTS.student2.lastName)

    const downloaded = (await phase.downloadFilledCampusCsv()).toString('utf-8')
    const rows = parseCampusOnlineCsv(downloaded)

    // The header must come back exactly as CampusOnline wrote it.
    expect(downloaded.split('\r\n')[0]).toBe(campusOnlineHeaderLine())
    expect(rows[0]).toEqual([...CAMPUS_ONLINE_HEADER])
    expect(rows).toHaveLength(3)

    const stanRow = rows.find((row) =>
      row[REGISTRATION_INDEX].endsWith(FULL_COURSE_STUDENT.matriculationNumber),
    )
    expect(stanRow, 'row of the graded student').toBeDefined()
    expect(stanRow![GRADE_INDEX]).toBe(STAN_GRADE.toFixed(1))
    // DD.MM.YYYY, and the assessment was completed in this test run.
    expect(stanRow![DATE_INDEX]).toMatch(/^\d{2}\.\d{2}\.\d{4}$/)
    // Pass-through columns survive untouched, including a quoted delimiter.
    expect(stanRow![NOTES_INDEX]).toBe(EDITING_NOTE)
    expect(stanRow![CAMPUS_ONLINE_HEADER.indexOf('COURSE_TITLE')]).toBe('iPraktikum; "Advanced"')
    // The registration number itself is never rewritten.
    expect(stanRow![REGISTRATION_INDEX]).toBe(`0${FULL_COURSE_STUDENT.matriculationNumber}`)

    const unknownRow = rows.find(
      (row) => row[REGISTRATION_INDEX] === UNKNOWN_REGISTRATION_NUMBER,
    )
    expect(unknownRow, 'row of the unknown student').toBeDefined()
    expect(unknownRow![GRADE_INDEX]).toBe('')
    expect(unknownRow![DATE_INDEX]).toBe('')
  })

  test('a Windows-1252 encoded CSV keeps its encoding and umlauts', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()

    const csv = buildCampusOnlineCsv([
      {
        REGISTRATION_NUMBER: FULL_COURSE_STUDENT.matriculationNumber,
        FAMILY_NAME_OF_STUDENT: SEEDED_PHASE_STUDENTS.student.lastName,
        FIRST_NAME_OF_STUDENT: SEEDED_PHASE_STUDENTS.student.firstName,
      },
      {
        REGISTRATION_NUMBER: UNKNOWN_REGISTRATION_NUMBER,
        FAMILY_NAME_OF_STUDENT: 'Müller',
        FIRST_NAME_OF_STUDENT: 'Jörg',
      },
    ])

    await phase.uploadCampusCsv('exam-list-latin1.csv', encodeWindows1252(csv))
    await phase.expectGradeFilled(FULL_COURSE_STUDENT.matriculationNumber, STAN_GRADE.toFixed(1))

    const bytes = await phase.downloadFilledCampusCsv()

    // "ü" must be the single Windows-1252 byte 0xFC, not the UTF-8 pair C3 BC:
    // re-encoding as UTF-8 would corrupt every umlaut for the CampusOnline import.
    expect(bytes.includes(0xfc), 'windows-1252 encoded umlaut').toBe(true)
    expect(bytes.includes(Buffer.from([0xc3, 0xbc])), 'utf-8 encoded umlaut').toBe(false)

    const rows = parseCampusOnlineCsv(bytes.toString('latin1'))
    const stanRow = rows.find(
      (row) => row[REGISTRATION_INDEX] === FULL_COURSE_STUDENT.matriculationNumber,
    )
    expect(stanRow![GRADE_INDEX]).toBe(STAN_GRADE.toFixed(1))

    const umlautRow = rows.find((row) => row[REGISTRATION_INDEX] === UNKNOWN_REGISTRATION_NUMBER)
    expect(umlautRow![CAMPUS_ONLINE_HEADER.indexOf('FAMILY_NAME_OF_STUDENT')]).toBe('Müller')
    expect(umlautRow![CAMPUS_ONLINE_HEADER.indexOf('FIRST_NAME_OF_STUDENT')]).toBe('Jörg')
  })

  test('a file that is not a CampusOnline export is rejected', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PHASE_ID, '/settings')
    await phase.expectSettingsLoaded()

    const wrongCsv = '"First name";"Last name";"Student number";"Rank"\r\n"Stan";"Stan";"5";"1"\r\n'
    await phase.uploadCampusCsv('matching-export.csv', Buffer.from(wrongCsv, 'utf-8'))

    await expect(
      phase.gradeExportCard().getByText('This does not look like a CampusOnline grade export.', {
        exact: false,
      }),
    ).toBeVisible({ timeout: 15_000 })
    await expect(
      phase.gradeExportCard().getByRole('button', { name: 'Download Filled CSV' }),
    ).toBeDisabled()
  })
})
