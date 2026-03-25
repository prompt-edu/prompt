import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { Check, MessageSquare, User, X } from 'lucide-react'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Button,
  Separator,
  TooltipProvider,
  MinimalTiptapEditor,
  ScoreLevelSelector,
  Input,
  Label,
} from '@tumaet/prompt-ui-components'
import { useCoursePhaseStore } from '../zustand/useCoursePhaseStore'
import { useParticipationStore } from '../zustand/useParticipationStore'
import type { InterviewQuestion } from '../interfaces/InterviewQuestion'
import type { InterviewAnswer } from '../interfaces/InterviewAnswer'
import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  PassStatus,
  ScoreLevel,
  useAuthStore,
} from '@tumaet/prompt-shared-state'
import { useUpdateCoursePhaseParticipation } from '@/hooks/useUpdateCoursePhaseParticipation'

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
  const { participations } = useParticipationStore()
  const participation = participations.find((p) => p.student.id === studentId)
  const restrictedData = (participation?.restrictedData ?? {}) as Record<string, unknown>

  const { user } = useAuthStore()
  const { coursePhase } = useCoursePhaseStore()
  const interviewQuestions =
    (coursePhase?.restrictedData?.interviewQuestions as InterviewQuestion[]) ?? []

  const [answers, setAnswers] = useState<InterviewAnswer[]>(
    (participation?.restrictedData?.interviewAnswers as InterviewAnswer[]) ?? [],
  )
  const [score, setScore] = useState<number | undefined>(undefined)
  const [interviewer, setInterviewer] = useState<string | undefined>(
    participation?.restrictedData?.interviewer,
  )

  const { mutate, isPending } = useUpdateCoursePhaseParticipation()

  // Compare current state with original data
  const isModified =
    answers.some((a) => {
      const originalAnswer = participation?.restrictedData?.interviewAnswers?.find(
        (oa: InterviewAnswer) => oa.questionID === a.questionID,
      )
      return originalAnswer?.answer !== a.answer
    }) ||
    score !== participation?.restrictedData?.score ||
    interviewer !== participation?.restrictedData?.interviewer

  useEffect(() => {
    if (participation && coursePhase) {
      const interviewAnswers = participation.restrictedData?.interviewAnswers as InterviewAnswer[]
      setAnswers(interviewAnswers ?? [])

      const interviewScore = participation.restrictedData?.score as number | undefined
      setScore(interviewScore)

      const newInterviewer = participation.restrictedData?.interviewer as string | undefined
      setInterviewer(newInterviewer)
    }
  }, [participation, coursePhase])

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

  const saveChanges = (
    passStatus?: PassStatus,
    overrides?: {
      answers?: InterviewAnswer[]
      score?: number
      interviewer?: string
    },
  ) => {
    if (participation && coursePhase) {
      mutate({
        coursePhaseID: coursePhase.id,
        courseParticipationID: participation.courseParticipationID,
        restrictedData: {
          ...restrictedData,
          interviewAnswers: overrides?.answers ?? answers,
          score: overrides?.score ?? score,
          interviewer: overrides?.interviewer ?? interviewer,
        },
        studentReadableData: {},
        passStatus: passStatus ?? participation.passStatus,
      })
    }
  }

  const saveRestrictedDataPatch = (patch: Record<string, unknown>) => {
    if (participation && coursePhase) {
      mutate({
        coursePhaseID: coursePhase.id,
        courseParticipationID: participation.courseParticipationID,
        restrictedData: patch,
        studentReadableData: {},
        passStatus: participation.passStatus,
      })
    }
  }

  const setInterviewerAsSelf = () => {
    if (user) {
      const nextInterviewer = `${user.firstName} ${user.lastName}`
      setInterviewer(nextInterviewer)
      saveChanges(undefined, { interviewer: nextInterviewer })
    }
  }

  // When an input loses focus, save changes if any modifications exist.
  const handleBlur = () => {
    if (isModified) {
      saveChanges()
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
                saveRestrictedDataPatch({ score: nextScore })
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
                onClick={() => saveChanges(PassStatus.FAILED)}
              >
                <X />
                Reject
              </Button>
              <Button
                variant={isPassed ? 'default' : 'outline'}
                size='sm'
                disabled={isPending}
                aria-pressed={isPassed}
                onClick={() => saveChanges(PassStatus.PASSED)}
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
                  editorClassName='focus:outline-none'
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
