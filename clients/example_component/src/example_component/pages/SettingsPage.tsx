import { useQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  ErrorPage,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'
import { getExampleInfo } from '../network/queries/getExampleInfo'

export const SettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const exampleInfoQuery = useQuery<string>({
    queryKey: ['exampleInfo', phaseId],
    queryFn: () => getExampleInfo(phaseId ?? ''),
  })

  return (
    <QueryGate
      queries={[exampleInfoQuery]}
      errorFallback={({ refetch }) => (
        <ErrorPage onRetry={refetch} description='Could not fetch example information' />
      )}
    >
      <div>
        <ManagementPageHeader>Example Component Settings</ManagementPageHeader>
        <p className='text-sm text-muted-foreground mb-4'>
          This is the settings page for the Example Component.
        </p>
        <Card className='w-full max-w-md'>
          <CardContent className='pt-6'>
            <p className='text-center'>{exampleInfoQuery.data}</p>
          </CardContent>
        </Card>
      </div>
    </QueryGate>
  )
}

export default SettingsPage
