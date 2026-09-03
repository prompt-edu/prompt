export enum ExportStatus {
  pending = 'pending',
  complete = 'complete',
  no_data = 'no_data',
  failed = 'failed',
  archived = 'archived',
}

export interface PrivacyExportDocument {
  id: string
  date_created: string
  source_name: string
  status: ExportStatus
  file_size: number | null
  downloaded_at: string | null
}

export interface PrivacyExport {
  id: string
  userID: string
  studentID: string
  status: ExportStatus
  date_created: string
  valid_until: string
  documents: PrivacyExportDocument[]
}

export type LatestExportResponse =
  | { status: 'exists'; export: PrivacyExport }
  | { status: 'rate_limited'; retry_after: string }
  | { status: 'ready' }

export interface AdminExportDoc {
  source_name: string
  status: ExportStatus
  downloaded: boolean
}

export interface AdminPrivacyExport {
  id: string
  user_id: string
  student_id: string | null
  student_first_name: string | null
  student_last_name: string | null
  student_email: string | null
  status: ExportStatus
  date_created: string
  valid_until: string
  next_request_allowed_at: string
  docs: AdminExportDoc[]
}

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
