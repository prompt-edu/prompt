import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { Lock, Unlock } from 'lucide-react'

import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Textarea,
  Alert,
  AlertDescription,
} from '@tumaet/prompt-ui-components'
import { useAuthStore } from '@tumaet/prompt-shared-state'

import { useCoursePhaseConfigStore } from '../../../../zustand/useCoursePhaseConfigStore'
import { useStudentAssessmentStore } from '../../../../zustand/useStudentAssessmentStore'

import { ActionItem } from '../../../../interfaces/actionItem'

import { AssessmentCompletionDialog } from '../../../components/AssessmentCompletionDialog'
import { DeadlineBadge } from '../../../components/badges'

import { ActionItemPanel } from './components/ActionItemPanel'
import { GradeSuggestion } from './components/GradeSuggestion'

import { useCreateOrUpdateAssessmentCompletion } from './hooks/useCreateOrUpdateAssessmentCompletion'
import { useMarkAssessmentAsComplete } from './hooks/useMarkAssessmentAsComplete'
import { useUnmarkAssessmentAsCompleted } from './hooks/useUnmarkAssessmentAsCompleted'

import { validateGrade } from '../../../utils/gradeConfig'

interface AssessmentCompletionProps {
  readOnly?: boolean
  actionItems?: ActionItem[]
}

export const AssessmentCompletion = ({
  readOnly = false,
  actionItems,
}: AssessmentCompletionProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const deadline = coursePhaseConfig?.deadline || undefined

  const { courseParticipationID, assessmentCompletion } = useStudentAssessmentStore()

  const [generalRemarks, setGeneralRemarks] = useState(assessmentCompletion?.comment || '')
  const [gradeSuggestion, setGradeSuggestion] = useState(
    assessmentCompletion?.gradeSuggestion?.toString() || '',
  )

  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    setGeneralRemarks(assessmentCompletion?.comment || '')
    setGradeSuggestion(assessmentCompletion?.gradeSuggestion?.toString() || '')
  }, [assessmentCompletion])

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
    }
  }, [generalRemarks])

  const [dialogOpen, setDialogOpen] = useState(false)
  const [error, setError] = useState<string | undefined>(undefined)

  const { mutate: createOrUpdateCompletion, isPending: isCreatePending } =
    useCreateOrUpdateAssessmentCompletion(setError)
  const { mutate: markAsComplete, isPending: isMarkPending } = useMarkAssessmentAsComplete(setError)
  const { mutate: unmarkAsCompleted, isPending: isUnmarkPending } =
    useUnmarkAssessmentAsCompleted(setError)

  const handleButtonClick = () => {
    setError(undefined)
    setDialogOpen(true)
  }

  const isPending = isCreatePending || isMarkPending || isUnmarkPending

  const isDeadlinePassed = deadline ? new Date() > new Date(deadline) : false

  const { user } = useAuthStore()
  const userName = user ? `${user.firstName} ${user.lastName}` : 'Unknown User'

  const handleConfirm = () => {
    if (readOnly) {
      return
    }
    const handleCompletion = async () => {
      try {
        if (assessmentCompletion?.completed ?? false) {
          if (isDeadlinePassed) {
            setError('Cannot unmark assessment as completed: deadline has passed.')
            return
          }
          await unmarkAsCompleted(courseParticipationID ?? '')
        } else {
          const gradeValidation = validateGrade(gradeSuggestion)
          if (!gradeValidation.isValid) {
            setError(`Cannot complete assessment: ${gradeValidation.error}`)
            return
          }

          await markAsComplete({
            courseParticipationID: courseParticipationID ?? '',
            coursePhaseID: phaseId ?? '',
            comment: generalRemarks.trim(),
            gradeSuggestion: gradeValidation.value ?? 5.0,
            author: userName,
            completed: true,
          })
        }
        setDialogOpen(false)
      } catch {
        setError('An error occurred while updating the assessment completion status.')
      }
    }

    handleCompletion()
  }

  const handleSaveFormData = async (newRemarks: string, newGrade: string) => {
    if (readOnly) {
      return
    }
    if (newRemarks.trim() || newGrade) {
      try {
        const gradeValidation = validateGrade(newGrade)
        if (!gradeValidation.isValid) {
          setError(`Failed to save: ${gradeValidation.error}`)
          return
        }

        await createOrUpdateCompletion({
          courseParticipationID: courseParticipationID ?? '',
          coursePhaseID: phaseId ?? '',
          comment: newRemarks.trim(),
          gradeSuggestion: gradeValidation.value ?? 5.0,
          author: userName,
          completed: false,
        })

        setError(undefined)
      } catch (err) {
        console.error('Failed to save form data:', err)
        setError('Failed to save form data. Please try again.')
      }
    }
  }

  const isCompleted = readOnly || (assessmentCompletion?.completed ?? false)

  return (
    <div className='space-y-4'>
      <h1 className='text-xl font-semibold tracking-tight'>Assessment Summary</h1>

      <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
        <div className='grid grid-cols-1 gap-4'>
          <Card className='flex flex-col grow'>
            <CardHeader>
              <CardTitle>General Remarks</CardTitle>
              {coursePhaseConfig?.actionItemsVisible && !readOnly && (
                <p className='text-xs text-gray-500 dark:text-gray-400 mt-1'>
                  These remarks will be visible to the student once results are released.
                </p>
              )}
            </CardHeader>
            <CardContent className='flex flex-col grow'>
              <Textarea
                ref={textareaRef}
                placeholder='What did this person do particularly well?'
                className='w-full resize-none min-h-[100px] overflow-hidden'
                rows={4}
                value={generalRemarks}
                onChange={(e) => setGeneralRemarks(e.target.value)}
                onBlur={() => handleSaveFormData(generalRemarks, gradeSuggestion)}
                disabled={isCompleted}
              />
            </CardContent>
          </Card>

          <GradeSuggestion
            onGradeSuggestionChange={(value) => {
              setGradeSuggestion(value)
              handleSaveFormData(generalRemarks, value)
            }}
            readOnly={readOnly}
          />
        </div>

        <ActionItemPanel readOnly={readOnly} actionItems={actionItems} />
      </div>

      {error && !dialogOpen && !readOnly && (
        <Alert variant='destructive' className='mt-4'>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {!readOnly && (
        <>
          <div className='flex justify-between items-center mt-8'>
            <div className='flex flex-col'>{deadline && <DeadlineBadge deadline={deadline} />}</div>

            <Button
              size='sm'
              disabled={isPending || (assessmentCompletion?.completed && isDeadlinePassed)}
              onClick={handleButtonClick}
            >
              {assessmentCompletion?.completed ? (
                <span className='flex items-center gap-1'>
                  <Unlock className='h-3.5 w-3.5' />
                  Unmark as Final
                </span>
              ) : (
                <span className='flex items-center gap-1'>
                  <Lock className='h-3.5 w-3.5' />
                  Mark Assessment as Final
                </span>
              )}
            </Button>
          </div>

          <AssessmentCompletionDialog
            completed={assessmentCompletion?.completed ?? false}
            completedAt={
              assessmentCompletion?.completedAt
                ? new Date(assessmentCompletion.completedAt)
                : undefined
            }
            author={assessmentCompletion?.author}
            isPending={isPending}
            dialogOpen={dialogOpen}
            setDialogOpen={setDialogOpen}
            error={error}
            setError={setError}
            handleConfirm={handleConfirm}
            isDeadlinePassed={isDeadlinePassed}
          />
        </>
      )}
    </div>
  )
}
