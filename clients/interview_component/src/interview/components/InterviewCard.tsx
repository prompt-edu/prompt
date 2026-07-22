import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  PassStatus,
  ScoreLevel,
  useAuthStore,
  useUpdateCoursePhaseParticipation,
} from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Input,
  Label,
  MinimalTiptapEditor,
  ScoreLevelSelector,
  Separator,
  TooltipProvider,
} from '@tumaet/prompt-ui-components'
import { Check, MessageSquare, User, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { InterviewAnswer } from '../interfaces/InterviewAnswer'
import type { InterviewQuestion } from '../interfaces/InterviewQuestion'
import { useUpdateInterviewReview } from '../network/hooks/useUpdateInterviewReview'
import { useCoursePhaseStore } from '../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../zustand/useParticipationStore'

const SCORE_LEVEL_LABELS: Partial<Record<ScoreLevel, string>> = {
  [ScoreLevel.VeryGood]: 'Very Good',
  [ScoreLevel.Good]: 'Good',
  [ScoreLevel.Ok]: 'Okay',
  [ScoreLevel.Bad]: 'Bad',
  [ScoreLevel.VeryBad]: 'Very Bad',
}

const SCORE_LEVEL_DESCRIPTIONS: Record<ScoreLevel, string> = {
  [ScoreLevel.VeryGood]: 'Strong final recommendation after the interview.',
  [ScoreLevel.Good]: 'Positive final recommendation with minor reservations.',
  [ScoreLevel.Ok]: 'Mixed interview outcome with a neutral final recommendation.',
  [ScoreLevel.Bad]: 'Weak interview outcome with notable concerns.',
  [ScoreLevel.VeryBad]: 'Do not recommend based on the interview outcome.',
}

const mapStoredScoreToScoreLevel = (score: number | undefined): ScoreLevel | undefined =>
  score !== undefined && Number.isInteger(score) && score >= 1 && score <= 5
    ? mapNumberToScoreLevel(score)
    : undefined

export const InterviewCard = () => {
  const { studentId } = useParams<{ studentId: string }>()
  const { participations, interviewReviews } = useParticipationStore()
  const participation = participations.find((p) => p.student.id === studentId)
  const review = participation ? interviewReviews[participation.courseParticipationID] : undefined

  const { user } = useAuthStore()
  const { coursePhase } = useCoursePhaseStore()
  const interviewQuestions =
    (coursePhase?.restrictedData?.interviewQuestions as InterviewQuestion[]) ?? []

  const [answers, setAnswers] = useState<InterviewAnswer[]>(review?.interviewAnswers ?? [])
  const [score, setScore] = useState<number | undefined>(review?.score)
  const [interviewer, setInterviewer] = useState<string | undefined>(review?.interviewer)

  const { mutate: mutatePassStatus, isPending: isPassStatusPending } =
    useUpdateCoursePhaseParticipation()
  const { mutate: mutateReview, isPending: isReviewPending } = useUpdateInterviewReview(
    coursePhase?.id,
  )
  const isPending = isPassStatusPending || isReviewPending

  // Compare current state with original data
  const isModified =
    answers.some((a) => {
      const originalAnswer = review?.interviewAnswers?.find(
        (oa: InterviewAnswer) => oa.questionID === a.questionID,
      )
      return originalAnswer?.answer !== a.answer
    }) ||
    score !== review?.score ||
    (interviewer ?? '') !== (review?.interviewer ?? '')

  useEffect(() => {
    if (participation && coursePhase) {
      setAnswers(review?.interviewAnswers ?? [])
      setScore(review?.score)
      setInterviewer(review?.interviewer)
    }
  }, [participation, coursePhase, review])

  const setAnswer = (questionID: number, answer: string) => {
    setAnswers((prevAnswers) => {
      const newAnswers = [...prevAnswers]
      const answerIndex = newAnswers.findIndex((a) => a.questionID === questionID)
      if (answerIndex === -1) {
        newAnswers.push({ questionID, answer })
      } else {
        newAnswers[answerIndex] = { questionID, answer }
      }
      return newAnswers
    })
  }

  // The interview review (score, interviewer, answers) is stored in the interview
  // microservice. The pass/fail resolution stays on the core participation.
  const saveReview = (overrides?: {
    answers?: InterviewAnswer[]
    score?: number
    interviewer?: string
  }) => {
    if (participation && coursePhase) {
      mutateReview({
        courseParticipationID: participation.courseParticipationID,
        review: {
          interviewAnswers: overrides?.answers ?? answers,
          score: overrides?.score ?? score,
          interviewer: overrides?.interviewer ?? interviewer ?? '',
        },
      })
    }
  }

  const savePassStatus = (passStatus: PassStatus) => {
    if (participation && coursePhase) {
      mutatePassStatus({
        coursePhaseID: coursePhase.id,
        courseParticipationID: participation.courseParticipationID,
        restrictedData: {},
        studentReadableData: {},
        passStatus,
      })
    }
  }

  const setInterviewerAsSelf = () => {
    if (user) {
      const nextInterviewer = `${user.firstName} ${user.lastName}`
      setInterviewer(nextInterviewer)
      saveReview({ interviewer: nextInterviewer })
    }
  }

  // When an input loses focus, save changes if any modifications exist.
  const handleBlur = () => {
    if (isModified) {
      saveReview()
    }
  }

  const isRejected = participation?.passStatus === PassStatus.FAILED
  const isPassed = participation?.passStatus === PassStatus.PASSED

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center'>
          <MessageSquare className='h-5 w-5 mr-2' />
          Interview
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className='mb-4 space-y-4'>
          <div className='space-y-2'>
            <Label>Interview Score</Label>
            <ScoreLevelSelector
              className='grid gap-2 xl:grid-cols-5'
              selectedScore={mapStoredScoreToScoreLevel(score)}
              onScoreChange={(value) => {
                if (isPending) {
                  return
                }
                const nextScore = mapScoreLevelToNumber(value)
                setScore(nextScore)
                saveReview({ score: nextScore })
              }}
              completed={false}
              descriptionsByLevel={SCORE_LEVEL_DESCRIPTIONS}
              labelsByLevel={SCORE_LEVEL_LABELS}
              showIndicators={false}
              hideUnselectedOnDesktop={false}
            />
            {score !== undefined && (!Number.isInteger(score) || score < 1 || score > 5) && (
              <p className='text-sm text-muted-foreground'>
                Existing numeric score {score} is outside the 1-5 selector range. Selecting a level
                will replace it.
              </p>
            )}
          </div>
          <div className='space-y-2'>
            <Label>Resolution</Label>
            <div className='flex flex-wrap gap-2'>
              <Button
                variant={isRejected ? 'destructive' : 'outline'}
                size='sm'
                disabled={isPending}
                aria-pressed={isRejected}
                onClick={() => savePassStatus(PassStatus.FAILED)}
              >
                <X />
                Reject
              </Button>
              <Button
                variant={isPassed ? 'default' : 'outline'}
                size='sm'
                disabled={isPending}
                aria-pressed={isPassed}
                onClick={() => savePassStatus(PassStatus.PASSED)}
              >
                <Check />
                Accept
              </Button>
            </div>
          </div>
        </div>
        <div className='space-y-2'>
          <div className='space-y-2'>
            <Label htmlFor='interviewer'>Interviewer</Label>
            <div className='flex space-x-2'>
              <Input
                id='interviewer'
                value={interviewer}
                onChange={(e) => setInterviewer(e.target.value)}
                onBlur={handleBlur}
                placeholder='Enter interviewer name'
              />
              <Button
                variant='outline'
                className='shrink-0'
                disabled={isPending || user === undefined}
                onClick={setInterviewerAsSelf}
              >
                <User className='h-4 w-4 mr-2' />
                Set as Self
              </Button>
            </div>
          </div>
        </div>
        {interviewQuestions.length > 0 && (
          <>
            <Separator className='mt-3 mb-3' />
            <Label className='text-lg'>Interview Questions</Label>
          </>
        )}
        {interviewQuestions
          .sort((a, b) => a.orderNum - b.orderNum)
          .map((question, index) => (
            <div key={question.id}>
              <Label>{question.question}</Label>
              <TooltipProvider>
                <MinimalTiptapEditor
                  value={answers.find((a) => a.questionID === question.id)?.answer ?? ''}
                  onChange={(value) => {
                    setAnswer(question.id, value as string)
                  }}
                  // Auto-save on blur
                  onBlur={handleBlur}
                  className='w-full'
                  editorContentClassName='p-3'
                  output='html'
                  placeholder='Type your answer here...'
                  editable={true}
                  editorClassName='focus:outline-hidden'
                />
              </TooltipProvider>
              {index < interviewQuestions.length - 1 && <Separator className='mt-3 mb-3' />}
            </div>
          ))}
        {/* The manual save button has been removed */}
      </CardContent>
    </Card>
  )
}
