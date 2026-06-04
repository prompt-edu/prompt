import {
  type AdminPrivacyExport,
  deleteExport,
  ExportStatus,
  getAllExports,
} from '@core/network/queries/privacyStudentDataExport'
import {
  ManagementPageHeader,
  PromptTable,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@tumaet/prompt-ui-components'
import {
  adminExportColumns,
  exportStatusLabel,
} from '../shared/components/PrivacyExport/adminExportColumns'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Trash2 } from 'lucide-react'

export function AdminPrivacyPage() {
  const queryClient = useQueryClient()
  const allExportsQuery = useQuery({
    queryKey: ['privacy', 'admin', 'exports'],
    queryFn: getAllExports,
  })

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
          <p className='-mt-4 mb-4'>Requester Identities are hidden for privacy reasons</p>
          {allExportsQuery.isLoading && <p>Loading...</p>}
          {allExportsQuery.isSuccess && (
            <PromptTable<AdminPrivacyExport>
              data={allExportsQuery.data}
              columns={adminExportColumns}
              actions={[
                {
                  label: 'Delete (Keeps metadata)',
                  icon: <Trash2 className='w-4 h-4' />,
                  onAction: async (rows) => {
                    await Promise.all(rows.map((r) => deleteExport(r.id)))
                    queryClient.invalidateQueries({ queryKey: ['privacy', 'admin', 'exports'] })
                  },
                  hide: (rows) => rows.every((r) => r.status === ExportStatus.archived),
                },
                {
                  label: 'Delete + reset rate limit',
                  icon: <Trash2 className='w-4 h-4' />,
                  onAction: async (rows) => {
                    await Promise.all(rows.map((r) => deleteExport(r.id, { resetRateLimit: true })))
                    queryClient.invalidateQueries({ queryKey: ['privacy', 'admin', 'exports'] })
                  },
                },
              ]}
              filters={[
                {
                  type: 'select',
                  id: 'status',
                  label: 'Status',
                  options: Object.values(ExportStatus),
                  optionLabel: (value) => exportStatusLabel[value as ExportStatus],
                },
              ]}
              pageSize={20}
            />
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
