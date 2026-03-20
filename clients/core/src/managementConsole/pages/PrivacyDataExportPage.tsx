import {
  ExportStatus,
  getLatestStudentDataExport,
  getStudentDataExportStatus,
  requestStudentDataExport,
} from '@core/network/queries/privacyStudentDataExport'
import { useQuery } from '@tanstack/react-query'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  Button,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportDocument } from '../shared/components/PrivacyExport/PrivacyExportDoc'
import { PrivacyExportBanner } from '../shared/components/PrivacyExport/PrivacyExportBanner'

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
  const exportID = exportQuery.data?.id ?? latestExportQuery.data?.id

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

  return (
    <div>
      <ManagementPageHeader>Data Export</ManagementPageHeader>

      {!statusQuery.data && (
        <>
          <p className='mb-6 text-gray-600'>
            Download a copy of all personal data stored about you in our systems.
          </p>
          <Button disabled={exportQuery.isLoading} onClick={() => setExportConfirmDialogOpen(true)}>
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

      <AlertDialog open={exportConfirmDialogOpen} onOpenChange={setExportConfirmDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Request Data Export</AlertDialogTitle>
            <AlertDialogDescription className='mt-2 space-y-1'>
              <p>You can only request a data export every 30 days.</p>
              <p>The export process might take a few minutes.</p>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmExport}>Request Data Export</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
