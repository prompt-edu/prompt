import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { getTemplateInfo } from '../network/queries/getTemplateInfo'

export const SettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: templateInfo,
    isPending: isTemplateInfoPending,
    isError: isTemplateInfoError,
    refetch: refetchTemplateInfo,
  } = useQuery<string>({
    queryKey: ['templateInfo', phaseId],
    queryFn: () => getTemplateInfo(phaseId ?? ''),
  })

  if (isTemplateInfoError)
    return (
      <ErrorPage onRetry={refetchTemplateInfo} description='Could not fetch template information' />
    )
  if (isTemplateInfoPending)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div>
      <ManagementPageHeader>Template Component Settings</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        This is the settings page for the Template Component.
      </p>
      <Card className='w-full max-w-md'>
        <CardContent className='pt-6'>
          <p className='text-center'>{templateInfo}</p>
        </CardContent>
      </Card>
    </div>
  )
}

export default SettingsPage
