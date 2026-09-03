import type { GetApplication } from '@core/interfaces/application/getApplication'
import type { PostApplication } from '@core/interfaces/application/postApplication'
import type { AdditionalScore } from '@core/managementConsole/applicationAdministration/interfaces/additionalScore/additionalScore'
import type { AdditionalScoreUpload } from '@core/managementConsole/applicationAdministration/interfaces/additionalScore/additionalScoreUpload'
import type { ApplicationAssessment } from '@core/managementConsole/applicationAdministration/interfaces/applicationAssessment'
import type { ApplicationParticipation } from '@core/managementConsole/applicationAdministration/interfaces/applicationParticipation'
import type { ExportedApplicationAnswersResponse } from '@core/managementConsole/applicationAdministration/interfaces/exportedApplicationAnswers'
import type { ApplicationForm } from '@core/managementConsole/applicationAdministration/interfaces/form/applicationForm'
import type { UpdateApplicationForm } from '@core/managementConsole/applicationAdministration/interfaces/form/updateApplicationForm'
import type { ImportApplicationRequest } from '@core/managementConsole/applicationAdministration/interfaces/import/importApplicationRequest'
import type { ImportResult } from '@core/managementConsole/applicationAdministration/interfaces/import/importResult'
import type { UpdateCoursePhaseParticipationStatus } from '@tumaet/prompt-shared-state'
import { API_PREFIX, coreRequest } from '../client'

const path = (coursePhaseID: string) => `${API_PREFIX}/applications/${coursePhaseID}`

/** The staff-facing side of an application phase. Students reach their own through `apply`. */
export const applications = {
  listParticipations: (coursePhaseID: string): Promise<ApplicationParticipation[]> =>
    coreRequest.get(`${path(coursePhaseID)}/participations`),

  ofParticipant: (coursePhaseID: string, courseParticipationID: string): Promise<GetApplication> =>
    coreRequest.get(`${path(coursePhaseID)}/${courseParticipationID}`),

  form: (coursePhaseID: string): Promise<ApplicationForm> =>
    coreRequest.get(`${path(coursePhaseID)}/form`),

  exportedAnswers: (coursePhaseID: string): Promise<ExportedApplicationAnswersResponse> =>
    coreRequest.get(`${path(coursePhaseID)}/exported-answers`),

  // The server answers with null rather than an empty list when the phase defines no scores
  additionalScoreNames: async (coursePhaseID: string): Promise<AdditionalScore[]> =>
    (await coreRequest.get<AdditionalScore[] | null>(`${path(coursePhaseID)}/score`)) ?? [],

  fileDownloadURL: async (coursePhaseID: string, fileID: string): Promise<string> =>
    (
      await coreRequest.get<{ downloadUrl: string }>(
        `${path(coursePhaseID)}/files/${fileID}/download-url`,
      )
    ).downloadUrl,

  addManually: (coursePhaseID: string, application: PostApplication): Promise<void> =>
    coreRequest.post(path(coursePhaseID), application),

  importStudents: (
    coursePhaseID: string,
    request: ImportApplicationRequest,
  ): Promise<ImportResult> => coreRequest.post(`${path(coursePhaseID)}/import`, request),

  addAdditionalScore: (
    coursePhaseID: string,
    additionalScore: AdditionalScoreUpload,
  ): Promise<void> => coreRequest.post(`${path(coursePhaseID)}/score`, additionalScore),

  saveForm: (coursePhaseID: string, applicationForm: UpdateApplicationForm): Promise<void> =>
    coreRequest.put(`${path(coursePhaseID)}/form`, applicationForm),

  updateStatuses: (
    coursePhaseID: string,
    update: UpdateCoursePhaseParticipationStatus,
  ): Promise<void> => coreRequest.put(`${path(coursePhaseID)}/assessment`, update),

  saveAssessment: (
    coursePhaseID: string,
    courseParticipationID: string,
    assessment: ApplicationAssessment,
  ): Promise<void> =>
    coreRequest.put(`${path(coursePhaseID)}/${courseParticipationID}/assessment`, assessment),

  // The ids to delete travel in the body of the DELETE, not the path
  remove: (coursePhaseID: string, courseParticipationIDs: string[]): Promise<void> =>
    coreRequest.del(path(coursePhaseID), { data: courseParticipationIDs }),
}
