import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@tumaet/prompt-ui-components'

import { useCoursePhaseConfigStore } from '../../../../../zustand/useCoursePhaseConfigStore'

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
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

  const gradeSuggestionVisible = coursePhaseConfig?.gradeSuggestionVisible ?? true
  const actionItemsVisible = coursePhaseConfig?.actionItemsVisible ?? true
  const gradingSheetVisible = coursePhaseConfig?.gradingSheetVisible ?? false

  const getVisibilityDescription = () => {
    const visibleItems: string[] = []

    if (gradeSuggestionVisible) {
      visibleItems.push('grade suggestions')
    }
    if (actionItemsVisible) {
      visibleItems.push('action items')
    }
    if (gradingSheetVisible) {
      visibleItems.push('grading sheet')
    }

    if (visibleItems.length === 0) {
      return 'assessment information'
    }

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
          <AlertDialogTitle>Release Assessment Results?</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to release the assessment results to students? This will make{' '}
            {getVisibilityDescription()} for all final assessments ({completedAssessments}/
            {totalAssessments}) visible to students. This action cannot be undone.
          </AlertDialogDescription>
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
