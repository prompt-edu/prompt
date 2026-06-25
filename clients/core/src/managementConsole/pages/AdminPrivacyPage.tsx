import {
  type AdminPrivacyExport,
  getAllExports,
} from '@core/network/queries/privacyStudentDataExport'
import {
  type AdminPrivacyDeletionRequest,
  DeletionRequestStatus,
  getAllDeletionRequests,
} from '@core/network/queries/privacyStudentDataDeletion'
import {
  ManagementPageHeader,
  PromptTable,
  type TableFilter,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from '@tumaet/prompt-ui-components'
import {
  adminExportColumns,
  exportStatusLabel,
} from '../shared/components/PrivacyExport/adminExportColumns'
import { getAdminExportActions } from '../shared/components/PrivacyExport/adminExportActions'
import {
  adminDeletionColumns,
  deletionRequestStatusLabel,
} from '../shared/components/PrivacyDeletion/adminDeletionColumns'
import { getAdminDeletionActions } from '../shared/components/PrivacyDeletion/adminDeletionActions'
import { PrivacyDeletionReviewDialog } from '../shared/components/PrivacyDeletion/PrivacyDeletionReviewDialog'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Trash2 } from 'lucide-react'
import { useState } from 'react'

export function AdminPrivacyPage() {
  const queryClient = useQueryClient()
  const [reviewing, setReviewing] = useState<AdminPrivacyDeletionRequest | null>(null)

  const allExportsQuery = useQuery({
    queryKey: ['privacy', 'admin', 'exports'],
    queryFn: getAllExports,
  })
  const allDeletionsQuery = useQuery({
    queryKey: ['privacy', 'admin', 'deletions'],
    queryFn: getAllDeletionRequests,
    refetchInterval: (query) =>
      query.state.data?.some((r) => r.status === DeletionRequestStatus.in_progress) ? 3000 : false,
  })

  return (
    <div>
      <ManagementPageHeader>Privacy</ManagementPageHeader>
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
          {allDeletionsQuery.isLoading && <p>Loading...</p>}
          {allDeletionsQuery.isSuccess && (
            <PromptTable<AdminPrivacyDeletionRequest>
              data={allDeletionsQuery.data}
              columns={adminDeletionColumns}
              actions={getAdminDeletionActions({ onReview: setReviewing })}
              onRowClick={(row) => {
                if (row.status === DeletionRequestStatus.pending_approval) setReviewing(row)
              }}
              filters={statusSelectFilter(deletionRequestStatusLabel)}
              pageSize={20}
            />
          )}
        </TabsContent>
        <TabsContent value='export'>
          <h2 className='text-lg font-semibold text-foreground mb-4 mt-2'>Data Exports</h2>
          {allExportsQuery.isLoading && <p>Loading...</p>}
          {allExportsQuery.isSuccess && (
            <PromptTable<AdminPrivacyExport>
              data={allExportsQuery.data}
              columns={adminExportColumns}
              actions={getAdminExportActions({ queryClient })}
              filters={statusSelectFilter(exportStatusLabel)}
              pageSize={20}
            />
          )}
          <p className='text-xs text-muted-foreground mt-4'>
            <span className='font-medium'>Archived</span> means the corresponding export files have
            been deleted from storage. The metadata row is kept for audit purposes.
          </p>
        </TabsContent>
      </Tabs>

      <PrivacyDeletionReviewDialog request={reviewing} onClose={() => setReviewing(null)} />
    </div>
  )
}

function statusSelectFilter<T extends string>(optionLabels: Record<T, string>): TableFilter[] {
  return [
    {
      type: 'select',
      id: 'status',
      label: 'Status',
      options: Object.keys(optionLabels),
      optionLabel: (value) => optionLabels[value as T],
      badge: {
        label: 'Status',
        displayValue: (filtervalue) => optionLabels[filtervalue as T],
      },
    },
  ]
}
