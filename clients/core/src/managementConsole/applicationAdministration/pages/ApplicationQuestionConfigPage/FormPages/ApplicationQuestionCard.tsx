import { forwardRef, useEffect, useImperativeHandle, useState } from 'react'
import { useForm, UseFormReturn } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { ChevronDown, ChevronUp, GripVertical, Trash2 } from 'lucide-react'
import { DraggableProvidedDragHandleProps } from '@hello-pangea/dnd'
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  Button,
  Form,
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from '@tumaet/prompt-ui-components'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import {
  QuestionConfigFormData,
  QuestionConfigFormDataMultiSelect,
  questionConfigSchema,
} from '@core/validations/questionConfig'
import { MultiSelectConfig } from './MultiSelectConfig'
import { DeleteConfirmation } from '../components/DeleteConfirmation'
import { questionsEqual } from '../handlers/computeQuestionsModified'
import { QuestionStatus, QuestionStatusBadge } from '../components/QuestionStatusBadge'
import { checkCheckBoxQuestion } from '@core/publicPages/application/pages/ApplicationForm/utils/CheckBoxRequirements'
import {
  TitleField,
  DescriptionField,
  RequiredField,
  AllowedLengthField,
  PlaceholderField,
  ErrorMessageField,
  ValidationRegexField,
  ExportSettingsFields,
  AllowedFileTypesField,
  MaxFileSizeField,
} from './FormFields'

// If you plan to expose methods via this ref, define them here:
export interface ApplicationQuestionCardRef {
  validate: () => Promise<boolean>
  getValues: () => QuestionConfigFormData
}

interface ApplicationQuestionCardProps {
  question: ApplicationQuestionMultiSelect | ApplicationQuestionText | ApplicationQuestionFileUpload
  originalQuestion:
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionText
    | ApplicationQuestionFileUpload
    | undefined
  index: number
  onUpdate: (
    updatedQuestion:
      | ApplicationQuestionMultiSelect
      | ApplicationQuestionText
      | ApplicationQuestionFileUpload,
  ) => void
  submitAttempted: boolean
  onDelete: (id: string) => void
  dragHandleProps?: DraggableProvidedDragHandleProps | null
}

export const ApplicationQuestionCard = forwardRef<
  ApplicationQuestionCardRef | undefined, // or null if you prefer
  ApplicationQuestionCardProps
>(function ApplicationQuestionCard(
  { question, index, originalQuestion, onUpdate, submitAttempted, onDelete, dragHandleProps },
  ref,
) {
  const isNewQuestion = question.title === '' ? true : false
  const [isExpanded, setIsExpanded] = useState(isNewQuestion)
  const isMultiSelectType = 'options' in question
  const isFileUploadType = 'allowedFileTypes' in question
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [advancedSettingsOpen, setAdvancedSettingsOpen] = useState(false)

  const status: QuestionStatus = originalQuestion
    ? questionsEqual(question, originalQuestion)
      ? 'saved'
      : 'modified'
    : 'new'

  const questionType = isMultiSelectType
    ? 'multi-select'
    : isFileUploadType
      ? 'file-upload'
      : 'text'

  const form = useForm<QuestionConfigFormData>({
    resolver: zodResolver(questionConfigSchema),
    defaultValues: { type: questionType, ...question },
    mode: 'onTouched',
  })

  function shouldCollapseAdvancedOptions(
    formValues: Partial<
      Pick<
        QuestionConfigFormData & { validationRegex?: string },
        | 'placeholder'
        | 'validationRegex'
        | 'errorMessage'
        | 'accessKey'
        | 'accessibleForOtherPhases'
      >
    >,
  ): boolean {
    const isEmpty = (v: unknown) => v == null || (typeof v === 'string' && v.trim() === '')
    const hasAnyAdvanced =
      !!formValues.accessibleForOtherPhases ||
      !isEmpty(formValues.placeholder) ||
      !isEmpty(formValues.validationRegex) ||
      !isEmpty(formValues.errorMessage) ||
      !isEmpty(formValues.accessKey)

    return !hasAnyAdvanced
  }

  useEffect(() => {
    const subscription = form.watch((value) => {
      onUpdate({ ...question, ...value })
    })
    setAdvancedSettingsOpen(!shouldCollapseAdvancedOptions(form.getValues()))
    // Cleanup subscription on unmount
    return () => subscription.unsubscribe()
  }, [form.watch, question, onUpdate, form])

  // allow to call validate from the parent component
  useImperativeHandle(ref, () => ({
    async validate() {
      const valid = await form.trigger()
      return valid
    },
    getValues() {
      return form.getValues()
    },
  }))

  const isCheckboxQuestion =
    isMultiSelectType && checkCheckBoxQuestion(question as ApplicationQuestionMultiSelect)
  const isActualMultiSelect = isMultiSelectType && !isCheckboxQuestion

  return (
    <>
      <Card
        className={`mb-4 overflow-hidden ${submitAttempted && !form.formState.isValid ? 'border-red-500' : ''}`}
      >
        <CardHeader
          className='cursor-pointer'
          onClick={() => setIsExpanded(!isExpanded)}
          {...dragHandleProps} // This is the drag handle for the card
        >
          <div className='flex items-center justify-between'>
            <div className='flex items-center space-x-2'>
              <GripVertical className='cursor-move h-6 w-6 text-muted-foreground' />
              <div>
                <CardTitle>{question.title || `Untitled Question`}</CardTitle>
                <p className='text-sm text-muted-foreground mt-1'>
                  Question {index + 1}:{' '}
                  {isFileUploadType
                    ? 'File upload question'
                    : isMultiSelectType
                      ? isCheckboxQuestion
                        ? 'Checkbox'
                        : 'Multi-select question'
                      : 'Text question'}
                </p>
              </div>
            </div>
            <div className='flex items-center space-x-2'>
              <QuestionStatusBadge status={status} />
              <Button
                variant='ghost'
                size='sm'
                onClick={(e) => {
                  e.stopPropagation()
                  if (question.id.startsWith('no-valid-id')) {
                    onDelete(question.id)
                  } else {
                    setDeleteDialogOpen(true)
                  }
                }}
                aria-label='Delete question'
              >
                <Trash2 className='h-4 w-4 text-destructive' />
              </Button>
              {isExpanded ? <ChevronUp className='h-6 w-6' /> : <ChevronDown className='h-6 w-6' />}
            </div>
          </div>
        </CardHeader>
        {isExpanded && (
          <div>
            <CardContent>
              <Form {...form}>
                <div className='space-y-4'>
                  {/** For multi-select question the isRequired is controlled by min-select */}
                  {!isActualMultiSelect && <RequiredField form={form} />}

                  <TitleField form={form} />

                  <DescriptionField form={form} initialDescription={question.description} />

                  {!isMultiSelectType && !isFileUploadType && <AllowedLengthField form={form} />}

                  {isFileUploadType && (
                    <>
                      <AllowedFileTypesField form={form} />
                      <MaxFileSizeField form={form} />
                    </>
                  )}

                  {isMultiSelectType && !isCheckboxQuestion && (
                    <MultiSelectConfig
                      form={form as UseFormReturn<QuestionConfigFormDataMultiSelect>}
                    />
                  )}
                </div>
              </Form>
            </CardContent>
            <CardContent className='bg-[#fafafa] dark:bg-[#18181c] pt-2 -mt-1'>
              <Form {...form}>
                <Collapsible open={advancedSettingsOpen} onOpenChange={setAdvancedSettingsOpen}>
                  <CollapsibleTrigger className='flex w-full items-center justify-between cursor-pointer'>
                    <h3 className='text-xl mt-2 mb-2 font-medium'>Advanced Settings</h3>
                    {advancedSettingsOpen ? (
                      <ChevronUp className='h-4 w-4' />
                    ) : (
                      <ChevronDown className='h-4 w-4' />
                    )}
                  </CollapsibleTrigger>
                  <CollapsibleContent className='space-y-4'>
                    {/** Checkbox Questions do not have a placeholder */}
                    {!isCheckboxQuestion && !isFileUploadType && <PlaceholderField form={form} />}

                    {!isMultiSelectType && !isFileUploadType && (
                      <ValidationRegexField form={form} />
                    )}

                    {/** For multi-select question there is no need to specify an error message - it will be determined by max and min error */}
                    {!isActualMultiSelect && !isFileUploadType && (
                      <ErrorMessageField
                        form={form}
                        isCheckboxQuestion={isCheckboxQuestion}
                        isMultiSelectType={isMultiSelectType}
                      />
                    )}

                    <ExportSettingsFields form={form} />
                  </CollapsibleContent>
                </Collapsible>
              </Form>
            </CardContent>
          </div>
        )}
      </Card>
      {deleteDialogOpen && (
        <DeleteConfirmation
          isOpen={deleteDialogOpen}
          setOpen={setDeleteDialogOpen}
          onClick={(deleteConfirmed) => (deleteConfirmed ? onDelete(question.id) : null)}
        />
      )}
    </>
  )
})
