import { useState } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@tumaet/prompt-ui-components'
import { useQuery } from '@tanstack/react-query'
import { getTemplateCourses } from '@core/network/queries/getTemplateCourses'
import { CopyCourseDialog } from '@managementConsole/courseOverview/components/CopyCourseDialog'
import { DialogLoadingDisplay } from '@tumaet/prompt-ui-components'
import { DialogErrorDisplay } from '@tumaet/prompt-ui-components'
import { FileIcon, BookOpen } from 'lucide-react'
import { Course } from '@tumaet/prompt-shared-state'
import { CourseTemplateIcon } from './CourseTemplateIcon'
import { CourseTypeDetails } from '@tumaet/prompt-shared-state'

interface TemplateSelectionDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export const TemplateSelectionDialog = ({ open, onOpenChange }: TemplateSelectionDialogProps) => {
  const [selectedTemplate, setSelectedTemplate] = useState<Course | null>(null)
  const [showCopyDialog, setShowCopyDialog] = useState(false)

  const {
    data: templateCourses,
    isLoading,
    error,
    isError,
  } = useQuery({
    queryKey: ['template-courses'],
    queryFn: getTemplateCourses,
    enabled: open,
  })

  const handleTemplateSelect = (template: Course) => {
    setSelectedTemplate(template)
    onOpenChange(false)
    setShowCopyDialog(true)
  }

  const handleCopyDialogClose = () => {
    setShowCopyDialog(false)
    setSelectedTemplate(null)
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className='sm:max-w-[600px] max-h-[80vh]'>
          <DialogHeader>
            <DialogTitle className='text-2xl font-bold text-center'>Choose a Template</DialogTitle>
          </DialogHeader>
          {isLoading ? (
            <DialogLoadingDisplay customMessage='Loading templates...' />
          ) : isError ? (
            <DialogErrorDisplay error={error} />
          ) : (
            <div className='max-h-[60vh] overflow-y-auto px-2'>
              {templateCourses && templateCourses.length > 0 ? (
                <div className='grid gap-3 py-4'>
                  {templateCourses.map((template) => (
                    <div
                      key={template.id}
                      onClick={() => handleTemplateSelect(template)}
                      className={
                        'flex items-center gap-4 p-4 border rounded-lg cursor-pointer ' +
                        'hover:bg-accent hover:text-accent-foreground transition-colors duration-200'
                      }
                    >
                      <CourseTemplateIcon
                        iconName={template.studentReadableData?.['icon'] || 'graduation-cap'}
                        bgColor={template.studentReadableData?.['bg-color'] || 'bg-gray-100'}
                      />
                      <div className='flex flex-col flex-1 min-w-0'>
                        <span className='font-semibold text-lg truncate'>{template.name}</span>
                        <div className='flex items-center gap-4 text-sm text-muted-foreground mt-1'>
                          <div className='flex items-center gap-1'>
                            <BookOpen className='w-3 h-3' />
                            <span>{template.ects} ECTS</span>
                          </div>
                          <span className='text-xs bg-primary/10 text-primary px-2 py-1 rounded-full font-medium'>
                            {CourseTypeDetails[template.courseType].name}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className='text-center py-8'>
                  <FileIcon className='w-12 h-12 mx-auto text-muted-foreground mb-4' />
                  <p className='text-muted-foreground'>No templates available</p>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>
      {selectedTemplate && (
        <CopyCourseDialog
          courseId={selectedTemplate.id}
          isOpen={showCopyDialog}
          onClose={handleCopyDialogClose}
          useTemplateCopy={true}
          createTemplate={false}
        />
      )}
    </>
  )
}
