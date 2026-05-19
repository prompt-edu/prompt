import React, { useCallback, useEffect } from 'react'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import type { CourseFormValues } from '@core/validations/course'
import { AddCourseProperties } from './AddCourseProperties'
import { AddCourseAppearance } from './AddCourseAppearance'
import type { CourseAppearanceFormValues } from '@core/validations/courseAppearance'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { PostCourse } from '../interfaces/postCourse'
import { postNewCourse } from '@core/network/mutations/postNewCourse'
import { useNavigate } from 'react-router-dom'
import { useKeycloak } from '@core/keycloak/useKeycloak'
import { DialogLoadingDisplay } from '@tumaet/prompt-ui-components'
import { DialogErrorDisplay } from '@tumaet/prompt-ui-components'

interface AddCourseDialogProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export const AddCourseDialog = ({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
}: AddCourseDialogProps) => {
  const [internalOpen, setInternalOpen] = React.useState(false)
  const [currentPage, setCurrentPage] = React.useState(1)
  const [coursePropertiesFormValues, setCoursePropertiesFormValues] =
    React.useState<CourseFormValues | null>(null)

  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const { forceTokenRefresh } = useKeycloak()

  // Use controlled or internal state
  const isOpen = controlledOpen !== undefined ? controlledOpen : internalOpen
  const setIsOpen = controlledOnOpenChange || setInternalOpen

  const { mutate, isPending, error, isError, reset } = useMutation({
    mutationFn: (course: PostCourse) => {
      return postNewCourse(course)
    },
    onSuccess: (data: string | undefined) => {
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
          setIsOpen(false)
          navigate(`/management/course/${data}`)
        })
        .catch((err) => {
          console.error('Error during token refresh or query invalidation:', err)
          return err
        })
    },
  })

  const onSubmit = (data: CourseAppearanceFormValues) => {
    const course: PostCourse = {
      name: coursePropertiesFormValues?.name || '',
      startDate: coursePropertiesFormValues?.dateRange?.from || new Date(),
      endDate: coursePropertiesFormValues?.dateRange?.to || new Date(),
      courseType: coursePropertiesFormValues?.courseType || '',
      ects: coursePropertiesFormValues?.ects || 0,
      semesterTag: coursePropertiesFormValues?.semesterTag || '',
      restrictedMetaData: {},
      studentReadableData: { icon: data.icon, 'bg-color': data.color },
      shortDescription: coursePropertiesFormValues?.shortDescription || '',
      longDescription: coursePropertiesFormValues?.longDescription || '',
      template: false,
    }
    mutate(course)
  }

  const handleNext = (data) => {
    setCoursePropertiesFormValues(data)
    setCurrentPage(2)
  }

  const handleBack = () => {
    setCurrentPage(1)
  }

  const handleCancel = useCallback(() => {
    setIsOpen(false)
  }, [setIsOpen])

  // makes sure that window is first closed, before resetting the form
  useEffect(() => {
    if (!isOpen) {
      reset()
      setCoursePropertiesFormValues(null)
      setCurrentPage(1)
    }
  }, [isOpen, reset])

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      handleCancel()
    } else {
      setIsOpen(open)
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      {controlledOpen === undefined}
      <DialogContent className='sm:max-w-[550px]'>
        {isPending ? (
          <DialogLoadingDisplay customMessage='Updating course data...' />
        ) : isError && (error as any)?.response?.status === 409 ? (
          <>
            <DialogHeader>
              <DialogTitle className='text-2xl font-bold text-center'>Add a New Course</DialogTitle>
            </DialogHeader>
            <Alert variant='destructive'>
              <AlertTitle>Name already taken</AlertTitle>
              <AlertDescription>
                A course with the name &quot;{coursePropertiesFormValues?.name}&quot; and semester
                tag &quot;{coursePropertiesFormValues?.semesterTag}&quot; already exists. Please
                choose a different name or semester tag.
              </AlertDescription>
            </Alert>
            <div className='flex justify-end pt-2'>
              <Button
                variant='outline'
                onClick={() => {
                  reset()
                  setCurrentPage(1)
                }}
              >
                Go Back
              </Button>
            </div>
          </>
        ) : isError ? (
          <DialogErrorDisplay error={error} />
        ) : (
          <>
            <DialogHeader>
              <DialogTitle className='text-2xl font-bold text-center'>Add a New Course</DialogTitle>
            </DialogHeader>
            {currentPage === 1 ? (
              <AddCourseProperties
                onNext={handleNext}
                onCancel={handleCancel}
                initialValues={coursePropertiesFormValues || undefined}
              />
            ) : (
              <AddCourseAppearance onBack={handleBack} onSubmit={onSubmit} />
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
