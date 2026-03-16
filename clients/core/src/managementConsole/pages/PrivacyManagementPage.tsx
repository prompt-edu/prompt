import { requestStudentDataExport } from '@core/network/queries/privacyStudentDataExport'
import { useQuery } from '@tanstack/react-query'
import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'

export function PrivacyManagementPage() {
  const exportQuery = useQuery({
    queryKey: ['student-data-export'],
    queryFn: () => requestStudentDataExport(),
    enabled: false,
  })

  return (
    <div>
      <ManagementPageHeader>Privacy</ManagementPageHeader>
      <p className='mb-4'>Export all data that is stored on our systems and related to you</p>
      <Button disabled={exportQuery.isLoading} onClick={() => exportQuery.refetch()}>
        {exportQuery.isLoading && <Loader2 className='animate-spin' />}
        Request data export
      </Button>
      {exportQuery.isSuccess && (
        <pre className='mt-4'>
          <code>{JSON.stringify(exportQuery.data, null, 2)}</code>
        </pre>
      )}
    </div>
  )
}
