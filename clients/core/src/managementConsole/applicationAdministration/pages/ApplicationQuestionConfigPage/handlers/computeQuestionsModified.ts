import { ApplicationForm } from '../../../interfaces/form/applicationForm'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'

export const computeQuestionsModified = (
  fetchedForm: ApplicationForm | undefined,
  storedApplicationQuestions: (
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload
  )[],
) => {
  const combinedQuestions: (
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload
  )[] = [
    ...(fetchedForm?.questionsMultiSelect ?? []),
    ...(fetchedForm?.questionsText ?? []),
    ...(fetchedForm?.questionsFileUpload ?? []),
  ]
  const modified = !storedApplicationQuestions.every((q) => {
    const originalQuestion = combinedQuestions.find((aq) => aq.id === q.id)
    if (!originalQuestion) return false
    return questionsEqual(q, originalQuestion)
  })
  const deletedQuestion = !combinedQuestions.every((q) =>
    storedApplicationQuestions.some((aq) => aq.id === q.id),
  )

  return modified || deletedQuestion
}

export const questionsEqual = (
  question1:
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload,
  question2:
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload,
): boolean => {
  const question1IsMultiSelect = 'options' in question1
  const question2IsMultiSelect = 'options' in question2
  const question1IsFileUpload = 'allowedFileTypes' in question1
  const question2IsFileUpload = 'allowedFileTypes' in question2

  if (question1IsMultiSelect !== question2IsMultiSelect) return false
  if (question1IsFileUpload !== question2IsFileUpload) return false

  const basicEqual =
    question1.id === question2.id &&
    question1.coursePhaseID === question2.coursePhaseID &&
    question1.title === question2.title &&
    question1.description === question2.description &&
    question1.isRequired === question2.isRequired &&
    question1.orderNum === question2.orderNum &&
    question1.accessibleForOtherPhases === question2.accessibleForOtherPhases &&
    // key will be ignored if accessibleForOtherPhases is false
    (question1.accessibleForOtherPhases ? question1.accessKey === question2.accessKey : true)

  if (!basicEqual) return false

  // Compare properties that only exist on certain types
  const placeholderEqual =
    'placeholder' in question1 && 'placeholder' in question2
      ? question1.placeholder === question2.placeholder
      : true

  const errorMessageEqual =
    'errorMessage' in question1 && 'errorMessage' in question2
      ? question1.errorMessage === question2.errorMessage
      : true

  if (!placeholderEqual || !errorMessageEqual) return false

  if (question1IsMultiSelect && question2IsMultiSelect) {
    return (
      question1.minSelect === question2.minSelect &&
      question1.maxSelect === question2.maxSelect &&
      JSON.stringify(question1.options) === JSON.stringify(question2.options)
    )
  }
  if (
    !question1IsMultiSelect &&
    !question2IsMultiSelect &&
    !question1IsFileUpload &&
    !question2IsFileUpload
  ) {
    // Type narrowing: both are ApplicationQuestionText
    const q1 = question1 as ApplicationQuestionText
    const q2 = question2 as ApplicationQuestionText
    return (
      q1.id === q2.id &&
      q1.allowedLength === q2.allowedLength &&
      q1.validationRegex === q2.validationRegex
    )
  }
  if (question1IsFileUpload && question2IsFileUpload) {
    // Type narrowing: both are ApplicationQuestionFileUpload
    const q1 = question1 as ApplicationQuestionFileUpload
    const q2 = question2 as ApplicationQuestionFileUpload
    return q1.allowedFileTypes === q2.allowedFileTypes && q1.maxFileSizeMB === q2.maxFileSizeMB
  }
  return false
}
