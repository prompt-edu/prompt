export interface ServiceInfo {
  serviceName: string
  version?: string
  healthy: boolean
  capabilities: Record<string, boolean>
}

// Mirrors the constant keys defined in prompt-sdk/promptTypes/capabilities.go
export const CAPABILITY_PRIVACY_STUDENT_EXPORT = 'privacy.studentExport'
export const CAPABILITY_PRIVACY_STUDENT_DELETION = 'privacy.studentDeletion'
export const CAPABILITY_PHASE_COPY = 'phase.copy'
export const CAPABILITY_PHASE_CONFIG = 'phase.config'

export const CAPABILITY_LABELS: Record<string, string> = {
  [CAPABILITY_PHASE_COPY]: 'Phase Copy',
  [CAPABILITY_PHASE_CONFIG]: 'Phase Config',
  [CAPABILITY_PRIVACY_STUDENT_EXPORT]: 'Student Export (GDPR)',
  [CAPABILITY_PRIVACY_STUDENT_DELETION]: 'Student Deletion (GDPR)',
}
