import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import type { DialogStep } from '../../managementConsole/courseOverview/interfaces/copyCourseDialogProps'
import { MakeTemplateCourseFormValues } from '@core/validations/makeTemplateCourse'
import { makeTemplateCourseSchema } from '@core/validations/makeTemplateCourse'

export const useTemplateForm = (courseId: string, setCurrentStep: (step: DialogStep) => void) => {
  const { courses } = useCourseStore()
  const course = courses.find((c) => c.id === courseId)
  const [formData, setFormData] = useState<MakeTemplateCourseFormValues | null>(null)

  const form = useForm<MakeTemplateCourseFormValues>({
    resolver: zodResolver(makeTemplateCourseSchema),
    defaultValues: {
      name: course?.name,
      semesterTag: 'template',
      shortDescription: course?.shortDescription ?? '',
      longDescription: course?.longDescription ?? '',
    },
  })

  const onFormSubmit = (data: MakeTemplateCourseFormValues) => {
    setFormData(data)
    setCurrentStep('warning')
  }

  const resetForm = () => {
    setFormData(null)
    form.reset()
  }

  return {
    form,
    formData,
    course,
    onFormSubmit,
    resetForm,
  }
}
