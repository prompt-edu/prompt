import type { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { FormDescriptionHTML } from '@core/publicPages/application/pages/ApplicationForm/components/FormDescriptionHTML'
import { Tooltip, TooltipContent, TooltipTrigger } from '@tumaet/prompt-ui-components'
import { FileUploadAnswer, MultiSelectAnswer, TextAnswer } from './answerRenderers'
import { type ApplicationQuestion, getQuestionKind, QUESTION_KIND_META } from './questionKind'

interface ApplicationAnswerItemProps {
  coursePhaseId: string
  number: number
  question: ApplicationQuestion
  textAnswer: string
  multiSelectAnswer: string[]
  fileAnswer?: ApplicationAnswerFileUpload
}

// A single question section: a number gutter, then the title (with type icon +
// required marker), the question prompt as rich-text description, and the answer.
export const ApplicationAnswerItem = ({
  coursePhaseId,
  number,
  question,
  textAnswer,
  multiSelectAnswer,
  fileAnswer,
}: ApplicationAnswerItemProps) => {
  const kind = getQuestionKind(question)
  const { label, icon: Icon } = QUESTION_KIND_META[kind]

  return (
    <div className='flex gap-3 py-5 first:pt-0 last:pb-0'>
      <div className='flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-muted text-xs font-medium text-muted-foreground'>
        {number}
      </div>

      <div className='min-w-0 flex-1 space-y-2.5'>
        <div className='space-y-1'>
          <div className='flex items-center gap-2'>
            <Tooltip>
              <TooltipTrigger
                type='button'
                aria-label={`${label} question`}
                className='inline-flex shrink-0 items-center border-0 bg-transparent p-0 text-muted-foreground'
              >
                <Icon className='h-4 w-4' aria-hidden='true' />
              </TooltipTrigger>
              <TooltipContent>{label}</TooltipContent>
            </Tooltip>
            <span className='font-semibold'>{question.title}</span>
            {question.isRequired && (
              <span className='text-destructive'>
                <span aria-hidden='true'>*</span>
                <span className='sr-only'>required</span>
              </span>
            )}
          </div>

          {question.description && <FormDescriptionHTML htmlCode={question.description} />}
        </div>

        {kind === 'fileUpload' ? (
          <FileUploadAnswer coursePhaseId={coursePhaseId} answer={fileAnswer} />
        ) : kind === 'multiSelect' ? (
          <MultiSelectAnswer answer={multiSelectAnswer} />
        ) : (
          <TextAnswer answer={textAnswer} />
        )}
      </div>
    </div>
  )
}
