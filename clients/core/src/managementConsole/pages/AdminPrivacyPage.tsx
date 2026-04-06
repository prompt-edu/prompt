import {
  devResetExports,
  ExportStatus,
  getAllExports,
} from '@core/network/queries/privacyStudentDataExport'
import { Button, ManagementPageHeader, PromptTable } from '@tumaet/prompt-ui-components'
import { adminExportColumns } from '../shared/components/PrivacyExport/adminExportColumns'
import { useQuery, useQueryClient } from '@tanstack/react-query'

export function AdminPrivacyPage() {
  const allExportsQuery = useQuery({
    queryKey: ['privacy', 'admin', 'exports'],
    queryFn: getAllExports,
  })

  const queryClient = useQueryClient()

  // DEV ONLY - delete later
  const handleDevReset = async () => {
    await devResetExports()
    queryClient.clear()
  }

  return (
    <div>
      {/* DEV ONLY - delete later */}
      <Button variant='destructive' size='sm' onClick={handleDevReset} className='mb-4'>
        [DEV] Reset all exports
      </Button>
      <ManagementPageHeader>Privacy Dashboard</ManagementPageHeader>
      <p className='mb-8 text-muted-foreground'>
        Monitor and manage privacy-related requests from users.
      </p>

      <div className='space-y-10'>
        <section>
          <h2 className='text-lg font-semibold text-foreground mb-4'>Deletion Requests</h2>
          <PromptTable data={[]} columns={[]} pageSize={10} />
        </section>
        <section>
          <h2 className='text-lg font-semibold text-foreground mb-4'>Data Exports</h2>
          {allExportsQuery.isLoading && <p>Loading...</p>}
          {allExportsQuery.isSuccess && (
            <PromptTable
              data={allExportsQuery.data}
              columns={adminExportColumns}
              filters={[
                {
                  type: 'select',
                  id: 'status',
                  label: 'Status',
                  options: Object.values(ExportStatus),
                },
              ]}
              pageSize={10}
            />
          )}
        </section>
      </div>
    </div>
  )
}
