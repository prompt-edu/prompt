import type { QueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@tumaet/prompt-ui-components'
import { AlertTriangle, Info } from 'lucide-react'
import type { CopyabilityData } from '../interfaces/copyCourseDialogProps'

interface WarningStepProps {
  copyabilityData?: CopyabilityData
  isCheckingCopyability: boolean
  copyabilityError: Error | null
  isCopying: boolean
  courseId: string
  onBack: () => void
  onProceed: () => void
  queryClient: QueryClient
  useTemplateCopy?: boolean
  createTemplate?: boolean
}

export const WarningStep = ({
  copyabilityData,
  isCheckingCopyability,
  copyabilityError,
  isCopying,
  courseId,
  onBack,
  onProceed,
  queryClient,
  useTemplateCopy,
  createTemplate,
}: WarningStepProps) => {
  const getButtonLabel = (
    isCopyingButton: boolean,
    useTemplateCopyButton?: boolean,
    createTemplateButton?: boolean,
    context: 'ready' | 'warning' = 'ready',
  ): string => {
    if (isCopyingButton) {
      if (useTemplateCopyButton && !createTemplateButton) return 'Applying Template...'
      if (createTemplateButton) return 'Creating Template...'
      return 'Copying Course...'
    }

    if (context === 'warning') {
      if (useTemplateCopyButton && !createTemplateButton) return 'Proceed And Apply Template'
      if (createTemplateButton) return 'Proceed And Create Template'
      return 'Proceed And Copy Course'
    }

    // ready context
    if (useTemplateCopyButton && !createTemplateButton) return 'Apply Template'
    if (createTemplateButton) return 'Create Template'
    return 'Copy Course'
  }

  if (isCheckingCopyability) {
    return (
      <>
        <DialogHeader>
          <DialogTitle>Checking Compatibility</DialogTitle>
          <DialogDescription>
            {useTemplateCopy && !createTemplate
              ? 'Checking if all phases can be applied from the template...'
              : createTemplate
                ? 'Checking if all phases can be saved in the template...'
                : 'Checking if all phases can be copied...'}
          </DialogDescription>
        </DialogHeader>
        <div className='flex items-center justify-center py-8'>
          <div className='animate-spin rounded-full h-8 w-8 border-b-2 border-primary'></div>
        </div>
      </>
    )
  }

  if (copyabilityError) {
    return (
      <>
        <DialogHeader>
          <DialogTitle>Compatibility Check Failed</DialogTitle>
          <DialogDescription>
            {useTemplateCopy && !createTemplate
              ? 'Could not check if the template can be applied.'
              : createTemplate
                ? 'Could not check if the course can be made into a template.'
                : 'Could not check if the course can be copied.'}
          </DialogDescription>
        </DialogHeader>
        <Alert variant='destructive'>
          <AlertTriangle className='h-4 w-4' />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Something went wrong. Please try again.</AlertDescription>
        </Alert>
        <DialogFooter>
          <Button variant='outline' onClick={onBack}>
            Back
          </Button>
          <Button
            onClick={() =>
              queryClient.invalidateQueries({ queryKey: ['course-copyability', courseId] })
            }
          >
            Retry
          </Button>
        </DialogFooter>
      </>
    )
  }

  if (!copyabilityData) return null

  const { copyable, missingPhaseTypes = [] } = copyabilityData

  if (copyable) {
    return (
      <>
        <DialogHeader>
          <DialogTitle>
            {useTemplateCopy && !createTemplate
              ? 'Template Ready to Apply'
              : createTemplate
                ? 'Template Ready to Create'
                : 'Course Ready to Copy'}
          </DialogTitle>
          <DialogDescription>
            {useTemplateCopy && !createTemplate
              ? 'All phases and settings can be applied from the template.'
              : createTemplate
                ? 'All phases and settings will be saved in the template.'
                : 'All phases and settings will be copied.'}
          </DialogDescription>
        </DialogHeader>
        <Alert>
          <Info className='h-4 w-4' />
          <AlertTitle>All Good</AlertTitle>
          <AlertDescription>
            {useTemplateCopy && !createTemplate
              ? 'All phases and their settings will be applied automatically.'
              : createTemplate
                ? 'All phases and their settings will be saved in the template.'
                : 'All phases and their settings will be copied automatically.'}
          </AlertDescription>
        </Alert>
        <DialogFooter>
          <Button variant='outline' onClick={onBack}>
            Back
          </Button>
          <Button onClick={onProceed} disabled={isCopying}>
            {getButtonLabel(isCopying, useTemplateCopy, createTemplate, 'ready')}
          </Button>
        </DialogFooter>
      </>
    )
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {useTemplateCopy && !createTemplate
            ? 'Some Template Phases Need Setup'
            : createTemplate
              ? 'Some Template Phases Need Setup'
              : 'Some Copied Phases Need Setup'}
        </DialogTitle>
        <DialogDescription>
          {useTemplateCopy && !createTemplate
            ? 'Not all phases in this template include their settings. See details below.'
            : createTemplate
              ? 'Not all phases will keep their settings in the template. See details below.'
              : 'Not all copied phases will include their settings. See details below.'}
        </DialogDescription>
      </DialogHeader>
      <Alert className='bg-yellow-50 dark:bg-yellow-900/20 border-yellow-400 dark:border-yellow-600'>
        <AlertTriangle className='h-4 w-4 text-yellow-600 dark:text-yellow-400' />
        <AlertTitle className='text-yellow-800 dark:text-yellow-200'>
          Manual Setup Required
        </AlertTitle>
        <AlertDescription className='text-yellow-700 dark:text-yellow-300'>
          {useTemplateCopy && !createTemplate
            ? "These phases will be included in the template, but their settings won't be. You'll need to set them up manually afterwards:"
            : createTemplate
              ? "These phases will be in the template, but their settings won't be saved. You'll need to set them up manually afterwards:"
              : "These phases will be copied, but their settings won't be. You'll need to set them up manually afterwards:"}
          <ul className='list-disc list-inside mt-2'>
            {missingPhaseTypes.map((phaseType, index) => (
              <li key={index}>{phaseType}</li>
            ))}
          </ul>
        </AlertDescription>
      </Alert>
      <DialogFooter>
        <Button variant='outline' onClick={onBack}>
          Back
        </Button>
        <Button onClick={onProceed} disabled={isCopying}>
          {getButtonLabel(isCopying, useTemplateCopy, createTemplate, 'warning')}
        </Button>
      </DialogFooter>
    </>
  )
}
