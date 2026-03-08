import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AlertCircle, Loader2 } from 'lucide-react'
import { DragDropContext, Draggable, Droppable, DropResult } from '@hello-pangea/dnd'
import {
  Card,
  CardContent,
  SaveChangesAlert,
  Alert,
  AlertTitle,
  AlertDescription,
} from '@tumaet/prompt-ui-components'
import {
  ApplicationQuestionCard,
  ApplicationQuestionCardRef,
} from './FormPages/ApplicationQuestionCard'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationForm } from '../../interfaces/form/applicationForm'
import { UpdateApplicationForm } from '../../interfaces/form/updateApplicationForm'
import { getApplicationForm } from '@core/network/queries/applicationForm'
import { updateApplicationForm } from '@core/network/mutations/updateApplicationForm'
import { handleSubmitAllQuestions } from './handlers/handleSubmitAllQuestions'
import { computeQuestionsModified } from './handlers/computeQuestionsModified'
import { handleQuestionUpdate } from './handlers/handleQuestionUpdate'
import { AddQuestionMenu } from './components/AddQuestionMenu'
import { ApplicationPreview } from '@core/publicPages/application/pages/ApplicationPreview/ApplicationPreview'

export const ApplicationQuestionConfig = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const [applicationQuestions, setApplicationQuestions] = useState<
    (ApplicationQuestionText | ApplicationQuestionMultiSelect | ApplicationQuestionFileUpload)[]
  >([])
  const questionRefs = useRef<Array<ApplicationQuestionCardRef | null | undefined>>([])
  // required to highlight questions red if submit is attempted and not valid
  const [submitAttempted, setSubmitAttempted] = useState(false)
  const queryClient = useQueryClient()

  const {
    data: fetchedForm,
    isPending: isApplicationFormPending,
    isError: isApplicationFormError,
    error: applicationFormError,
  } = useQuery<ApplicationForm>({
    queryKey: ['application_form', phaseId],
    queryFn: () => getApplicationForm(phaseId ?? 'undefined'),
  })
  const originalQuestions = [
    ...(fetchedForm?.questionsMultiSelect ?? []),
    ...(fetchedForm?.questionsText ?? []),
    ...(fetchedForm?.questionsFileUpload ?? []),
  ]
  const questionsModified = computeQuestionsModified(fetchedForm, applicationQuestions)

  const {
    mutate: mutateApplicationForm,
    isError: isMutateError,
    error: mutateError,
    isPending: isMutatePending,
  } = useMutation({
    mutationFn: (updateForm: UpdateApplicationForm) => {
      return updateApplicationForm(phaseId ?? 'undefined', updateForm)
    },
    onSuccess: () => {
      // invalidate query
      queryClient.invalidateQueries({ queryKey: ['application_form', phaseId] })
      // close this window
    },
  })

  const setQuestionsFromForm = (form: ApplicationForm) => {
    const combinedQuestions: (
      | ApplicationQuestionText
      | ApplicationQuestionMultiSelect
      | ApplicationQuestionFileUpload
    )[] = [
      ...(form.questionsMultiSelect ?? []),
      ...(form.questionsText ?? []),
      ...(form.questionsFileUpload ?? []),
    ]

    // Sort the combined questions by ordernum
    combinedQuestions.sort((a, b) => (a.orderNum ?? 0) - (b.orderNum ?? 0))

    setApplicationQuestions(combinedQuestions)
  }

  useEffect(() => {
    if (fetchedForm) {
      setQuestionsFromForm(fetchedForm)
    }
  }, [fetchedForm])

  const handleRevertAllQuestions = () => {
    if (fetchedForm) {
      setQuestionsFromForm(fetchedForm)
    }
  }

  const handleDeleteQuestion = (id: string) => {
    setApplicationQuestions((prev) => prev.filter((q) => q.id !== id))
  }

  const onDragEnd = (result: DropResult) => {
    if (!result.destination) {
      return
    }

    const newQuestions = Array.from(applicationQuestions)
    const [reorderedItem] = newQuestions.splice(result.source.index, 1)
    newQuestions.splice(result.destination.index, 0, reorderedItem)

    // Update orderNum for all questions
    const updatedQuestions = newQuestions.map((question, index) => ({
      ...question,
      orderNum: index + 1,
    }))

    setApplicationQuestions(updatedQuestions)
  }

  if (isApplicationFormPending || isApplicationFormError || isMutatePending || isMutateError) {
    return (
      <div className='space-y-6 w-full mx-auto'>
        <div className='flex justify-between items-center'>
          <h2 className='text-2xl font-semibold'>Application Questions</h2>
        </div>
        {(isApplicationFormPending || isMutatePending) && (
          <div className='flex justify-center items-center h-32'>
            <Loader2 className='h-8 w-8 animate-spin' />
          </div>
        )}
        {(isApplicationFormError || isMutateError) && (
          <Alert variant='destructive'>
            <AlertCircle className='h-4 w-4' />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              {applicationFormError instanceof Error
                ? applicationFormError.message
                : 'An error occurred while fetching the application form.'}
              {mutateError instanceof Error
                ? mutateError.message
                : 'An error occurred while updating the application form.'}
            </AlertDescription>
          </Alert>
        )}
      </div>
    )
  }

  return (
    <div className='w-full'>
      <h1 className='text-4xl font-bold mb-8'>Application Questions</h1>
      <div className='w-full mt-5'>
        <div className={`space-y-6 w-full mx-auto  ${questionsModified ? 'pb-10' : ''}`}>
          <div className='flex justify-between items-center'>
            <ApplicationPreview
              questionsMultiSelect={applicationQuestions.filter(
                (question) => 'options' in question,
              )}
              questionsText={applicationQuestions.filter(
                (question) => !('options' in question) && !('allowedFileTypes' in question),
              )}
              questionsFileUpload={applicationQuestions.filter(
                (question) => 'allowedFileTypes' in question,
              )}
            />
            <AddQuestionMenu
              setApplicationQuestions={setApplicationQuestions}
              applicationQuestions={applicationQuestions}
            />
          </div>
          {applicationQuestions.length > 0 ? (
            <DragDropContext onDragEnd={onDragEnd}>
              <Droppable droppableId='questions'>
                {(provided) => (
                  <div {...provided.droppableProps} ref={provided.innerRef}>
                    {applicationQuestions.map((question, index) => (
                      <Draggable key={question.id} draggableId={question.id} index={index}>
                        {(providedQuestionItem) => (
                          <div
                            ref={providedQuestionItem.innerRef}
                            {...providedQuestionItem.draggableProps}
                          >
                            <ApplicationQuestionCard
                              // Pass just the handleProps to restrict dragging to the header/icon
                              dragHandleProps={providedQuestionItem.dragHandleProps}
                              question={question}
                              originalQuestion={originalQuestions.find((q) => q.id === question.id)}
                              index={index}
                              onUpdate={(updatedQuestion) => {
                                handleQuestionUpdate(updatedQuestion, setApplicationQuestions)
                              }}
                              ref={(el) => {
                                questionRefs.current[index] = el
                              }}
                              submitAttempted={submitAttempted}
                              onDelete={handleDeleteQuestion}
                            />
                          </div>
                        )}
                      </Draggable>
                    ))}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </DragDropContext>
          ) : (
            <Card>
              <CardContent className='text-center py-8'>
                <p className='text-lg mb-4'>No questions added yet.</p>
              </CardContent>
            </Card>
          )}

          {questionsModified && (
            <SaveChangesAlert
              message='You have unsaved changes'
              handleRevert={handleRevertAllQuestions}
              saveChanges={() =>
                handleSubmitAllQuestions({
                  questionRefs,
                  fetchedForm,
                  applicationQuestions,
                  setSubmitAttempted,
                  mutateApplicationForm,
                })
              }
            />
          )}
        </div>
      </div>
    </div>
  )
}
