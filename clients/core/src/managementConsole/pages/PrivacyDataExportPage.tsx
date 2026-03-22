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
import { PrivacyExportDocument } from '../shared/components/PrivacyExport/PrivacyExportDoc'
import { PrivacyExportBanner } from '../shared/components/PrivacyExport/PrivacyExportBanner'
import { PrivacyExportConfirmationDialog } from '../shared/components/PrivacyExport/PrivacyExportConfirmDialog'
import { PrivacyExportRateLimitNotice } from '../shared/components/PrivacyExport/PrivacyExportRateLimitNotice'

function getExportIDFromLatest(latest: LatestExportResponse | undefined): string | undefined {
  if (latest?.status === 'exists') return latest.export.id
  return undefined
}

export function PrivacyDataExportPage() {
  const [exportConfirmDialogOpen, setExportConfirmDialogOpen] = useState(false)

  const latestExportQuery = useQuery({
    queryKey: ['data-export-latest'],
    queryFn: () => getLatestStudentDataExport(),
  })

  const exportQuery = useQuery({
    queryKey: ['data-export'],
    queryFn: () => requestStudentDataExport(),
    enabled: false,
  })
  // Prefer a freshly triggered export ID; fall back to the existing one from the server
  const exportID = exportQuery.data?.id ?? getExportIDFromLatest(latestExportQuery.data)

  const statusQuery = useQuery({
    queryKey: ['data-export-status', exportID],
    queryFn: () => getStudentDataExportStatus(exportID!),
    enabled: !!exportID,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status === ExportStatus.complete || status === ExportStatus.failed) return false
      return 3000
    },
  })

  const handleConfirmExport = () => {
    setExportConfirmDialogOpen(false)
    exportQuery.refetch()
  }

  const isPolling =
    statusQuery.data?.status !== ExportStatus.complete &&
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
          <p className='mb-6 text-gray-600'>
            Download a copy of all personal data stored about you in our systems.
          </p>
          <Button
            disabled={exportQuery.isLoading || isRateLimited}
            onClick={() => setExportConfirmDialogOpen(true)}
          >
            {exportQuery.isLoading && <Loader2 className='animate-spin mr-2 h-4 w-4' />}
            Request data export
          </Button>
        </>
      )}

      {statusQuery.data && (
        <div className='mt-8 space-y-4'>
          <PrivacyExportBanner inProgress={isPolling} privacyExport={statusQuery.data} />

          {statusQuery.data.documents.length > 0 && (
            <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3'>
              {statusQuery.data.documents.map((doc) => (
                <PrivacyExportDocument
                  key={doc.id}
                  exportId={statusQuery.data.id}
                  privacy_export_document={doc}
                />
              ))}
            </div>
          )}
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
