import { useState } from 'react'
import { Check, Trash2, X } from 'lucide-react'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Input,
  Label,
  Separator,
} from '@tumaet/prompt-ui-components'
import { PassStatus, useAuthStore } from '@tumaet/prompt-shared-state'
import { InstructorComment } from '../../../../../interfaces/instructorComment'
import { useModifyAssessment } from '../hooks/mutateAssessment'
import { ApplicationAssessment } from '@core/managementConsole/applicationAdministration/interfaces/applicationAssessment'

interface AssessmentCardProps {
  score: number | null
  restrictedData: { [key: string]: any }
  acceptanceStatus: PassStatus
  courseParticipationID: string
}

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

  const { mutate: mutateAssessment } = useModifyAssessment(courseParticipationID)

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
        <div className='grid grid-cols-2'>
          <div>
            <Label htmlFor='new-score' className='text-sm font-medium'>
              Application Score
            </Label>
            <div className='flex items-center space-x-2 mt-1'>
              <Input
                id='new-score'
                title='Assessment Score'
                type='number'
                value={currentScore ?? ''}
                placeholder='New score'
                onChange={(e) =>
                  setCurrentScore(e.target.value === '' ? null : Number(e.target.value))
                }
                className='w-28 h-9'
              />
              <Button
                disabled={!currentScore || currentScore === score}
                onClick={() => handleScoreSubmit(currentScore ?? 0)}
                size='sm'
              >
                Submit
              </Button>
            </div>
          </div>
          <div>
            <Label htmlFor='resolution' className='text-sm font-medium'>
              Resolution
            </Label>
            <div className='flex items-center space-x-4 mt-1'>
              <Button
                variant='outline'
                size='sm'
                disabled={acceptanceStatus === PassStatus.FAILED}
                // className='border-red-500 text-red-500 hover:bg-red-50 hover:text-red-600'
                onClick={() => handleAcceptanceStatusChange(PassStatus.FAILED)}
              >
                <X />
                Reject
              </Button>
              <Button
                variant='default'
                size='sm'
                disabled={acceptanceStatus === PassStatus.PASSED}
                // className='bg-green-500 hover:bg-green-600 text-white'
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
