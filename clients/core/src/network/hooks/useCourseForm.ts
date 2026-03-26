import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import { type CopyCourseFormValues, copyCourseSchema } from '@core/validations/copyCourse'
import type { DialogStep } from '../../managementConsole/courseOverview/interfaces/copyCourseDialogProps'

export const useCourseForm = (courseId: string, setCurrentStep: (step: DialogStep) => void) => {
  const { courses } = useCourseStore()
  const course = courses.find((c) => c.id === courseId)
  const [formData, setFormData] = useState<CopyCourseFormValues | null>(null)

  const form = useForm<CopyCourseFormValues>({
    resolver: zodResolver(copyCourseSchema),
    defaultValues: {
      name: course?.name,
      semesterTag: '',
      dateRange: {
        from: undefined,
        to: undefined,
      },
      shortDescription: course?.shortDescription ?? '',
      longDescription: course?.longDescription ?? '',
    },
  })

  const onFormSubmit = (data: CopyCourseFormValues) => {
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
