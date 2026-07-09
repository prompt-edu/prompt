import { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { Role } from '../../src/data/roles'
import { ASSESSMENT_API, BASE_URL } from '../../src/env'

// All assessment API calls go through the client-core nginx proxy on the
// browser origin (same path prefix as prod Traefik), NOT the core API.
export function assessmentUrl(phaseId: string, path: string): string {
  return `${BASE_URL}${ASSESSMENT_API}/course_phase/${phaseId}/${path}`
}

// Default schemas created by the assessment server's migrations on its fresh DB.
export const DEFAULT_SCHEMAS = {
  assessment: 'Assessment Template',
  self: 'Self Evaluation Template',
  peer: 'Peer Evaluation Template',
  tutor: 'Tutor Evaluation Template',
}

export interface AssessmentSchema {
  id: string
  name: string
}

export interface CategoryWithCompetencies {
  id: string
  name: string
  competencies: { id: string; name: string }[]
}

export interface CoursePhaseConfig {
  coursePhaseID: string
  assessmentSchemaID: string
  selfEvaluationSchema: string
  peerEvaluationSchema: string
  tutorEvaluationSchema: string
  resultsReleased: boolean
}

async function ok(resPromise: ReturnType<APIRequestContext['get']>, what: string) {
  const res = await resPromise
  if (!res.ok()) {
    throw new Error(`${what} failed: ${res.status()} ${await res.text()}`)
  }
  return res
}

export async function getSchemas(
  api: APIRequestContext,
  phaseId: string,
): Promise<AssessmentSchema[]> {
  const res = await ok(api.get(assessmentUrl(phaseId, 'assessment-schema')), 'GET schemas')
  return (await res.json()) as AssessmentSchema[]
}

export async function getSchemaByName(
  api: APIRequestContext,
  phaseId: string,
  name: string,
): Promise<AssessmentSchema | undefined> {
  return (await getSchemas(api, phaseId)).find((schema) => schema.name === name)
}

export async function createSchema(
  api: APIRequestContext,
  phaseId: string,
  name: string,
): Promise<AssessmentSchema> {
  await ok(
    api.post(assessmentUrl(phaseId, 'assessment-schema'), {
      data: { name, description: `${name} (created by e2e)` },
    }),
    'POST schema',
  )
  const schema = await getSchemaByName(api, phaseId, name)
  if (!schema) throw new Error(`schema "${name}" not found after creation`)
  return schema
}

// GET /config lazily creates the default config row bound to the default
// template schemas, so this is also the setup primitive for a fresh phase.
export async function getConfig(
  api: APIRequestContext,
  phaseId: string,
): Promise<CoursePhaseConfig> {
  const res = await ok(api.get(assessmentUrl(phaseId, 'config')), 'GET config')
  return (await res.json()) as CoursePhaseConfig
}

export interface ConfigOverrides {
  assessmentSchemaId?: string
  selfEvaluationEnabled?: boolean
  selfEvaluationSchema?: string
}

// PUT /config requires the full request body; start/deadline are set to a
// generous open window so grading, completion, and (cleanup) unmarking work.
export async function putConfig(
  api: APIRequestContext,
  phaseId: string,
  overrides: ConfigOverrides = {},
): Promise<void> {
  const current = await getConfig(api, phaseId)
  const past = new Date(Date.now() - 24 * 3600 * 1000).toISOString()
  const future = new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString()
  await ok(
    api.put(assessmentUrl(phaseId, 'config'), {
      data: {
        coursePhaseId: phaseId,
        assessmentSchemaId: overrides.assessmentSchemaId ?? current.assessmentSchemaID,
        start: past,
        deadline: future,
        selfEvaluationEnabled: overrides.selfEvaluationEnabled ?? false,
        selfEvaluationSchema: overrides.selfEvaluationSchema ?? current.selfEvaluationSchema,
        selfEvaluationStart: past,
        selfEvaluationDeadline: future,
        peerEvaluationEnabled: false,
        peerEvaluationSchema: current.peerEvaluationSchema,
        peerEvaluationStart: past,
        peerEvaluationDeadline: future,
        tutorEvaluationEnabled: false,
        tutorEvaluationSchema: current.tutorEvaluationSchema,
        tutorEvaluationStart: past,
        tutorEvaluationDeadline: future,
        evaluationResultsVisible: true,
      },
    }),
    'PUT config',
  )
}

export async function createCategory(
  api: APIRequestContext,
  phaseId: string,
  schemaId: string,
  name: string,
): Promise<void> {
  await ok(
    api.post(assessmentUrl(phaseId, 'category'), {
      data: {
        name,
        shortName: name.slice(0, 10),
        description: `${name} (created by e2e)`,
        weight: 1,
        assessmentSchemaID: schemaId,
      },
    }),
    'POST category',
  )
}

export async function createCompetency(
  api: APIRequestContext,
  phaseId: string,
  categoryId: string,
  name: string,
): Promise<void> {
  await ok(
    api.post(assessmentUrl(phaseId, 'competency'), {
      data: {
        categoryID: categoryId,
        name,
        shortName: name.slice(0, 10),
        description: `${name} (created by e2e)`,
        descriptionVeryBad: 'Far below expectations',
        descriptionBad: 'Below expectations',
        descriptionOk: 'Meets expectations',
        descriptionGood: 'Above expectations',
        descriptionVeryGood: 'Far above expectations',
        weight: 1,
      },
    }),
    'POST competency',
  )
}

export async function getAssessmentCategories(
  api: APIRequestContext,
  phaseId: string,
): Promise<CategoryWithCompetencies[]> {
  const res = await ok(
    api.get(assessmentUrl(phaseId, 'category/assessment/with-competencies')),
    'GET categories',
  )
  return (await res.json()) as CategoryWithCompetencies[]
}

export async function gradeCompetency(
  api: APIRequestContext,
  phaseId: string,
  courseParticipationId: string,
  competencyId: string,
  scoreLevel: 'veryBad' | 'bad' | 'ok' | 'good' | 'veryGood',
): Promise<void> {
  await ok(
    api.post(assessmentUrl(phaseId, 'student-assessment'), {
      data: {
        courseParticipationID: courseParticipationId,
        coursePhaseID: phaseId,
        competencyID: competencyId,
        scoreLevel,
      },
    }),
    'POST assessment',
  )
}

// Two steps like the UI: the completion row (remarks + grade suggestion) must
// exist before mark-complete — that endpoint only flips an EXISTING row to
// completed (and validates that no competency is left unassessed).
export async function markAssessmentComplete(
  api: APIRequestContext,
  phaseId: string,
  courseParticipationId: string,
  gradeSuggestion: number,
  comment: string,
): Promise<void> {
  const completion = {
    courseParticipationID: courseParticipationId,
    coursePhaseID: phaseId,
    comment,
    gradeSuggestion,
    author: 'e2e',
  }
  await ok(
    api.post(assessmentUrl(phaseId, 'student-assessment/completed'), {
      data: { ...completion, completed: false },
    }),
    'POST completion',
  )
  await ok(
    api.post(assessmentUrl(phaseId, 'student-assessment/completed/mark-complete'), {
      data: { ...completion, completed: true },
    }),
    'POST mark-complete',
  )
}

export async function releaseResults(api: APIRequestContext, phaseId: string): Promise<void> {
  await ok(api.post(assessmentUrl(phaseId, 'config/release')), 'POST release')
}

// ── Cleanup ────────────────────────────────────────────────────────────────

// Restores a fixture phase to its seeded state, in dependency order:
// 1. unrelease results and drop assessment completions (completed assessments
//    are not editable, so completions must go before their assessments),
// 2. delete every assessment of the phase — category/competency deletion and
//    schema changes are both rejected while assessment/evaluation data exists,
// 3. delete the configured schemas' categories (cascades competencies and any
//    leftover evaluations),
// 4. point the config back at the default template schemas,
// 5. delete the e2e-created schemas (only possible after step 4: the config's
//    schema FKs are ON DELETE RESTRICT).
// Student-authored evaluations cannot be deleted by admins — specs that create
// them clean those up as the student before calling this. Runs as admin
// (schema deletion is PromptAdmin-only) and tolerates partially set-up state
// so watch-mode re-runs start clean. The default template schemas ship without
// categories, so deleting all listed categories is safe even if a previous run
// already restored the defaults.
export async function resetAssessmentPhase(
  phaseId: string,
  options: {
    courseParticipationIds?: string[]
    schemaNames?: string[]
  } = {},
): Promise<void> {
  const admin = await apiContextFor('admin')
  try {
    await admin.post(assessmentUrl(phaseId, 'config/unrelease'))
    for (const participationId of options.courseParticipationIds ?? []) {
      await admin.delete(
        assessmentUrl(phaseId, `student-assessment/completed/course-participation/${participationId}`),
      )
    }
    const assessmentsRes = await admin.get(assessmentUrl(phaseId, 'student-assessment'))
    if (assessmentsRes.ok()) {
      const assessments = (await assessmentsRes.json()) as { id: string }[]
      for (const assessment of assessments) {
        await admin.delete(assessmentUrl(phaseId, `student-assessment/${assessment.id}`))
      }
    }
    for (const endpoint of [
      'category/assessment/with-competencies',
      'category/self/with-competencies',
    ]) {
      const res = await admin.get(assessmentUrl(phaseId, endpoint))
      if (!res.ok()) continue
      const categories = (await res.json()) as CategoryWithCompetencies[]
      for (const category of categories) {
        await admin.delete(assessmentUrl(phaseId, `category/${category.id}`))
      }
    }
    const schemas = await getSchemas(admin, phaseId)
    await putConfig(admin, phaseId, {
      assessmentSchemaId: schemas.find((s) => s.name === DEFAULT_SCHEMAS.assessment)?.id,
      selfEvaluationSchema: schemas.find((s) => s.name === DEFAULT_SCHEMAS.self)?.id,
    })
    for (const name of options.schemaNames ?? []) {
      const schema = schemas.find((s) => s.name === name)
      if (schema) {
        await admin.delete(assessmentUrl(phaseId, `assessment-schema/${schema.id}`))
      }
    }
  } finally {
    await admin.dispose()
  }
}

export async function apiAsRole(role: Role): Promise<APIRequestContext> {
  return apiContextFor(role)
}
