import { useEffect, useState } from 'react'
import { Check, Trash2, X } from 'lucide-react'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Label,
  Separator,
  ScoreLevelSelector,
} from '@tumaet/prompt-ui-components'
import {
  mapNumberToScoreLevel,
  mapScoreLevelToNumber,
  PassStatus,
  ScoreLevel,
  useAuthStore,
} from '@tumaet/prompt-shared-state'
import { InstructorComment } from '../../../../../interfaces/instructorComment'
import { useModifyAssessment } from '../hooks/mutateAssessment'
import { ApplicationAssessment } from '@core/managementConsole/applicationAdministration/interfaces/applicationAssessment'

interface AssessmentCardProps {
  score: number | null
  restrictedData: { [key: string]: any }
  acceptanceStatus: PassStatus
  courseParticipationID: string
}

const SCORE_LEVEL_LABELS: Partial<Record<ScoreLevel, string>> = {
  [ScoreLevel.VeryGood]: 'Very Good',
  [ScoreLevel.Good]: 'Good',
  [ScoreLevel.Ok]: 'Okay',
  [ScoreLevel.Bad]: 'Bad',
  [ScoreLevel.VeryBad]: 'Very Bad',
}

const SCORE_LEVEL_DESCRIPTIONS: Record<ScoreLevel, string> = {
  [ScoreLevel.VeryGood]: 'Outstanding final application assessment.',
  [ScoreLevel.Good]: 'Strong application assessment with minor reservations.',
  [ScoreLevel.Ok]: 'Neutral application assessment.',
  [ScoreLevel.Bad]: 'Weak application assessment with notable concerns.',
  [ScoreLevel.VeryBad]: 'Do not recommend based on the application.',
}

const mapStoredScoreToScoreLevel = (score: number | null): ScoreLevel | undefined =>
  score !== null && Number.isInteger(score) && score >= 1 && score <= 5
    ? mapNumberToScoreLevel(score)
    : undefined

export const AssessmentCard = ({
  score,
  restrictedData,
  acceptanceStatus,
  courseParticipationID,
}: AssessmentCardProps) => {
  const [currentScore, setCurrentScore] = useState<number | null>(score)
  const comments = (restrictedData.comments as InstructorComment[]) ?? []
  const { user } = useAuthStore()
  const author = `${user?.firstName} ${user?.lastName}`

  const { mutate: mutateAssessment, isPending } = useModifyAssessment(courseParticipationID)

  useEffect(() => {
    setCurrentScore(score)
  }, [score])

  const handleScoreSubmit = (newScore: number) => {
    const assessment: ApplicationAssessment = {
      Score: newScore,
    }
    mutateAssessment(assessment)
  }

  const handleAcceptanceStatusChange = (newStatus: PassStatus) => {
    const assessment: ApplicationAssessment = {
      passStatus: newStatus,
    }
    mutateAssessment(assessment)
  }

  const handleDeleteComment = (comment: InstructorComment) => {
    const filteredComments = comments.filter(
      (c) => !(c.author == comment.author && c.timestamp === comment.timestamp),
    )
    console.log(filteredComments)
    const assessment: ApplicationAssessment = {
      restrictedData: {
        comments: filteredComments,
      },
    }
    mutateAssessment(assessment)
  }

  return (
    <Card className='w-full'>
      <CardHeader>
        <CardTitle>Assessment</CardTitle>
      </CardHeader>
      <CardContent>
        <div className='space-y-6'>
          <div>
            <Label className='text-sm font-medium'>Application Score</Label>
            <div className='mt-1 space-y-2'>
              <ScoreLevelSelector
                className='grid gap-2 xl:grid-cols-5'
                selectedScore={mapStoredScoreToScoreLevel(currentScore)}
                onScoreChange={(value) => {
                  if (isPending) {
                    return
                  }
                  const nextScore = mapScoreLevelToNumber(value)
                  setCurrentScore(nextScore)
                  handleScoreSubmit(nextScore)
                }}
                completed={false}
                descriptionsByLevel={SCORE_LEVEL_DESCRIPTIONS}
                labelsByLevel={SCORE_LEVEL_LABELS}
                showIndicators={false}
                hideUnselectedOnDesktop={false}
              />
              {currentScore !== null &&
                (!Number.isInteger(currentScore) || currentScore < 1 || currentScore > 5) && (
                  <p className='text-sm text-muted-foreground'>
                    Existing numeric score {currentScore} is outside the 1-5 selector range.
                    Selecting a level will replace it.
                  </p>
                )}
            </div>
          </div>
          <div>
            <Label htmlFor='resolution' className='text-sm font-medium'>
              Resolution
            </Label>
            <div className='mt-1 flex flex-wrap items-center gap-4'>
              <Button
                variant='outline'
                size='sm'
                disabled={isPending || acceptanceStatus === PassStatus.FAILED}
                onClick={() => handleAcceptanceStatusChange(PassStatus.FAILED)}
              >
                <X />
                Reject
              </Button>
              <Button
                variant='default'
                size='sm'
                disabled={isPending || acceptanceStatus === PassStatus.PASSED}
                onClick={() => handleAcceptanceStatusChange(PassStatus.PASSED)}
              >
                <Check />
                Accept
              </Button>
            </div>
          </div>
        </div>

        {comments && comments.length > 0 && (
          <>
            <Separator />
            <div className='space-y-2'>
              <Label className='text-sm font-medium'>Previous Comments</Label>
              <div className='space-y-2 max-h-60 overflow-y-auto'>
                {comments.map((comment, index) => (
                  <div
                    key={index}
                    className='border border-border p-3 rounded-md bg-secondary text-card-foreground flex justify-between items-start'
                  >
                    <div>
                      <p className='text-sm text-muted-foreground mb-1'>
                        <strong className='font-medium text-foreground'>{comment.author}</strong>{' '}
                        {comment.timestamp && (
                          <span className='text-muted-foreground'>
                            - {new Date(comment.timestamp).toLocaleString()}
                          </span>
                        )}
                      </p>
                      <p className='text-foreground whitespace-pre-line'>{comment.text}</p>
                    </div>
                    {comment.author === author && (
                      <Button
                        variant='ghost'
                        size='sm'
                        disabled={isPending}
                        onClick={() => handleDeleteComment(comment)}
                        className='text-red-500 hover:text-red-600 hover:bg-red-50'
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
