import type { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import type { ApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/applicationAnswerMultiSelect'
import type { ApplicationAnswerText } from '@core/interfaces/application/applicationAnswer/text/applicationAnswerText'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  TooltipProvider,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { ApplicationAnswerItem } from './ApplicationAnswerItem'
import type { ApplicationQuestion } from './questionKind'

interface ApplicationAnswersProps {
  coursePhaseId: string
  questions: ApplicationQuestion[]
  answersMultiSelect: ApplicationAnswerMultiSelect[]
  answersText: ApplicationAnswerText[]
  answersFileUpload: ApplicationAnswerFileUpload[]
}

export const ApplicationAnswers = ({
  coursePhaseId,
  questions,
  answersMultiSelect,
  answersText,
  answersFileUpload,
}: ApplicationAnswersProps) => {
  const sortedQuestions = useMemo(
    () => [...questions].sort((a, b) => a.orderNum - b.orderNum),
    [questions],
  )
  const textById = useMemo(
    () => new Map(answersText.map((a) => [a.applicationQuestionID, a])),
    [answersText],
  )
  const multiSelectById = useMemo(
    () => new Map(answersMultiSelect.map((a) => [a.applicationQuestionID, a])),
    [answersMultiSelect],
  )
  const fileById = useMemo(
    () => new Map(answersFileUpload.map((a) => [a.applicationQuestionID, a])),
    [answersFileUpload],
  )

  return (
    <Card className='w-full' data-testid='application-answers'>
      <CardHeader>
        <CardTitle>Application Answers</CardTitle>
        <CardDescription>
          Review the applicant&apos;s responses to the application questions.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {sortedQuestions.length === 0 ? (
          <p className='text-sm text-muted-foreground'>No application questions configured.</p>
        ) : (
          <TooltipProvider>
            <div className='divide-y divide-border'>
              {sortedQuestions.map((question, index) => (
                <ApplicationAnswerItem
                  key={question.id}
                  coursePhaseId={coursePhaseId}
                  number={index + 1}
                  question={question}
                  textAnswer={textById.get(question.id)?.answer ?? ''}
                  multiSelectAnswer={multiSelectById.get(question.id)?.answer ?? []}
                  fileAnswer={fileById.get(question.id)}
                />
              ))}
            </div>
          </TooltipProvider>
        )}
      </CardContent>
    </Card>
  )
}
