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
  useToast,
} from '@tumaet/prompt-ui-components'
import { AddCourseAppearance } from './AddCourseAppearance'
import type { CourseAppearanceFormValues } from '@core/validations/courseAppearance'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { PostCourse } from '../interfaces/postCourse'
import { postNewCourse } from '@core/network/mutations/postNewCourse'
import { useNavigate } from 'react-router-dom'
import { useKeycloak } from '@core/keycloak/useKeycloak'
import { DialogLoadingDisplay } from '@tumaet/prompt-ui-components'
import { DialogErrorDisplay } from '@tumaet/prompt-ui-components'
import { AddTemplateProperties } from './AddTemplateProperties'
import { TemplateFormValues } from '@core/validations/template'

interface AddTemplateDialogProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

export const AddTemplateDialog = ({
  open: controlledOpen,
  onOpenChange: controlledOnOpenChange,
}: AddTemplateDialogProps) => {
  const [internalOpen, setInternalOpen] = React.useState(false)
  const [currentPage, setCurrentPage] = React.useState(1)
  const [templatePropertiesFormValues, setTemplatePropertiesFormValues] =
    React.useState<TemplateFormValues | null>(null)

  const { toast } = useToast()

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
      toast({
        title: 'Template created successfully',
        description: 'You can now use this template for new courses.',
        variant: 'success',
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
          setIsOpen(false)
          navigate(`/management/course/${data}`)
        })
        .catch((err) => {
          console.error('Error during token refresh or query invalidation:', err)
          toast({
            title: 'Template created but navigation failed',
            description: 'Please refresh the page to see your new template.',
            variant: 'destructive',
          })
          return err
        })
    },
  })

  const onSubmit = (data: CourseAppearanceFormValues) => {
    const course: PostCourse = {
      name: templatePropertiesFormValues?.name || '',
      startDate: new Date(),
      endDate: new Date(),
      courseType: templatePropertiesFormValues?.courseType || '',
      ects: templatePropertiesFormValues?.ects || 0,
      semesterTag: templatePropertiesFormValues?.semesterTag || '',
      restrictedMetaData: {},
      studentReadableData: { icon: data.icon, 'bg-color': data.color },
      shortDescription: templatePropertiesFormValues?.shortDescription || '',
      longDescription: templatePropertiesFormValues?.longDescription || '',
      template: true,
    }
    mutate(course)
  }

  const handleNext = (data: TemplateFormValues) => {
    setTemplatePropertiesFormValues(data)
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
      setTemplatePropertiesFormValues(null)
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
      <DialogContent className='sm:max-w-[550px]'>
        {isPending ? (
          <DialogLoadingDisplay customMessage='Updating template data...' />
        ) : isError && (error as any)?.response?.status === 409 ? (
          <>
            <DialogHeader>
              <DialogTitle className='text-2xl font-bold text-center'>
                Add a New Template
              </DialogTitle>
            </DialogHeader>
            <Alert variant='destructive'>
              <AlertTitle>Name already taken</AlertTitle>
              <AlertDescription>
                A template with this name already exists. Please choose a different name.
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
              <DialogTitle className='text-2xl font-bold text-center'>
                Add a New Template
              </DialogTitle>
            </DialogHeader>
            {currentPage === 1 ? (
              <AddTemplateProperties
                onNext={handleNext}
                onCancel={handleCancel}
                initialValues={templatePropertiesFormValues || undefined}
              />
            ) : (
              <AddCourseAppearance onBack={handleBack} onSubmit={onSubmit} createTemplate={true} />
            )}
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
