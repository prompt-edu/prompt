import { unarchiveCourses, archiveCourses } from '@core/network/mutations/updateCourseArchiveStatus'
import { Button, Tooltip, TooltipContent, TooltipTrigger } from '@tumaet/prompt-ui-components'
import { Archive, ArchiveRestore } from 'lucide-react'

interface CourseArchiveButtonProps {
  archived: boolean
  courseID: string
}

export const handleArchive = async (archived: boolean, courseID: string): Promise<void> => {
  if (archived) {
    await unarchiveCourses([courseID])
  } else {
    await archiveCourses([courseID])
  }
}
export function CourseArchiveButton({ archived, courseID: courseId }: CourseArchiveButtonProps) {
  return (
    <Tooltip>
      <TooltipContent>{archived ? 'Unarchive' : 'Archive'} this course</TooltipContent>
      <TooltipTrigger>
        <Button
          variant='outline'
          onClick={() => handleArchive(archived, courseId)}
          className='shrink-0 focus-visible:ring-2 focus-visible:ring-offset-2'
          aria-label={archived ? 'Unarchive course' : 'Archive course'}
        >
          {archived ? (
            <ArchiveRestore className='w-6 h-6 text-gray-600 dark:text-gray-100' />
          ) : (
            <Archive className='w-6 h-6 text-gray-600 dark:text-gray-100' />
          )}
        </Button>
      </TooltipTrigger>
    </Tooltip>
  )
}
