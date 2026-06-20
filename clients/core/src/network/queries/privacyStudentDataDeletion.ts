import { axiosInstance } from '@tumaet/prompt-shared-state'

export enum DeletionRequestStatus {
  pending_approval = 'pending_approval',
  in_progress = 'in_progress',
  succeeded = 'succeeded',
  failed = 'failed',
  rejected = 'rejected',
}

export enum DeletionSubrequestStatus {
  pending = 'pending',
  in_progress = 'in_progress',
  succeeded = 'succeeded',
  failed = 'failed',
}

export interface PrivacyDeletionSubrequest {
  id: string
  source_name: string
  status: DeletionSubrequestStatus
  created_at: string
  completed_at: string | null
}

export interface AdminPrivacyDeletionSubrequest extends PrivacyDeletionSubrequest {
  error_message: string
}

export interface PrivacyDeletionRequest {
  id: string
  user_id: string
  student_id: string | null
  requested_at: string
  status: DeletionRequestStatus
  auditor_id: string | null
  auditor_name: string
  auditor_email: string
  auditor_responded_at: string | null
  auditor_note: string
  completed_at: string | null
  subrequests: PrivacyDeletionSubrequest[]
}

export interface AdminPrivacyDeletionRequest extends Omit<PrivacyDeletionRequest, 'subrequests'> {
  subrequests: AdminPrivacyDeletionSubrequest[]
  student_first_name: string | null
  student_last_name: string | null
  student_email: string | null
}

export type AuditorDecision = 'approve' | 'reject'

export interface AuditorDecisionRequest {
  decision: AuditorDecision
  note: string
}

export type LatestDeletionResponse =
  | { status: 'exists'; request: PrivacyDeletionRequest }
  | { status: 'ready' }

export const requestStudentDataDeletion = async (): Promise<PrivacyDeletionRequest> => {
  try {
    return (await axiosInstance.post('/api/privacy/data-deletion')).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getLatestStudentDataDeletion = async (): Promise<LatestDeletionResponse> => {
  try {
    const response = await axiosInstance.get('/api/privacy/data-deletion')
    if (response.status === 204) return { status: 'ready' }
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getStudentDataDeletionStatus = async (
  requestID: string,
): Promise<PrivacyDeletionRequest> => {
  try {
    return (await axiosInstance.get(`/api/privacy/data-deletion/${requestID}`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getAllDeletionRequests = async (): Promise<AdminPrivacyDeletionRequest[]> => {
  try {
    return (await axiosInstance.get('/api/privacy/admin/data-deletions')).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const decideOnDeletionRequest = async (
  requestID: string,
  decision: AuditorDecisionRequest,
): Promise<PrivacyDeletionRequest> => {
  try {
    return (await axiosInstance.post(`/api/privacy/admin/data-deletions/${requestID}`, decision))
      .data
  } catch (err) {
    console.error(err)
    throw err
  }
}

// Admin-initiated deletion

export const adminInitiateDataDeletions = async (
  studentIDs: string[],
): Promise<PrivacyDeletionRequest[]> => {
  try {
    return (
      await axiosInstance.post('/api/privacy/admin/data-deletions', { student_ids: studentIDs })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getDataDeletionsStatus = async (ids: string[]): Promise<PrivacyDeletionRequest[]> => {
  try {
    return (await axiosInstance.post('/api/privacy/admin/data-deletions/status', { ids })).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
