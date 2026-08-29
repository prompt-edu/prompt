import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { beforeEach, describe, expect, it } from 'vitest'
import type { UpdateActionItemRequest } from '../../interfaces/actionItem'
import type { CreateOrUpdateAssessmentRequest } from '../../interfaces/assessment'
import type { CreateOrUpdateAssessmentCompletionRequest } from '../../interfaces/assessmentCompletion'
import type {
  CreateAssessmentSchemaRequest,
  UpdateAssessmentSchemaRequest,
} from '../../interfaces/assessmentSchema'
import { AssessmentType } from '../../interfaces/assessmentType'
import type { CreateCategoryRequest, UpdateCategoryRequest } from '../../interfaces/category'
import type { CreateOrUpdateCategoryAssessmentRequest } from '../../interfaces/categoryAssessment'
import type { CreateCompetencyRequest, UpdateCompetencyRequest } from '../../interfaces/competency'
import type { CreateOrUpdateCoursePhaseConfigRequest } from '../../interfaces/coursePhaseConfig'
import type { CreateOrUpdateEvaluationRequest } from '../../interfaces/evaluation'
import type { EvaluationCompletionRequest } from '../../interfaces/evaluationCompletion'
import type {
  CreateFeedbackItemRequest,
  UpdateFeedbackItemRequest,
} from '../../interfaces/feedbackItem'
import { assessmentAxiosInstance } from '../client'
import { assessmentApi } from './index'

const PHASE = 'phase-1'
const PARTICIPATION = 'participation-1'
const SCHEMA = 'schema-1'
const ASSESSMENT_BASE = `http://assessment.test/assessment/api/course_phase/${PHASE}`
const CORE_BASE = `http://core.test/assessment/api/course_phase/${PHASE}`

const BODY = { marker: 'body' }
const schemaRequest = BODY as unknown as CreateAssessmentSchemaRequest
const schemaUpdate = BODY as unknown as UpdateAssessmentSchemaRequest
const category = BODY as unknown as CreateCategoryRequest
const categoryUpdate = { id: 'category-1' } as unknown as UpdateCategoryRequest
const competency = BODY as unknown as CreateCompetencyRequest
const competencyUpdate = { id: 'competency-1' } as unknown as UpdateCompetencyRequest
const categoryAssessment = BODY as unknown as CreateOrUpdateCategoryAssessmentRequest
const assessment = BODY as unknown as CreateOrUpdateAssessmentRequest
const completion = BODY as unknown as CreateOrUpdateAssessmentCompletionRequest
const actionItem = BODY as unknown as CreateActionItemRequestLike
const actionItemUpdate = { id: 'action-item-1' } as unknown as UpdateActionItemRequest
const evaluation = BODY as unknown as CreateOrUpdateEvaluationRequest
const evaluationCompletion = BODY as unknown as EvaluationCompletionRequest
const feedbackItem = BODY as unknown as CreateFeedbackItemRequest
const feedbackItemUpdate = BODY as unknown as UpdateFeedbackItemRequest
const coursePhaseConfig = BODY as unknown as CreateOrUpdateCoursePhaseConfigRequest

type CreateActionItemRequestLike = Parameters<typeof assessmentApi.actionItems.create>[1]

interface Route {
  name: string
  run: () => Promise<unknown>
  method: string
  url: string
  data?: unknown
  timeout?: number
  instance?: 'core'
}

const ROUTES: Route[] = [
  {
    name: 'schemas.list',
    run: () => assessmentApi.schemas.list(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/assessment-schema`,
  },
  {
    name: 'schemas.hasAssessmentData',
    run: () => assessmentApi.schemas.hasAssessmentData(PHASE, SCHEMA),
    method: 'get',
    url: `${ASSESSMENT_BASE}/assessment-schema/${SCHEMA}/has-assessment-data`,
  },
  {
    name: 'schemas.create',
    run: () => assessmentApi.schemas.create(PHASE, schemaRequest),
    method: 'post',
    url: `${ASSESSMENT_BASE}/assessment-schema`,
    data: BODY,
  },
  {
    name: 'schemas.update',
    run: () => assessmentApi.schemas.update(PHASE, SCHEMA, schemaUpdate),
    method: 'put',
    url: `${ASSESSMENT_BASE}/assessment-schema/${SCHEMA}`,
    data: BODY,
  },
  {
    name: 'categories.listWithCompetencies',
    run: () => assessmentApi.categories.listWithCompetencies(PHASE, AssessmentType.SELF),
    method: 'get',
    url: `${ASSESSMENT_BASE}/category/self/with-competencies`,
  },
  {
    name: 'categories.create',
    run: () => assessmentApi.categories.create(PHASE, category),
    method: 'post',
    url: `${ASSESSMENT_BASE}/category`,
    data: BODY,
  },
  {
    name: 'categories.update',
    run: () => assessmentApi.categories.update(PHASE, categoryUpdate),
    method: 'put',
    url: `${ASSESSMENT_BASE}/category/category-1`,
    data: { id: 'category-1' },
  },
  {
    name: 'categories.remove',
    run: () => assessmentApi.categories.remove(PHASE, 'category-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/category/category-1`,
  },
  {
    name: 'competencies.create',
    run: () => assessmentApi.competencies.create(PHASE, competency),
    method: 'post',
    url: `${ASSESSMENT_BASE}/competency`,
    data: BODY,
  },
  {
    name: 'competencies.update',
    run: () => assessmentApi.competencies.update(PHASE, competencyUpdate),
    method: 'put',
    url: `${ASSESSMENT_BASE}/competency/competency-1`,
    data: { id: 'competency-1' },
  },
  {
    name: 'competencies.remove',
    run: () => assessmentApi.competencies.remove(PHASE, 'competency-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/competency/competency-1`,
  },
  {
    name: 'categoryAssessments.save',
    run: () => assessmentApi.categoryAssessments.save(PHASE, categoryAssessment),
    method: 'post',
    url: `${ASSESSMENT_BASE}/category-assessment`,
    data: BODY,
    timeout: 10_000,
  },
  {
    name: 'assessments.listInPhase',
    run: () => assessmentApi.assessments.listInPhase(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment`,
  },
  {
    name: 'assessments.ofParticipant',
    run: () => assessmentApi.assessments.ofParticipant(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/${PARTICIPATION}`,
  },
  {
    name: 'assessments.scoreLevels',
    run: () => assessmentApi.assessments.scoreLevels(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/scoreLevel`,
  },
  {
    name: 'assessments.myResults',
    run: () => assessmentApi.assessments.myResults(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/my-results`,
  },
  {
    name: 'assessments.export',
    run: () => assessmentApi.assessments.export(PHASE, PARTICIPATION, 'json'),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/${PARTICIPATION}/export?format=json`,
  },
  {
    name: 'assessments.save',
    run: () => assessmentApi.assessments.save(PHASE, assessment),
    method: 'post',
    url: `${ASSESSMENT_BASE}/student-assessment`,
    data: BODY,
    timeout: 10_000,
  },
  {
    name: 'assessments.remove',
    run: () => assessmentApi.assessments.remove(PHASE, 'assessment-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/student-assessment/assessment-1`,
  },
  {
    name: 'completions.listInPhase',
    run: () => assessmentApi.completions.listInPhase(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/completed`,
  },
  {
    name: 'completions.myGradeSuggestion',
    run: () => assessmentApi.completions.myGradeSuggestion(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/completed/my-grade-suggestion`,
  },
  {
    name: 'completions.save',
    run: () => assessmentApi.completions.save(PHASE, completion),
    method: 'post',
    url: `${ASSESSMENT_BASE}/student-assessment/completed`,
    data: BODY,
  },
  {
    name: 'completions.markComplete',
    run: () => assessmentApi.completions.markComplete(PHASE, completion),
    method: 'post',
    url: `${ASSESSMENT_BASE}/student-assessment/completed/mark-complete`,
    data: BODY,
  },
  {
    name: 'completions.unmark',
    run: () => assessmentApi.completions.unmark(PHASE, PARTICIPATION),
    method: 'put',
    url: `${ASSESSMENT_BASE}/student-assessment/completed/course-participation/${PARTICIPATION}/unmark`,
    data: {},
  },
  {
    name: 'completions.remove',
    run: () => assessmentApi.completions.remove(PHASE, PARTICIPATION),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/student-assessment/completed/course-participation/${PARTICIPATION}`,
  },
  {
    name: 'actionItems.ofParticipant',
    run: () => assessmentApi.actionItems.ofParticipant(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/action-item/course-participation/${PARTICIPATION}`,
  },
  {
    name: 'actionItems.listMine',
    run: () => assessmentApi.actionItems.listMine(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/student-assessment/action-item/my-action-items`,
  },
  {
    name: 'actionItems.create',
    run: () => assessmentApi.actionItems.create(PHASE, actionItem),
    method: 'post',
    url: `${ASSESSMENT_BASE}/student-assessment/action-item`,
    data: BODY,
  },
  {
    name: 'actionItems.update',
    run: () => assessmentApi.actionItems.update(PHASE, actionItemUpdate),
    method: 'put',
    url: `${ASSESSMENT_BASE}/student-assessment/action-item/action-item-1`,
    data: { id: 'action-item-1' },
  },
  {
    name: 'actionItems.remove',
    run: () => assessmentApi.actionItems.remove(PHASE, 'action-item-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/student-assessment/action-item/action-item-1`,
  },
  {
    name: 'evaluations.listInPhase',
    run: () => assessmentApi.evaluations.listInPhase(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation`,
  },
  {
    name: 'evaluations.listMine',
    run: () => assessmentApi.evaluations.listMine(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/my-evaluations`,
  },
  {
    name: 'evaluations.ofSelf',
    run: () => assessmentApi.evaluations.ofSelf(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/self/${PARTICIPATION}`,
  },
  {
    name: 'evaluations.ofPeers',
    run: () => assessmentApi.evaluations.ofPeers(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/peer/${PARTICIPATION}`,
  },
  {
    name: 'evaluations.ofTutor',
    run: () => assessmentApi.evaluations.ofTutor(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/tutor/${PARTICIPATION}`,
  },
  {
    name: 'evaluations.myResults',
    run: () => assessmentApi.evaluations.myResults(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/my-results`,
  },
  {
    name: 'evaluations.save',
    run: () => assessmentApi.evaluations.save(PHASE, evaluation),
    method: 'post',
    url: `${ASSESSMENT_BASE}/evaluation`,
    data: BODY,
  },
  {
    name: 'evaluations.remove',
    run: () => assessmentApi.evaluations.remove(PHASE, 'evaluation-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/evaluation/evaluation-1`,
  },
  {
    name: 'evaluationCompletions.listInPhase',
    run: () => assessmentApi.evaluationCompletions.listInPhase(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/completed`,
  },
  {
    name: 'evaluationCompletions.listMine',
    run: () => assessmentApi.evaluationCompletions.listMine(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/completed/my-completions`,
  },
  {
    name: 'evaluationCompletions.markMine',
    run: () => assessmentApi.evaluationCompletions.markMine(PHASE, evaluationCompletion),
    method: 'post',
    url: `${ASSESSMENT_BASE}/evaluation/completed/my-completion/mark-complete`,
    data: BODY,
  },
  {
    name: 'evaluationCompletions.unmarkMine',
    run: () => assessmentApi.evaluationCompletions.unmarkMine(PHASE, evaluationCompletion),
    method: 'put',
    url: `${ASSESSMENT_BASE}/evaluation/completed/my-completion/unmark`,
    data: BODY,
  },
  {
    name: 'feedbackItems.listInPhase',
    run: () => assessmentApi.feedbackItems.listInPhase(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items`,
  },
  {
    name: 'feedbackItems.ofStudent',
    run: () => assessmentApi.feedbackItems.ofStudent(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items/course-participation/${PARTICIPATION}`,
  },
  {
    name: 'feedbackItems.ofTutor',
    run: () => assessmentApi.feedbackItems.ofTutor(PHASE, PARTICIPATION),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items/tutor/${PARTICIPATION}`,
  },
  {
    name: 'feedbackItems.listMine',
    run: () => assessmentApi.feedbackItems.listMine(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items/my-feedback`,
  },
  {
    name: 'feedbackItems.create',
    run: () => assessmentApi.feedbackItems.create(PHASE, feedbackItem),
    method: 'post',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items`,
    data: { ...BODY, coursePhaseID: PHASE },
  },
  {
    name: 'feedbackItems.update',
    run: () => assessmentApi.feedbackItems.update(PHASE, 'feedback-item-1', feedbackItemUpdate),
    method: 'put',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items/feedback-item-1`,
    data: { ...BODY, coursePhaseID: PHASE },
  },
  {
    name: 'feedbackItems.remove',
    run: () => assessmentApi.feedbackItems.remove(PHASE, 'feedback-item-1'),
    method: 'delete',
    url: `${ASSESSMENT_BASE}/evaluation/feedback-items/feedback-item-1`,
  },
  {
    name: 'config.get',
    run: () => assessmentApi.config.get(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/config`,
  },
  {
    name: 'config.participations',
    run: () => assessmentApi.config.participations(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/config/participations`,
  },
  {
    name: 'config.teams',
    run: () => assessmentApi.config.teams(PHASE),
    method: 'get',
    url: `${ASSESSMENT_BASE}/config/teams`,
  },
  {
    name: 'config.save',
    run: () => assessmentApi.config.save(PHASE, coursePhaseConfig),
    method: 'put',
    url: `${ASSESSMENT_BASE}/config`,
    data: BODY,
  },
  {
    name: 'config.releaseResults',
    run: () => assessmentApi.config.releaseResults(PHASE),
    method: 'post',
    url: `${ASSESSMENT_BASE}/config/release`,
    data: {},
  },
  {
    name: 'config.unreleaseResults',
    run: () => assessmentApi.config.unreleaseResults(PHASE),
    method: 'post',
    url: `${ASSESSMENT_BASE}/config/unrelease`,
    data: {},
  },
  {
    name: 'config.sendReminder',
    run: () => assessmentApi.config.sendReminder(PHASE, { evaluationType: AssessmentType.SELF }),
    method: 'post',
    url: `${CORE_BASE}/config/reminders/send`,
    data: { evaluationType: 'self' },
    instance: 'core',
  },
]

let captured: InternalAxiosRequestConfig[]
let responseStatus: number

const stubAdapter = (config: InternalAxiosRequestConfig): Promise<AxiosResponse> => {
  captured.push(config)
  return Promise.resolve({
    data: {},
    status: responseStatus,
    statusText: 'OK',
    headers: {},
    config,
  })
}

beforeEach(() => {
  captured = []
  responseStatus = 200
  assessmentAxiosInstance.defaults.adapter = stubAdapter
  axiosInstance.defaults.adapter = stubAdapter
})

describe('assessmentApi routes', () => {
  it.each(ROUTES)('$name sends the request it says it does', async (route) => {
    await route.run()

    expect(captured).toHaveLength(1)
    const config = captured[0]
    const instance = route.instance === 'core' ? axiosInstance : assessmentAxiosInstance

    expect(config.method).toBe(route.method)
    expect(instance.getUri(config)).toBe(route.url)
    expect(config.timeout || undefined).toBe(route.timeout)

    if (route.data === undefined) {
      expect(config.data).toBeUndefined()
    } else {
      expect(JSON.parse(config.data)).toEqual(route.data)
    }
  })

  it.each(ROUTES.filter((route) => route.method !== 'get'))(
    '$name declares a JSON body',
    async (route) => {
      await route.run()

      expect(captured[0].headers['Content-Type']).toBe('application/json')
    },
  )

  it.each(ROUTES.filter((route) => route.method === 'get'))(
    '$name sends no content type',
    async (route) => {
      await route.run()

      expect(captured[0].headers['Content-Type']).toBeUndefined()
    },
  )

  it('covers every endpoint the api exposes', () => {
    const exposed = Object.entries(assessmentApi).flatMap(([namespace, endpoints]) =>
      Object.keys(endpoints).map((endpoint) => `${namespace}.${endpoint}`),
    )

    expect([...new Set(ROUTES.map((route) => route.name))].sort()).toEqual(exposed.sort())
  })

  it('maps a 204 to null for the student evaluation results', async () => {
    responseStatus = 204

    await expect(assessmentApi.evaluations.myResults(PHASE)).resolves.toBeNull()
  })
})
