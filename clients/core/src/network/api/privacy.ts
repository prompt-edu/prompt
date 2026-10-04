import type {
  AdminPrivacyDeletionRequest,
  AdminPrivacyExport,
  AuditorDecisionRequest,
  LatestDeletionResponse,
  LatestExportResponse,
  PrivacyDeletionRequest,
  PrivacyExport,
} from '@core/interfaces/privacy'
import { API_PREFIX, coreRequest, NO_CONTENT } from '../client'

const path = `${API_PREFIX}/privacy`
const exportsPath = `${path}/data-export`
const deletionsPath = `${path}/data-deletion`
const adminExportsPath = `${path}/admin/data-exports`
const adminDeletionsPath = `${path}/admin/data-deletions`

export const privacy = {
  requestExport: (): Promise<PrivacyExport> => coreRequest.post(exportsPath),

  // 204 when the user has no export yet. React Query rejects an undefined result, so the two
  // "latest" reads answer with a discriminated status instead.
  latestExport: async (): Promise<LatestExportResponse> => {
    const response = await coreRequest.getResponse<LatestExportResponse>(exportsPath)
    return response.status === NO_CONTENT ? { status: 'ready' } : response.data
  },

  exportStatus: (exportID: string): Promise<PrivacyExport> =>
    coreRequest.get(`${exportsPath}/${exportID}`),

  exportDocDownloadURL: async (exportID: string, docID: string): Promise<string> =>
    (
      await coreRequest.get<{ downloadUrl: string }>(
        `${exportsPath}/${exportID}/docs/${docID}/download-url`,
      )
    ).downloadUrl,

  requestDeletion: (): Promise<PrivacyDeletionRequest> => coreRequest.post(deletionsPath),

  latestDeletion: async (): Promise<LatestDeletionResponse> => {
    const response = await coreRequest.getResponse<LatestDeletionResponse>(deletionsPath)
    return response.status === NO_CONTENT ? { status: 'ready' } : response.data
  },

  deletionStatus: (requestID: string): Promise<PrivacyDeletionRequest> =>
    coreRequest.get(`${deletionsPath}/${requestID}`),

  listExports: (): Promise<AdminPrivacyExport[]> => coreRequest.get(adminExportsPath),

  removeExport: (exportID: string, options?: { resetRateLimit?: boolean }): Promise<void> =>
    coreRequest.del(
      `${adminExportsPath}/${exportID}`,
      options?.resetRateLimit ? { params: { reset_rate_limit: true } } : undefined,
    ),

  listDeletions: (): Promise<AdminPrivacyDeletionRequest[]> => coreRequest.get(adminDeletionsPath),

  decideOnDeletion: (
    requestID: string,
    decision: AuditorDecisionRequest,
  ): Promise<PrivacyDeletionRequest> =>
    coreRequest.post(`${adminDeletionsPath}/${requestID}`, decision),

  initiateDeletions: (studentIDs: string[]): Promise<PrivacyDeletionRequest[]> =>
    coreRequest.post(adminDeletionsPath, { student_ids: studentIDs }),

  deletionsStatus: (ids: string[]): Promise<PrivacyDeletionRequest[]> =>
    coreRequest.post(`${adminDeletionsPath}/status`, { ids }),
}
