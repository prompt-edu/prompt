import { SettingsCard } from '@/components/SettingsCard'
import { CopyCourseDialog } from '@core/managementConsole/courseOverview/components/CopyCourseDialog'
import { ArchiveCourseConfirmationDialog } from '@core/managementConsole/shared/components/ArchiveCourseConfirmationDialog'
import { archiveCourses, unarchiveCourses } from '@core/network/mutations/updateCourseArchiveStatus'
import { deleteCourse } from '@core/network/mutations/deleteCourse'
import { useQueryClient, useMutation } from '@tanstack/react-query'
import { useCourseStore, Course } from '@tumaet/prompt-shared-state'
import { Button, DeleteConfirmation, useToast } from '@tumaet/prompt-ui-components'
import { CircleAlert } from 'lucide-react'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useCourseTemplate } from '../hooks/useCourseTemplate'

interface CourseDangeZoneActionProps {
  title: string
  description: string
  action: () => void
  label: string
  variant?: 'link' | 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost'
  disabled?: boolean
}

function CourseDangerZoneAction({
  title,
  description,
  action,
  label,
  variant = 'destructive',
  disabled = false,
}: CourseDangeZoneActionProps) {
  return (
    <div className='flex items-center justify-between px-3 py-4'>
      <div className='max-w-[70%]'>
        <h2 className='text-sm font-medium'>{title}</h2>
        <p className='text-sm text-muted-foreground mt-1'>{description}</p>
      </div>

      <Button variant={variant} onClick={action} disabled={disabled}>
        {label}
      </Button>
    </div>
  )
}
export default function CourseDangerZone() {
  const { courses } = useCourseStore()
  const { courseId } = useParams<{ courseId: string }>()
  const course = courses.find((c) => c.id === courseId) as Course | undefined
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [copyCourseDialogOpen, setCopyCourseDialogOpen] = useState(false)
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false)
  const [archiveDialogOpen, setArchiveDialogOpen] = useState(false)
  const { toast } = useToast()

  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { templateStatus, isLoading: isTemplateLoading } = useCourseTemplate(courseId ?? '')
  const isTemplate = templateStatus?.isTemplate || false

  const { mutate: mutateDeleteCourse } = useMutation({
    mutationFn: () => deleteCourse(courseId ?? ''),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['courses'] })
      navigate('/management/courses')
    },
    onError: () => {
      toast({
        title: 'Failed to Delete Course',
        description: 'Please try again later!',
        variant: 'destructive',
      })
    },
  })

  if (!course) {
    return <p>Invalid Course Id</p>
  }

  const handleDelete = (deleteConfirmed: boolean) => {
    if (deleteConfirmed) {
      mutateDeleteCourse()
    }
  }

  const handleArchiveConfirm = async () => {
    try {
      await archiveCourses([course.id])
      toast({ title: 'Archived Course' })
    } catch {
      toast({
        title: 'Failed to update course archive status',
        description: 'Please try again later!',
        variant: 'destructive',
      })
    }
  }

  const handleUnarchive = async () => {
    try {
      await unarchiveCourses([course.id])
      toast({ title: 'Unarchived Course' })
    } catch {
      toast({
        title: 'Failed to update course archive status',
        description: 'Please try again later!',
        variant: 'destructive',
      })
    }
  }

  return (
    <>
      <SettingsCard
        title='Danger Zone'
        description='Copy, Make Template, Archive or Delete the Course'
        icon={<CircleAlert />}
      >
        <div className='px-5 pb-5'>
          <div className='px-2 w-full flex flex-col border border-border rounded-md divide-y'>
            <CourseDangerZoneAction
              title='Copy Course'
              description='Create a copy of this course. You can change some of the copies properties. Leaves this course unaffected'
              action={() => setCopyCourseDialogOpen(true)}
              label='Copy'
              variant='default'
            />
            <CourseDangerZoneAction
              title='Make Template'
              description='Create a template version of this course that can be used to quickly create new courses with the same structure'
              action={() => setTemplateDialogOpen(true)}
              label={isTemplate ? 'Already a Template' : 'Make Template'}
              variant='default'
              disabled={isTemplateLoading || isTemplate}
            />
            <CourseDangerZoneAction
              title={`${course.archived ? 'Unarchive' : 'Archive'} Course`}
              description={
                course.archived
                  ? 'Unarchive this course to make it active again'
                  : 'The Course will no longer appear in the sidebar but is still retrievable at the Archived Courses page. This can be undone.'
              }
              action={course.archived ? handleUnarchive : () => setArchiveDialogOpen(true)}
              label={course.archived ? 'Unarchive' : 'Archive'}
              variant='default'
            />
            <CourseDangerZoneAction
              title='Delete Course'
              description='Permanently deletes the course. This action is non-reversible.'
              action={() => setDeleteDialogOpen(true)}
              label='Delete'
            />
          </div>
        </div>
      </SettingsCard>

      <ArchiveCourseConfirmationDialog
        courseID={course.id}
        isOpen={archiveDialogOpen}
        onOpenChange={setArchiveDialogOpen}
        onConfirm={handleArchiveConfirm}
      />

      {deleteDialogOpen && (
        <DeleteConfirmation
          isOpen={deleteDialogOpen}
          setOpen={setDeleteDialogOpen}
          deleteMessage='Are you sure you want to delete this course?'
          customWarning={`This action cannot be undone. All student associations with this course will be lost.
            If you want to keep the course data, consider archiving it instead.`}
          onClick={handleDelete}
        />
      )}

      {copyCourseDialogOpen && (
        <CopyCourseDialog
          courseId={courseId ?? ''}
          isOpen={copyCourseDialogOpen}
          onClose={() => setCopyCourseDialogOpen(false)}
          useTemplateCopy={false}
          createTemplate={false}
        />
      )}

      {templateDialogOpen && (
        <CopyCourseDialog
          courseId={courseId ?? ''}
          isOpen={templateDialogOpen}
          onClose={() => setTemplateDialogOpen(false)}
          useTemplateCopy={true}
          createTemplate={true}
        />
      )}
    </>
  )
}
