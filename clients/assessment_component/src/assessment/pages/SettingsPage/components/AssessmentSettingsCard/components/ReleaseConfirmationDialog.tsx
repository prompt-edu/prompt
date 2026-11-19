import { useGetMailingIsConfigured } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertTitle,
} from '@tumaet/prompt-ui-components'
import { EyeOff } from 'lucide-react'

import { useGetCoursePhaseConfig } from '../../../../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseMetaData } from '../../../../hooks/useGetCoursePhaseMetaData'
import { getVisibleResultSections } from '../../../../utils/studentResultsVisibility'
import { isTemplateComplete, parseResultsReleasedMail } from '../../ResultsReleasedMailCard/utils'

interface ReleaseConfirmationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
  isReleasing: boolean
  releaseError: string | null
  completedAssessments: number
  totalAssessments: number
}

export function ReleaseConfirmationDialog({
  open,
  onOpenChange,
  onConfirm,
  isReleasing,
  releaseError,
  completedAssessments,
  totalAssessments,
}: ReleaseConfirmationDialogProps) {
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: coursePhase } = useGetCoursePhaseMetaData()
  const resultsMail = parseResultsReleasedMail(coursePhase)
  const courseMailingIsConfigured = useGetMailingIsConfigured()
  const notifiesStudents =
    courseMailingIsConfigured && resultsMail.sendOnRelease && isTemplateComplete(resultsMail)

  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? true
  const visibleItems = getVisibleResultSections(coursePhaseConfig)
  const showsNothingToStudents = assessmentEnabled && visibleItems.length === 0

  const getVisibilityDescription = () => {
    if (visibleItems.length === 1) {
      return visibleItems[0]
    }

    if (visibleItems.length === 2) {
      return `${visibleItems[0]} and ${visibleItems[1]}`
    }

    return `${visibleItems[0]}, ${visibleItems[1]}, and ${visibleItems[2]}`
  }

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Release {assessmentEnabled ? 'Assessment' : 'Evaluation'} Results?
          </AlertDialogTitle>
          <AlertDialogDescription>
            {showsNothingToStudents ? (
              <>
                Are you sure you want to release the assessment results of all final assessments (
                {completedAssessments}/{totalAssessments}) to students?
              </>
            ) : assessmentEnabled ? (
              <>
                Are you sure you want to release the assessment results to students? This will make{' '}
                {getVisibilityDescription()} for all final assessments ({completedAssessments}/
                {totalAssessments}) visible to students.
              </>
            ) : (
              <>
                Are you sure you want to release the evaluation results to students? Each student
                will see their own self-evaluation and the averaged scores their peers gave them.
                Competencies rated by only one peer stay hidden.
              </>
            )}
            {notifiesStudents &&
              ' Students who have not been notified yet will also receive the results mail.'}
          </AlertDialogDescription>
          {showsNothingToStudents && (
            <Alert>
              <EyeOff className='h-4 w-4' />
              <AlertTitle>Students will not see any results</AlertTitle>
              <AlertDescription>
                The assessment sheet, grade suggestions, and action items are all hidden. Students
                only see a note that no results are shared for this phase. Switch on at least one of
                them under &quot;Student visibility after release&quot; to share results.
              </AlertDescription>
            </Alert>
          )}
          {releaseError && (
            <p className='text-sm font-medium text-destructive' role='alert'>
              {releaseError}
            </p>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isReleasing}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} disabled={isReleasing}>
            {isReleasing ? 'Releasing...' : 'Release Results'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
