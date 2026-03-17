import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useQuery } from '@tanstack/react-query'
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
import { isAfter } from 'date-fns'
import { getCoursePhaseByID } from '@core/network/queries/coursePhase'

interface ArchiveCourseConfirmationDialogProps {
  courseID: string
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}

export function ArchiveCourseConfirmationDialog({
  courseID,
  isOpen,
  onOpenChange,
  onConfirm,
}: ArchiveCourseConfirmationDialogProps) {
  const { courses } = useCourseStore()
  const course = courses.find((c) => c.id === courseID)

  const applicationPhase = course?.coursePhases.find((p) => p.coursePhaseType === 'Application')

  const { data: applicationPhaseData } = useQuery({
    queryKey: ['coursePhase', applicationPhase?.id],
    queryFn: () => getCoursePhaseByID(applicationPhase!.id),
    enabled: isOpen && !!applicationPhase,
  })

  const applicationEndDate = applicationPhaseData?.restrictedData?.['applicationEndDate']
    ? new Date(applicationPhaseData.restrictedData['applicationEndDate'])
    : undefined

  const hasActiveApplicationPhase = !!applicationEndDate && isAfter(applicationEndDate, new Date())

  return (
    <AlertDialog open={isOpen} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Archive this course?</AlertDialogTitle>
          <AlertDialogDescription className='mt-2'>
            {hasActiveApplicationPhase && (
              <>
                <p className='font-semibold text-black'>
                  This course has an active or upcoming application phase.
                </p>
                <p>Archiving it will prevent students from submitting applications.</p>
              </>
            )}
            <p>
              The course will no longer appear in the sidebar but is still retrievable at the
              Archived Courses page.
            </p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction onClick={onConfirm}>Archive</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
