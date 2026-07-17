import type { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { AlignLeft, CheckSquare, type LucideIcon, Paperclip } from 'lucide-react'

export type ApplicationQuestion =
  | ApplicationQuestionText
  | ApplicationQuestionMultiSelect
  | ApplicationQuestionFileUpload

export type QuestionKind = 'text' | 'multiSelect' | 'fileUpload'

// Duck-typing the config-driven question union: only file-upload questions carry
// `allowedFileTypes`, only multi-select questions carry `options`. Adding a new
// question type means extending the union above and adding a case here plus a
// renderer in `answerRenderers.tsx`.
export const getQuestionKind = (question: ApplicationQuestion): QuestionKind => {
  if ('allowedFileTypes' in question) {
    return 'fileUpload'
  }
  if ('options' in question) {
    return 'multiSelect'
  }
  return 'text'
}

export const QUESTION_KIND_META: Record<QuestionKind, { label: string; icon: LucideIcon }> = {
  text: { label: 'Text', icon: AlignLeft },
  multiSelect: { label: 'Multi-select', icon: CheckSquare },
  fileUpload: { label: 'File upload', icon: Paperclip },
}
