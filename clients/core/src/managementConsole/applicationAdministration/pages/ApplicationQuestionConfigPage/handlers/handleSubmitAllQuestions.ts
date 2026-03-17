import { ApplicationForm } from '../../../interfaces/form/applicationForm'
import { UpdateApplicationForm } from '../../../interfaces/form/updateApplicationForm'

import { ApplicationQuestionCardRef } from '../FormPages/ApplicationQuestionCard'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'

interface handleSubmitAllQuestionsProps {
  questionRefs: React.MutableRefObject<(ApplicationQuestionCardRef | null | undefined)[]>
  fetchedForm: ApplicationForm
  applicationQuestions: (
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionText
    | ApplicationQuestionFileUpload
  )[]
  setSubmitAttempted: (state: boolean) => void
  mutateApplicationForm: (updateForm: UpdateApplicationForm) => void
}

export const handleSubmitAllQuestions = async ({
  questionRefs,
  fetchedForm,
  applicationQuestions,
  setSubmitAttempted,
  mutateApplicationForm,
}: handleSubmitAllQuestionsProps) => {
  let allValid = true

  // Loop over each child's ref, call validate()
  for (const ref of questionRefs.current) {
    if (!ref) continue
    const isValid = await ref.validate()
    if (!isValid) {
      allValid = false
    }
  }
  setSubmitAttempted(true)
  if (allValid) {
    const deletedTextQuestion = fetchedForm?.questionsText
      .filter((q) => !applicationQuestions.some((aq) => aq.id === q.id))
      .map((q) => q.id)

    const deletedMultiSelectQuestion = fetchedForm?.questionsMultiSelect
      .filter((q) => !applicationQuestions.some((aq) => aq.id === q.id))
      .map((q) => q.id)

    const deletedFileUploadQuestion = fetchedForm?.questionsFileUpload
      .filter((q) => !applicationQuestions.some((aq) => aq.id === q.id))
      .map((q) => q.id)

    const questionsMultiSelect = applicationQuestions
      .filter((q) => 'options' in q)
      .map((q) => q as ApplicationQuestionMultiSelect)
      .map((q) => {
        if (!q.accessibleForOtherPhases) {
          return {
            ...q,
            accessKey: '', // Do not modify access key (is not shown if export switched off)
          }
        } else {
          return q
        }
      })

    const questionsText = applicationQuestions
      .filter((q) => !('options' in q) && !('allowedFileTypes' in q))
      .map((q) => q as ApplicationQuestionText)
      .map((q) => {
        if (!q.accessibleForOtherPhases) {
          return {
            ...q,
            accessKey: '', // Do not modify access key (is not shown if export switched off)
          }
        } else {
          return q
        }
      })

    const questionsFileUpload = applicationQuestions
      .filter((q) => 'allowedFileTypes' in q)
      .map((q) => q as ApplicationQuestionFileUpload)
      .map((q) => {
        if (!q.accessibleForOtherPhases) {
          return {
            ...q,
            accessKey: '', // Do not modify access key (is not shown if export switched off)
          }
        } else {
          return q
        }
      })

    const updateForm: UpdateApplicationForm = {
      deleteQuestionsText: deletedTextQuestion ?? [],
      deleteQuestionsMultiSelect: deletedMultiSelectQuestion ?? [],
      deleteQuestionsFileUpload: deletedFileUploadQuestion ?? [],
      createQuestionsText: questionsText.filter((q) => q.id.startsWith('not-valid-id-question-')),
      createQuestionsMultiSelect: questionsMultiSelect.filter((q) =>
        q.id.startsWith('not-valid-id-question-'),
      ),
      createQuestionsFileUpload: questionsFileUpload.filter((q) =>
        q.id.startsWith('not-valid-id-question-'),
      ),
      updateQuestionsText: questionsText.filter((q) => !q.id.startsWith('not-valid-id-question-')),
      updateQuestionsMultiSelect: questionsMultiSelect.filter(
        (q) => !q.id.startsWith('not-valid-id-question-'),
      ),
      updateQuestionsFileUpload: questionsFileUpload.filter(
        (q) => !q.id.startsWith('not-valid-id-question-'),
      ),
    }
    mutateApplicationForm(updateForm)
  } else {
    throw new Error('Not all questions are valid')
  }
}
