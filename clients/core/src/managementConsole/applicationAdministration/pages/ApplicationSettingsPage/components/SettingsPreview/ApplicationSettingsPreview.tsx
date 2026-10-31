import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { ApplicationPreview } from '@core/publicPages/application/pages/ApplicationPreview/ApplicationPreview'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ApplicationForm } from '../../../../interfaces/form/applicationForm'

interface ApplicationSettingsPreviewProps {
  welcomeText?: string
}

export const ApplicationSettingsPreview = ({ welcomeText }: ApplicationSettingsPreviewProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data: applicationForm } = useQuery<ApplicationForm>({
    queryKey: coreKeys.applications.form(phaseId),
    queryFn: () => coreApi.applications.form(phaseId ?? 'undefined'),
  })

  if (!applicationForm) {
    return null
  }

  return (
    <ApplicationPreview
      questionsText={applicationForm.questionsText}
      questionsMultiSelect={applicationForm.questionsMultiSelect}
      questionsFileUpload={applicationForm.questionsFileUpload}
      welcomeText={welcomeText}
    />
  )
}
