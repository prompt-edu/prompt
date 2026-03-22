import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { useToast } from '@tumaet/prompt-ui-components'
import { copyCourse } from '@core/network/mutations/copyCourse'
import { checkCourseCopyable } from '@core/network/mutations/checkCourseCopyable'
import type { CopyCourse } from '../../managementConsole/courseOverview/interfaces/copyCourse'
import type { CopyCourseFormValues } from '@core/validations/copyCourse'
import type { DialogStep } from '../../managementConsole/courseOverview/interfaces/copyCourseDialogProps'
import { useKeycloak } from '@core/keycloak/useKeycloak'

export const useCopyCourse = (
  courseId: string,
  currentStep: DialogStep,
  onClose: () => void,
  setCurrentStep: (step: DialogStep) => void,
  createTemplate?: boolean,
) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { forceTokenRefresh } = useKeycloak()

  // Query to check if course is copyable
  const {
    data: copyabilityData,
    isLoading: isCheckingCopyability,
    error: copyabilityError,
  } = useQuery({
    queryKey: ['course-copyability', courseId],
    queryFn: () => checkCourseCopyable(courseId),
    enabled: currentStep === 'warning' && !!courseId,
  })

  const { mutate: mutateCopyCourse, isPending: isCopying } = useMutation({
    mutationFn: (courseData: CopyCourse) => {
      return copyCourse(courseId ?? '', courseData)
    },
    onSuccess: (data: string | undefined) => {
      toast({
        title: 'Successfully Copied Course',
      })
      forceTokenRefresh() // refresh token to get permission for new course
        .then(() => {
          // Invalidate course queries
          return queryClient.invalidateQueries({ queryKey: ['courses'] })
        })
        .then(() => {
          // Wait for courses to be refetched
          return queryClient.refetchQueries({ queryKey: ['courses'] })
        })
        .then(() => {
          // Close the window and navigate
          navigate(`/management/course/${data}`)
          onClose()
        })
        .catch((err) => {
          console.error('Error during token refresh or query invalidation:', err)
          return err
        })
    },
    onError: () => {
      toast({
        title: 'Failed to Copy Course',
        description: 'Please try again later!',
        variant: 'destructive',
      })
      setCurrentStep('form')
    },
  })

  const handleProceedWithCopy = (formData: CopyCourseFormValues) => {
    const copyData: CopyCourse = {
      name: formData.name,
      semesterTag: formData.semesterTag,
      startDate: formData.dateRange?.from ?? new Date(),
      endDate: formData.dateRange?.to ?? new Date(),
      template: createTemplate ? true : false,
      shortDescription: formData.shortDescription,
      longDescription: formData.longDescription || undefined,
    }
    mutateCopyCourse(copyData)
  }

  return {
    copyabilityData,
    isCheckingCopyability,
    copyabilityError,
    isCopying,
    handleProceedWithCopy,
    queryClient,
  }
}
