import {
  ExportStatus,
  getLatestStudentDataExport,
  getStudentDataExportStatus,
  requestStudentDataExport,
  type LatestExportResponse,
} from '@core/network/queries/privacyStudentDataExport'
import { Button, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportBanner } from '../shared/components/PrivacyExport/PrivacyExportBanner'
import { PrivacyExportDocumentList } from '../shared/components/PrivacyExport/PrivacyExportDocumentList'
import { PrivacyExportConfirmationDialog } from '../shared/components/PrivacyExport/PrivacyExportConfirmDialog'
import { PrivacyExportRateLimitNotice } from '../shared/components/PrivacyExport/PrivacyExportRateLimitNotice'
import { usePrivacyRequestFlow } from '../shared/hooks/usePrivacyRequestFlow'

function isEndState(status: ExportStatus): boolean {
  return [ExportStatus.complete, ExportStatus.no_data, ExportStatus.failed].includes(status)
}

function getExportIDFromLatest(latest: LatestExportResponse | undefined): string | undefined {
  if (latest?.status === 'exists') return latest.export.id
  return undefined
}

export function PrivacyDataExportPage() {
  const [exportConfirmDialogOpen, setExportConfirmDialogOpen] = useState(false)

  const requestFlow = usePrivacyRequestFlow({
    resource: 'data-export',
    getLatest: getLatestStudentDataExport,
    extractIdFromLatest: getExportIDFromLatest,
    createRequest: requestStudentDataExport,
    getStatus: getStudentDataExportStatus,
    isEndState,
  })

  const handleConfirmExport = () => {
    setExportConfirmDialogOpen(false)
    requestFlow.triggerCreate()
  }

  const isRateLimited = requestFlow.latest?.status === 'rate_limited'
  const privacyExport = requestFlow.record

  return (
    <div>
      <ManagementPageHeader>Data Export</ManagementPageHeader>

      {isRateLimited && requestFlow.latest?.status === 'rate_limited' && (
        <PrivacyExportRateLimitNotice until={requestFlow.latest.retry_after} className='mb-6' />
      )}

      {!privacyExport && (
        <>
          <p className='text-muted-foreground'>
            Download a copy of all personal data stored about you in our systems.
          </p>
          <p className='text-muted-foreground'>
            You must wait 30 days before requesting another export.
          </p>
          <div className='mt-6 flex flex-col items-start gap-2'>
            <Button
              disabled={requestFlow.isCreating || isRateLimited}
              onClick={() => setExportConfirmDialogOpen(true)}
            >
              {requestFlow.isCreating && <Loader2 className='animate-spin mr-2 h-4 w-4' />}
              Request data export
            </Button>
          </div>
        </>
      )}

      {privacyExport && (
        <div className='mt-8 space-y-4'>
          <PrivacyExportBanner inProgress={requestFlow.isPolling} privacyExport={privacyExport} />
          <PrivacyExportDocumentList privacyExport={privacyExport} />
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
