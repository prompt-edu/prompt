import {
  ExportStatus,
  getLatestStudentDataExport,
  getStudentDataExportStatus,
  requestStudentDataExport,
  type LatestExportResponse,
} from '@core/network/queries/privacyStudentDataExport'
import { useQuery } from '@tanstack/react-query'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { PrivacyExportBanner } from '../shared/components/PrivacyExport/PrivacyExportBanner'
import { PrivacyExportDocumentList } from '../shared/components/PrivacyExport/PrivacyExportDocumentList'
import { PrivacyExportConfirmationDialog } from '../shared/components/PrivacyExport/PrivacyExportConfirmDialog'
import { PrivacyExportRateLimitNotice } from '../shared/components/PrivacyExport/PrivacyExportRateLimitNotice'
import { PrivacyExportTrigger } from '../shared/components/PrivacyExport/PrivacyExportTrigger'

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
        <PrivacyExportTrigger
          progressOngoing={exportQuery.isFetching}
          rateLimited={isRateLimited}
          openDialog={() => setExportConfirmDialogOpen(true)}
        />
      )}

      {statusQuery.data && (
        <div className='mt-8 space-y-4'>
          <PrivacyExportBanner inProgress={isPolling} privacyExport={statusQuery.data} />
          <PrivacyExportDocumentList privacyExport={statusQuery.data} inProgress={isPolling} />
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
