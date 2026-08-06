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

import { useGetCoursePhaseConfig } from '../../../../hooks/useGetCoursePhaseConfig'

interface UnreleaseConfirmationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
  isUnreleasing: boolean
  unreleaseError: string | null
}

export function UnreleaseConfirmationDialog({
  open,
  onOpenChange,
  onConfirm,
  isUnreleasing,
  unreleaseError,
}: UnreleaseConfirmationDialogProps) {
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? true

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Unrelease {assessmentEnabled ? 'Assessment' : 'Evaluation'} Results?
          </AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to hide the released{' '}
            {assessmentEnabled ? 'assessment' : 'evaluation'} results from students? Existing data
            and visibility settings will stay unchanged.
          </AlertDialogDescription>
          {unreleaseError && (
            <p className='text-sm font-medium text-destructive' role='alert'>
              {unreleaseError}
            </p>
          )}
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isUnreleasing}>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm} disabled={isUnreleasing}>
            {isUnreleasing ? 'Unreleasing...' : 'Unrelease Results'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
