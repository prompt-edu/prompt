import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { getExampleInfo } from '../network/queries/getExampleInfo'

export const SettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: exampleInfo,
    isPending: isExampleInfoPending,
    isError: isExampleInfoError,
    refetch: refetchExampleInfo,
  } = useQuery<string>({
    queryKey: ['exampleInfo', phaseId],
    queryFn: () => getExampleInfo(phaseId ?? ''),
  })

  if (isExampleInfoError)
    return (
      <ErrorPage onRetry={refetchExampleInfo} description='Could not fetch example information' />
    )
  if (isExampleInfoPending)
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )

  return (
    <div>
      <ManagementPageHeader>Example Component Settings</ManagementPageHeader>
      <p className='text-sm text-muted-foreground mb-4'>
        This is the settings page for the Example Component.
      </p>
      <Card className='w-full max-w-md'>
        <CardContent className='pt-6'>
          <p className='text-center'>{exampleInfo}</p>
        </CardContent>
      </Card>
    </div>
  )
}

export default SettingsPage
