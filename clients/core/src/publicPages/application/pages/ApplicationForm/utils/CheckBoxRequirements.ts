import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'

export const checkCheckBoxQuestion = (question: ApplicationQuestionMultiSelect): boolean => {
  return (
    question.maxSelect === 1 &&
    question.options.length === 1 &&
    question.options[0] === 'Yes' &&
    question.placeholder === 'CheckBoxQuestion'
  )
}
