export interface ServiceInfo {
  serviceName: string
  version?: string
  healthy: boolean
  capabilities: Record<string, boolean>
}

// Mirrors the constant keys defined in prompt-sdk/promptTypes/capabilities.go
export const CAPABILITY_PHASE_COPY = 'phase.copy'
export const CAPABILITY_PHASE_CONFIG = 'phase.config'
export const CAPABILITY_PRIVACY_EXPORT = 'privacy.export'
export const CAPABILITY_PRIVACY_DELETION = 'privacy.deletion'

export const CAPABILITY_LABELS: Record<string, string> = {
  [CAPABILITY_PHASE_COPY]: 'Phase Copy',
  [CAPABILITY_PHASE_CONFIG]: 'Phase Config',
  [CAPABILITY_PRIVACY_EXPORT]: 'Privacy Data Export',
  [CAPABILITY_PRIVACY_DELETION]: 'Privacy Data Deletion',
}
