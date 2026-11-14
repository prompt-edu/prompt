import type { ApplicationProfilePictureConfig } from '@core/interfaces/application/openApplicationDetails'
import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { ApplicationPreview } from '@core/publicPages/application/pages/ApplicationPreview/ApplicationPreview'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'
import type { ApplicationForm } from '../../../../interfaces/form/applicationForm'

interface ApplicationSettingsPreviewProps {
  welcomeText?: string
  profilePicture: ApplicationProfilePictureConfig
}

export const ApplicationSettingsPreview = ({
  welcomeText,
  profilePicture,
}: ApplicationSettingsPreviewProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: applicationForm,
    isError,
    refetch,
  } = useQuery<ApplicationForm>({
    queryKey: coreKeys.applications.form(phaseId),
    queryFn: () => coreApi.applications.form(phaseId ?? 'undefined'),
  })

  if (isError) {
    return (
      <div className='flex items-center gap-2 text-sm text-destructive'>
        <span>Could not load the preview.</span>
        <Button variant='outline' size='sm' onClick={() => refetch()}>
          Retry
        </Button>
      </div>
    )
  }

  if (!applicationForm) {
    return null
  }

  return (
    <ApplicationPreview
      questionsText={applicationForm.questionsText}
      questionsMultiSelect={applicationForm.questionsMultiSelect}
      questionsFileUpload={applicationForm.questionsFileUpload}
      welcomeText={welcomeText}
      profilePicture={profilePicture}
    />
  )
}
