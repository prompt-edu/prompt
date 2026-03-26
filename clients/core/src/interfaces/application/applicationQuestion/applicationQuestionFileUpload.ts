export interface ApplicationQuestionFileUpload {
  id: string
  coursePhaseID: string
  title: string
  description?: string
  isRequired: boolean
  allowedFileTypes?: string // Comma-separated extensions or MIME types
  maxFileSizeMB?: number
  orderNum: number
  accessibleForOtherPhases?: boolean
  accessKey?: string
}
