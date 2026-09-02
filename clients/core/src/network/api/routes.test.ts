import type { CopyCourse } from '@core/managementConsole/courseOverview/interfaces/copyCourse'
import type { PostCourse } from '@core/managementConsole/courseOverview/interfaces/postCourse'
import type { CoursePhaseType } from '@core/managementConsole/pages/SystemStatusPage/interfaces/coursePhaseType'
import { axiosInstance, notAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'
import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { beforeEach, describe, expect, it } from 'vitest'
import { coreApi } from './index'

const CORE = 'http://core.test'
const COURSE = 'course-1'
const PHASE = 'phase-1'
const PARTICIPATION = 'participation-1'
const STUDENT = 'student-1'
const CAMPAIGN = 'campaign-1'

const BODY = { marker: 'body' }
const body = <T>() => BODY as unknown as T

const DATE = new Date(2026, 8, 1)
const postCourse: PostCourse = {
  name: 'Course',
  startDate: DATE,
  endDate: DATE,
  courseType: 'lecture',
  ects: 10,
  semesterTag: 'ios2526',
  restrictedMetaData: {},
  studentReadableData: {},
  shortDescription: 'short',
  template: false,
}
const serializedCourse = { ...postCourse, startDate: '2026-09-01', endDate: '2026-09-01' }

const copyCourse: CopyCourse = {
  name: 'Course',
  semesterTag: 'ios2526',
  startDate: DATE,
  endDate: DATE,
  template: false,
}
const serializedCopy = { ...copyCourse, startDate: '2026-09-01', endDate: '2026-09-01' }

const service = { id: 'assessment', baseUrl: 'http://assessment.test' } as CoursePhaseType

type Instance = 'core' | 'public' | 'raw'

interface Route {
  name: string
  run: () => Promise<unknown>
  method: string
  url: string
  data?: unknown
  instance?: Instance
}

const applicationsPath = `${CORE}/api/applications/${PHASE}`
const coursesPath = `${CORE}/api/courses`
const notesPath = `${CORE}/api/instructor-notes`
const keycloakPath = `${CORE}/api/keycloak`
const campaignsPath = `${coursesPath}/${COURSE}/mail-campaigns`
const privacyPath = `${CORE}/api/privacy`
const studentsPath = `${CORE}/api/students`

const ROUTES: Route[] = [
  {
    name: 'applications.listParticipations',
    run: () => coreApi.applications.listParticipations(PHASE),
    method: 'get',
    url: `${applicationsPath}/participations`,
  },
  {
    name: 'applications.ofParticipant',
    run: () => coreApi.applications.ofParticipant(PHASE, PARTICIPATION),
    method: 'get',
    url: `${applicationsPath}/${PARTICIPATION}`,
  },
  {
    name: 'applications.form',
    run: () => coreApi.applications.form(PHASE),
    method: 'get',
    url: `${applicationsPath}/form`,
  },
  {
    name: 'applications.exportedAnswers',
    run: () => coreApi.applications.exportedAnswers(PHASE),
    method: 'get',
    url: `${applicationsPath}/exported-answers`,
  },
  {
    name: 'applications.additionalScoreNames',
    run: () => coreApi.applications.additionalScoreNames(PHASE),
    method: 'get',
    url: `${applicationsPath}/score`,
  },
  {
    name: 'applications.fileDownloadURL',
    run: () => coreApi.applications.fileDownloadURL(PHASE, 'file-1'),
    method: 'get',
    url: `${applicationsPath}/files/file-1/download-url`,
  },
  {
    name: 'applications.addManually',
    run: () => coreApi.applications.addManually(PHASE, body()),
    method: 'post',
    url: applicationsPath,
    data: BODY,
  },
  {
    name: 'applications.importStudents',
    run: () => coreApi.applications.importStudents(PHASE, body()),
    method: 'post',
    url: `${applicationsPath}/import`,
    data: BODY,
  },
  {
    name: 'applications.addAdditionalScore',
    run: () => coreApi.applications.addAdditionalScore(PHASE, body()),
    method: 'post',
    url: `${applicationsPath}/score`,
    data: BODY,
  },
  {
    name: 'applications.saveForm',
    run: () => coreApi.applications.saveForm(PHASE, body()),
    method: 'put',
    url: `${applicationsPath}/form`,
    data: BODY,
  },
  {
    name: 'applications.updateStatuses',
    run: () => coreApi.applications.updateStatuses(PHASE, body()),
    method: 'put',
    url: `${applicationsPath}/assessment`,
    data: BODY,
  },
  {
    name: 'applications.saveAssessment',
    run: () => coreApi.applications.saveAssessment(PHASE, PARTICIPATION, body()),
    method: 'put',
    url: `${applicationsPath}/${PARTICIPATION}/assessment`,
    data: BODY,
  },
  {
    name: 'applications.remove',
    run: () => coreApi.applications.remove(PHASE, [PARTICIPATION]),
    method: 'delete',
    url: applicationsPath,
    data: [PARTICIPATION],
  },

  {
    name: 'apply.listOpen',
    run: () => coreApi.apply.listOpen(),
    method: 'get',
    url: `${CORE}/api/apply`,
    instance: 'public',
  },
  {
    name: 'apply.form',
    run: () => coreApi.apply.form(PHASE),
    method: 'get',
    url: `${CORE}/api/apply/${PHASE}`,
    instance: 'public',
  },
  {
    name: 'apply.submitExternal',
    run: () => coreApi.apply.submitExternal(PHASE, body()),
    method: 'post',
    url: `${CORE}/api/apply/${PHASE}`,
    data: BODY,
    instance: 'public',
  },
  {
    name: 'apply.mine',
    run: () => coreApi.apply.mine(PHASE),
    method: 'get',
    url: `${CORE}/api/apply/authenticated/${PHASE}`,
  },
  {
    name: 'apply.submitAuthenticated',
    run: () => coreApi.apply.submitAuthenticated(PHASE, body()),
    method: 'post',
    url: `${CORE}/api/apply/authenticated/${PHASE}`,
    data: BODY,
  },

  {
    name: 'auditLog.inCourse',
    run: () => coreApi.auditLog.inCourse(COURSE, { outcome: 'denied' }, 50),
    method: 'get',
    url: `${coursesPath}/${COURSE}/audit-log?outcome=denied&limit=50`,
  },
  {
    name: 'auditLog.global',
    run: () => coreApi.auditLog.global({}, 50),
    method: 'get',
    url: `${CORE}/api/audit-log?limit=50`,
  },
  {
    name: 'auditLog.status',
    run: () => coreApi.auditLog.status(),
    method: 'get',
    url: `${CORE}/api/audit-log/status`,
  },

  {
    name: 'courseGraphs.phase',
    run: () => coreApi.courseGraphs.phase(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/phase_graph`,
  },
  {
    name: 'courseGraphs.phaseData',
    run: () => coreApi.courseGraphs.phaseData(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/phase_data_graph`,
  },
  {
    name: 'courseGraphs.participationData',
    run: () => coreApi.courseGraphs.participationData(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/participation_data_graph`,
  },
  {
    name: 'courseGraphs.savePhase',
    run: () => coreApi.courseGraphs.savePhase(COURSE, body()),
    method: 'put',
    url: `${coursesPath}/${COURSE}/phase_graph`,
    data: BODY,
  },
  {
    name: 'courseGraphs.savePhaseData',
    run: () => coreApi.courseGraphs.savePhaseData(COURSE, []),
    method: 'put',
    url: `${coursesPath}/${COURSE}/phase_data_graph`,
    data: [],
  },
  {
    name: 'courseGraphs.saveParticipationData',
    run: () => coreApi.courseGraphs.saveParticipationData(COURSE, []),
    method: 'put',
    url: `${coursesPath}/${COURSE}/participation_data_graph`,
    data: [],
  },

  {
    name: 'coursePhases.byID',
    run: () => coreApi.coursePhases.byID(PHASE),
    method: 'get',
    url: `${CORE}/api/course_phases/${PHASE}`,
  },
  {
    name: 'coursePhases.create',
    run: () => coreApi.coursePhases.create({ courseID: COURSE } as never),
    method: 'post',
    url: `${CORE}/api/course_phases/course/${COURSE}`,
    data: { courseID: COURSE },
  },
  {
    name: 'coursePhases.update',
    run: () => coreApi.coursePhases.update({ id: PHASE } as never),
    method: 'put',
    url: `${CORE}/api/course_phases/${PHASE}`,
    data: { id: PHASE },
  },
  {
    name: 'coursePhases.remove',
    run: () => coreApi.coursePhases.remove(PHASE),
    method: 'delete',
    url: `${CORE}/api/course_phases/${PHASE}`,
  },
  {
    name: 'coursePhases.listTypes',
    run: () => coreApi.coursePhases.listTypes(),
    method: 'get',
    url: `${CORE}/api/course_phase_types`,
  },
  {
    name: 'coursePhases.listTypesForScope',
    run: () => coreApi.coursePhases.listTypesForScope(true),
    method: 'get',
    url: `${CORE}/api/course_phase_types?for_self=true`,
  },

  {
    name: 'courses.list',
    run: () => coreApi.courses.list(),
    method: 'get',
    url: `${coursesPath}/`,
  },
  {
    name: 'courses.listOwnIDs',
    run: () => coreApi.courses.listOwnIDs(),
    method: 'get',
    url: `${coursesPath}/self`,
  },
  {
    name: 'courses.listTemplates',
    run: () => coreApi.courses.listTemplates(),
    method: 'get',
    url: `${coursesPath}/template`,
  },
  {
    name: 'courses.nameExists',
    run: () => coreApi.courses.nameExists('Course', 'ios2526'),
    method: 'get',
    url: `${coursesPath}/check-name?name=Course&semesterTag=ios2526`,
  },
  {
    name: 'courses.myParticipation',
    run: () => coreApi.courses.myParticipation(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/participations/self`,
  },
  {
    name: 'courses.templateStatus',
    run: () => coreApi.courses.templateStatus(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/template`,
  },
  {
    name: 'courses.copyability',
    run: () => coreApi.courses.copyability(COURSE),
    method: 'get',
    url: `${coursesPath}/${COURSE}/copyable`,
  },
  {
    name: 'courses.create',
    run: () => coreApi.courses.create(postCourse),
    method: 'post',
    url: `${coursesPath}/`,
    data: serializedCourse,
  },
  {
    name: 'courses.copy',
    run: () => coreApi.courses.copy(COURSE, copyCourse),
    method: 'post',
    url: `${coursesPath}/${COURSE}/copy`,
    data: serializedCopy,
  },
  {
    name: 'courses.update',
    run: () => coreApi.courses.update(COURSE, { name: 'Course' } as never),
    method: 'put',
    url: `${coursesPath}/${COURSE}`,
    data: { name: 'Course' },
  },
  {
    name: 'courses.setTemplateStatus',
    run: () => coreApi.courses.setTemplateStatus(COURSE, body()),
    method: 'put',
    url: `${coursesPath}/${COURSE}/template`,
    data: BODY,
  },
  {
    name: 'courses.setArchived',
    run: () => coreApi.courses.setArchived(COURSE, { archived: true }),
    method: 'put',
    url: `${coursesPath}/${COURSE}/archive`,
    data: { archived: true },
  },
  {
    name: 'courses.remove',
    run: () => coreApi.courses.remove(COURSE),
    method: 'delete',
    url: `${coursesPath}/${COURSE}`,
  },

  {
    name: 'instructorNotes.ofStudent',
    run: () => coreApi.instructorNotes.ofStudent(STUDENT),
    method: 'get',
    url: `${notesPath}/s/${STUDENT}`,
  },
  {
    name: 'instructorNotes.create',
    run: () => coreApi.instructorNotes.create(STUDENT, body()),
    method: 'post',
    url: `${notesPath}/s/${STUDENT}`,
    data: BODY,
  },
  {
    name: 'instructorNotes.remove',
    run: () => coreApi.instructorNotes.remove('note-1'),
    method: 'delete',
    url: `${notesPath}/note-1`,
  },
  {
    name: 'instructorNotes.listTags',
    run: () => coreApi.instructorNotes.listTags(),
    method: 'get',
    url: `${notesPath}/tags`,
  },
  {
    name: 'instructorNotes.createTag',
    run: () => coreApi.instructorNotes.createTag(body()),
    method: 'post',
    url: `${notesPath}/tags`,
    data: BODY,
  },
  {
    name: 'instructorNotes.updateTag',
    run: () => coreApi.instructorNotes.updateTag('tag-1', body()),
    method: 'put',
    url: `${notesPath}/tags/tag-1`,
    data: BODY,
  },
  {
    name: 'instructorNotes.removeTag',
    run: () => coreApi.instructorNotes.removeTag('tag-1'),
    method: 'delete',
    url: `${notesPath}/tags/tag-1`,
  },

  {
    name: 'keycloak.status',
    run: () => coreApi.keycloak.status(),
    method: 'get',
    url: `${keycloakPath}/status`,
  },
  {
    name: 'keycloak.courseStaff',
    run: () => coreApi.keycloak.courseStaff(COURSE),
    method: 'get',
    url: `${keycloakPath}/${COURSE}/group/staff`,
  },
  {
    name: 'keycloak.searchUsers',
    run: () => coreApi.keycloak.searchUsers('ab'),
    method: 'get',
    url: `${keycloakPath}/users/search?q=ab&limit=20`,
  },
  {
    name: 'keycloak.addStaffMember',
    run: () => coreApi.keycloak.addStaffMember(COURSE, 'Lecturer', 'user-1'),
    method: 'put',
    url: `${keycloakPath}/${COURSE}/group/Lecturer/members/user-1`,
  },
  {
    name: 'keycloak.removeStaffMember',
    run: () => coreApi.keycloak.removeStaffMember(COURSE, 'Editor', 'user-1'),
    method: 'delete',
    url: `${keycloakPath}/${COURSE}/group/Editor/members/user-1`,
  },

  {
    name: 'mailCampaigns.list',
    run: () => coreApi.mailCampaigns.list(COURSE),
    method: 'get',
    url: campaignsPath,
  },
  {
    name: 'mailCampaigns.byID',
    run: () => coreApi.mailCampaigns.byID(COURSE, CAMPAIGN),
    method: 'get',
    url: `${campaignsPath}/${CAMPAIGN}`,
  },
  {
    name: 'mailCampaigns.recipientPreview',
    run: () => coreApi.mailCampaigns.recipientPreview(COURSE, CAMPAIGN),
    method: 'get',
    url: `${campaignsPath}/${CAMPAIGN}/recipients-preview`,
  },
  {
    name: 'mailCampaigns.create',
    run: () => coreApi.mailCampaigns.create(COURSE, body()),
    method: 'post',
    url: campaignsPath,
    data: BODY,
  },
  {
    name: 'mailCampaigns.update',
    run: () => coreApi.mailCampaigns.update(COURSE, CAMPAIGN, body()),
    method: 'put',
    url: `${campaignsPath}/${CAMPAIGN}`,
    data: BODY,
  },
  {
    name: 'mailCampaigns.remove',
    run: () => coreApi.mailCampaigns.remove(COURSE, CAMPAIGN),
    method: 'delete',
    url: `${campaignsPath}/${CAMPAIGN}`,
  },
  {
    name: 'mailCampaigns.copy',
    run: () => coreApi.mailCampaigns.copy(COURSE, CAMPAIGN),
    method: 'post',
    url: `${campaignsPath}/${CAMPAIGN}/copy`,
  },
  {
    name: 'mailCampaigns.send',
    run: () => coreApi.mailCampaigns.send(COURSE, CAMPAIGN),
    method: 'post',
    url: `${campaignsPath}/${CAMPAIGN}/send`,
  },
  {
    name: 'mailCampaigns.resendFailed',
    run: () => coreApi.mailCampaigns.resendFailed(COURSE, CAMPAIGN),
    method: 'post',
    url: `${campaignsPath}/${CAMPAIGN}/resend-failed`,
  },
  {
    name: 'mailCampaigns.testSend',
    run: () => coreApi.mailCampaigns.testSend(COURSE, CAMPAIGN),
    method: 'post',
    url: `${campaignsPath}/${CAMPAIGN}/test`,
  },

  {
    name: 'mailing.sendStatusMail',
    run: () => coreApi.mailing.sendStatusMail(PHASE, body()),
    method: 'put',
    url: `${CORE}/api/mailing/${PHASE}`,
    data: BODY,
  },

  {
    name: 'privacy.requestExport',
    run: () => coreApi.privacy.requestExport(),
    method: 'post',
    url: `${privacyPath}/data-export`,
  },
  {
    name: 'privacy.latestExport',
    run: () => coreApi.privacy.latestExport(),
    method: 'get',
    url: `${privacyPath}/data-export`,
  },
  {
    name: 'privacy.exportStatus',
    run: () => coreApi.privacy.exportStatus('export-1'),
    method: 'get',
    url: `${privacyPath}/data-export/export-1`,
  },
  {
    name: 'privacy.exportDocDownloadURL',
    run: () => coreApi.privacy.exportDocDownloadURL('export-1', 'doc-1'),
    method: 'get',
    url: `${privacyPath}/data-export/export-1/docs/doc-1/download-url`,
  },
  {
    name: 'privacy.requestDeletion',
    run: () => coreApi.privacy.requestDeletion(),
    method: 'post',
    url: `${privacyPath}/data-deletion`,
  },
  {
    name: 'privacy.latestDeletion',
    run: () => coreApi.privacy.latestDeletion(),
    method: 'get',
    url: `${privacyPath}/data-deletion`,
  },
  {
    name: 'privacy.deletionStatus',
    run: () => coreApi.privacy.deletionStatus('request-1'),
    method: 'get',
    url: `${privacyPath}/data-deletion/request-1`,
  },
  {
    name: 'privacy.listExports',
    run: () => coreApi.privacy.listExports(),
    method: 'get',
    url: `${privacyPath}/admin/data-exports`,
  },
  {
    name: 'privacy.removeExport',
    run: () => coreApi.privacy.removeExport('export-1', { resetRateLimit: true }),
    method: 'delete',
    url: `${privacyPath}/admin/data-exports/export-1?reset_rate_limit=true`,
  },
  {
    name: 'privacy.listDeletions',
    run: () => coreApi.privacy.listDeletions(),
    method: 'get',
    url: `${privacyPath}/admin/data-deletions`,
  },
  {
    name: 'privacy.decideOnDeletion',
    run: () => coreApi.privacy.decideOnDeletion('request-1', { decision: 'approve', note: '' }),
    method: 'post',
    url: `${privacyPath}/admin/data-deletions/request-1`,
    data: { decision: 'approve', note: '' },
  },
  {
    name: 'privacy.initiateDeletions',
    run: () => coreApi.privacy.initiateDeletions([STUDENT]),
    method: 'post',
    url: `${privacyPath}/admin/data-deletions`,
    data: { student_ids: [STUDENT] },
  },
  {
    name: 'privacy.deletionsStatus',
    run: () => coreApi.privacy.deletionsStatus(['request-1']),
    method: 'post',
    url: `${privacyPath}/admin/data-deletions/status`,
    data: { ids: ['request-1'] },
  },

  {
    name: 'students.byID',
    run: () => coreApi.students.byID(STUDENT),
    method: 'get',
    url: `${studentsPath}/${STUDENT}`,
  },
  {
    name: 'students.enrollments',
    run: () => coreApi.students.enrollments(STUDENT),
    method: 'get',
    url: `${studentsPath}/${STUDENT}/enrollments`,
  },
  {
    name: 'students.withCourses',
    run: () => coreApi.students.withCourses(),
    method: 'get',
    url: `${studentsPath}/with-courses`,
  },
  {
    name: 'students.search',
    run: () => coreApi.students.search('mueller'),
    method: 'get',
    url: `${studentsPath}/search/mueller`,
  },
  {
    name: 'students.update',
    run: () => coreApi.students.update({ id: STUDENT } as never),
    method: 'put',
    url: `${studentsPath}/${STUDENT}`,
    data: { id: STUDENT },
  },

  {
    name: 'system.coreInfo',
    run: () => coreApi.system.coreInfo(),
    method: 'get',
    url: `${CORE}/api/hello`,
  },
  {
    name: 'system.serviceInfo',
    run: () => coreApi.system.serviceInfo(service),
    method: 'get',
    url: 'http://assessment.test/info',
    instance: 'raw',
  },
]

const INSTANCES = {
  core: axiosInstance,
  public: notAuthenticatedAxiosInstance,
  raw: axios,
} as const

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
  for (const instance of Object.values(INSTANCES)) {
    instance.defaults.adapter = stubAdapter
  }
})

describe('coreApi routes', () => {
  it.each(ROUTES)('$name sends the request it says it does', async (route) => {
    await route.run()

    expect(captured).toHaveLength(1)
    const config = captured[0]
    const instance = INSTANCES[route.instance ?? 'core']

    expect(config.method).toBe(route.method)
    expect(instance.getUri(config)).toBe(route.url)

    if (route.data === undefined) {
      expect(config.data).toBeUndefined()
    } else {
      expect(JSON.parse(config.data)).toEqual(route.data)
    }
  })

  it.each(ROUTES.filter((route) => route.method !== 'get' && route.instance !== 'raw'))(
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
    const exposed = Object.entries(coreApi).flatMap(([namespace, endpoints]) =>
      Object.keys(endpoints).map((endpoint) => `${namespace}.${endpoint}`),
    )

    expect([...new Set(ROUTES.map((route) => route.name))].sort()).toEqual(exposed.sort())
  })

  it('maps a 204 to a ready status on the two privacy reads that can answer empty', async () => {
    responseStatus = 204

    await expect(coreApi.privacy.latestExport()).resolves.toEqual({ status: 'ready' })
    await expect(coreApi.privacy.latestDeletion()).resolves.toEqual({ status: 'ready' })
  })

  it('answers with an empty list when a user has access to no course', async () => {
    const unauthorized = (config: InternalAxiosRequestConfig) =>
      Promise.reject(
        Object.assign(new Error('Request failed with status code 401'), {
          isAxiosError: true,
          config,
          response: { status: 401, data: {}, statusText: '', headers: {}, config },
        }),
      )
    axiosInstance.defaults.adapter = unauthorized

    await expect(coreApi.courses.list()).resolves.toEqual([])
    await expect(coreApi.courses.listOwnIDs()).resolves.toEqual([])
  })

  it('rethrows anything else, so the caller still sees the failure', async () => {
    axiosInstance.defaults.adapter = () => Promise.reject(new Error('network down'))

    await expect(coreApi.courses.list()).rejects.toThrow('network down')
  })

  it('answers with an empty list where the server sends null instead of one', async () => {
    axiosInstance.defaults.adapter = (config) =>
      Promise.resolve({ data: null, status: 200, statusText: 'OK', headers: {}, config })

    await expect(coreApi.mailCampaigns.list(COURSE)).resolves.toEqual([])
    await expect(coreApi.applications.additionalScoreNames(PHASE)).resolves.toEqual([])
  })
})
