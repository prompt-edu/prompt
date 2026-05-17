import {
  ExportStatus,
  getLatestStudentDataExport,
  getStudentDataExportStatus,
  requestStudentDataExport,
  type LatestExportResponse,
} from '@core/network/queries/privacyStudentDataExport'
import { useQuery } from '@tanstack/react-query'
import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportBanner } from '../shared/components/PrivacyExport/PrivacyExportBanner'
import { PrivacyExportDocumentList } from '../shared/components/PrivacyExport/PrivacyExportDocumentList'
import { PrivacyExportConfirmationDialog } from '../shared/components/PrivacyExport/PrivacyExportConfirmDialog'
import { PrivacyExportRateLimitNotice } from '../shared/components/PrivacyExport/PrivacyExportRateLimitNotice'

function getExportIDFromLatest(latest: LatestExportResponse | undefined): string | undefined {
  if (latest?.status === 'exists') return latest.export.id
  return undefined
}

export function PrivacyDataExportPage() {
  const [exportConfirmDialogOpen, setExportConfirmDialogOpen] = useState(false)

  const latestExportQuery = useQuery({
    queryKey: ['privacy', 'data-export-latest'],
    queryFn: () => getLatestStudentDataExport(),
  })

  const exportQuery = useQuery({
    queryKey: ['privacy', 'data-export'],
    queryFn: () => requestStudentDataExport(),
    enabled: false,
  })
  // Prefer a freshly triggered export ID; fall back to the existing one from the server
  const exportID = exportQuery.data?.id ?? getExportIDFromLatest(latestExportQuery.data)

  const statusQuery = useQuery({
    queryKey: ['privacy', 'data-export-status', exportID],
    queryFn: () => getStudentDataExportStatus(exportID!),
    enabled: !!exportID,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (
        status === ExportStatus.complete ||
        status === ExportStatus.no_data ||
        status === ExportStatus.failed
      )
        return false
      return 3000
    },
  })

  const handleConfirmExport = () => {
    setExportConfirmDialogOpen(false)
    exportQuery.refetch()
  }

  const isPolling =
    statusQuery.data?.status !== ExportStatus.complete &&
    statusQuery.data?.status !== ExportStatus.no_data &&
    statusQuery.data?.status !== ExportStatus.failed &&
    !!exportID

  const isRateLimited = latestExportQuery.data?.status === 'rate_limited'

  return (
    <div>
      <ManagementPageHeader>Data Export</ManagementPageHeader>

      {isRateLimited && latestExportQuery.data?.status === 'rate_limited' && (
        <PrivacyExportRateLimitNotice
          until={latestExportQuery.data?.retry_after}
          className='mb-6'
        />
      )}

      {!statusQuery.data && (
        <>
          <p className='text-muted-foreground'>
            Download a copy of all personal data stored about you in our systems.
          </p>
          <p className='text-muted-foreground'>
            You must wait 30 days before requesting another export.
          </p>
          <div className='mt-6 flex flex-col items-start gap-2'>
            <Button
              disabled={exportQuery.isFetching || isRateLimited}
              onClick={() => setExportConfirmDialogOpen(true)}
            >
              {exportQuery.isFetching && <Loader2 className='animate-spin mr-2 h-4 w-4' />}
              Request data export
            </Button>
          </div>
        </>
      )}

      {statusQuery.data && (
        <div className='mt-8 space-y-4'>
          <PrivacyExportBanner inProgress={isPolling} privacyExport={statusQuery.data} />
          <PrivacyExportDocumentList privacyExport={statusQuery.data} />
        </div>
      )}

      <PrivacyExportConfirmationDialog
        isOpen={exportConfirmDialogOpen}
        setIsOpen={setExportConfirmDialogOpen}
        handleConfirm={handleConfirmExport}
      />
    </div>
  )
}
