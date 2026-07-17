import type { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { AlignLeft, CheckSquare, type LucideIcon, Paperclip } from 'lucide-react'

export type QuestionKind = 'text' | 'multiSelect' | 'fileUpload'

// Discriminated union: the application form already separates questions by kind,
// so each question carries an explicit `kind` tag (added where the arrays are
// combined). This is reliable, unlike duck-typing optional fields such as
// `allowedFileTypes`, which the backend may omit for a file-upload question.
export type ApplicationQuestion =
  | (ApplicationQuestionText & { kind: 'text' })
  | (ApplicationQuestionMultiSelect & { kind: 'multiSelect' })
  | (ApplicationQuestionFileUpload & { kind: 'fileUpload' })

export const getQuestionKind = (question: ApplicationQuestion): QuestionKind => question.kind

export const QUESTION_KIND_META: Record<QuestionKind, { label: string; icon: LucideIcon }> = {
  text: { label: 'Text', icon: AlignLeft },
  multiSelect: { label: 'Multi-select', icon: CheckSquare },
  fileUpload: { label: 'File upload', icon: Paperclip },
}
