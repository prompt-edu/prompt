import { useState } from 'react'
import { Dialog, DialogContent } from '@tumaet/prompt-ui-components'
import { useCopyCourse } from '../../../network/hooks/useCopyCourse'
import { useCourseForm } from '../../../network/hooks/useCourseForm'
import { CopyCourseForm } from './CopyCourseForm'
import { WarningStep } from './WarningStep'
import type { CourseCopyDialogProps, DialogStep } from '../interfaces/copyCourseDialogProps'
import { MakeTemplateCourseForm } from './MakeTemplateCourseForm'
import { useTemplateForm } from '../../../network/hooks/useTemplateForm'

export const CopyCourseDialog = ({
  courseId,
  isOpen,
  onClose,
  useTemplateCopy,
  createTemplate,
}: CourseCopyDialogProps) => {
  const [currentStep, setCurrentStep] = useState<DialogStep>('form')

  const {
    copyabilityData,
    isCheckingCopyability,
    copyabilityError,
    isCopying,
    handleProceedWithCopy,
    queryClient,
  } = useCopyCourse(courseId, currentStep, onClose, setCurrentStep, createTemplate)

  const {
    form: copyForm,
    formData: copyFormData,
    course: copyCourse,
    onFormSubmit: onCopyFormSubmit,
    resetForm: resetCopyForm,
  } = useCourseForm(courseId, setCurrentStep)

  const {
    form: templateForm,
    formData: templateFormData,
    course: templateCourse,
    onFormSubmit: onTemplateFormSubmit,
    resetForm: resetTemplateForm,
  } = useTemplateForm(courseId, setCurrentStep)

  const handleClose = () => {
    setCurrentStep('form')
    resetCopyForm()
    resetTemplateForm()
    onClose()
  }

  const handleBackToForm = () => {
    setCurrentStep('form')
  }

  const handleProceed = () => {
    if (!copyFormData && !templateFormData) return
    if (createTemplate) {
      if (templateFormData) {
        const name = templateFormData.name
        const semesterTag = templateFormData.semesterTag || 'template'
        const dateRange = copyFormData?.dateRange || { from: new Date(), to: new Date() }
        handleProceedWithCopy({
          name,
          semesterTag,
          dateRange,
          shortDescription: templateFormData.shortDescription,
          longDescription: templateFormData.longDescription,
        })
      }
    } else {
      if (copyFormData) {
        handleProceedWithCopy(copyFormData)
      }
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose} aria-hidden='true'>
      <DialogContent className='max-w-md'>
        {currentStep === 'form' && !createTemplate && (
          <CopyCourseForm
            form={copyForm}
            courseName={copyCourse?.name || ''}
            onSubmit={onCopyFormSubmit}
            onClose={handleClose}
            useTemplateCopy={useTemplateCopy}
          />
        )}
        {currentStep === 'form' && createTemplate && (
          <MakeTemplateCourseForm
            form={templateForm}
            courseName={templateCourse?.name || ''}
            onSubmit={onTemplateFormSubmit}
            onClose={handleClose}
          />
        )}
        {currentStep === 'warning' && (
          <WarningStep
            copyabilityData={copyabilityData}
            isCheckingCopyability={isCheckingCopyability}
            copyabilityError={copyabilityError}
            isCopying={isCopying}
            courseId={courseId}
            onBack={handleBackToForm}
            onProceed={handleProceed}
            queryClient={queryClient}
            useTemplateCopy={useTemplateCopy}
            createTemplate={createTemplate}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}
