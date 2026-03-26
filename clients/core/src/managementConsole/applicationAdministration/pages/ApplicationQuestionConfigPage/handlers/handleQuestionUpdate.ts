import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'

export const handleQuestionUpdate = (
  updatedQuestion:
    | ApplicationQuestionText
    | ApplicationQuestionMultiSelect
    | ApplicationQuestionFileUpload,
  setApplicationQuestions: React.Dispatch<
    React.SetStateAction<
      (ApplicationQuestionText | ApplicationQuestionMultiSelect | ApplicationQuestionFileUpload)[]
    >
  >,
) => {
  setApplicationQuestions((prev) => {
    // save it in the right order to be able to use stringify comparison

    return prev.map((q) => (q.id === updatedQuestion.id ? updatedQuestion : q))
  })
}
