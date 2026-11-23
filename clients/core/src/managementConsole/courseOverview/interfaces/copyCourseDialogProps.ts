import type { Course } from '@tumaet/prompt-shared-state'

export interface CourseCopyDialogProps {
  courseId: string
  // The course to copy, when the caller has it at hand
  course?: Course
  isOpen: boolean
  onClose: () => void
  useTemplateCopy?: boolean
  createTemplate?: boolean
}

export type DialogStep = 'form' | 'warning' | 'loading'

export interface CopyabilityData {
  copyable: boolean
  missingPhaseTypes?: string[]
}
