import {
  devResetExports,
  ExportStatus,
  getAllExports,
} from '@core/network/queries/privacyStudentDataExport'
import {
  Button,
  ManagementPageHeader,
  PromptTable,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@tumaet/prompt-ui-components'
import { adminExportColumns } from '../shared/components/PrivacyExport/adminExportColumns'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Trash2 } from 'lucide-react'

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
      <ManagementPageHeader>Privacy Audit</ManagementPageHeader>
      <Tabs defaultValue='deletion' className='-mt-3'>
        <TabsList className='w-full'>
          <TabsTrigger value='deletion' className='flex-1 flex gap-1'>
            <Trash2 className='w-5 h-5' />
            Deletion
          </TabsTrigger>
          <TabsTrigger value='export' className='flex-1 flex gap-1'>
            <Download className='w-5 h-5' />
            Export
          </TabsTrigger>
        </TabsList>
        <TabsContent value='deletion'>
          <h2 className='text-lg font-semibold text-foreground mb-4 mt-2'>Deletion Requests</h2>
          <PromptTable data={[]} columns={[]} pageSize={20} />
        </TabsContent>
        <TabsContent value='export'>
          <h2 className='text-lg font-semibold text-foreground mb-4 mt-2'>Data Exports</h2>
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
              pageSize={20}
            />
          )}
        </TabsContent>
      </Tabs>

      {/* TODO: DEV ONLY - delete later */}
      <Button variant='destructive' size='sm' onClick={handleDevReset} className='mt-4'>
        [DEV] Reset all exports
      </Button>
    </div>
  )
}
