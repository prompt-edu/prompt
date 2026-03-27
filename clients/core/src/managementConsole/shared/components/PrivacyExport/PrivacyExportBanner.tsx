import { AlertTriangle, Download, Loader2, ShieldCheck } from 'lucide-react'
import {
  ExportStatus,
  getExportDocDownloadURL,
  PrivacyExport,
} from '@core/network/queries/privacyStudentDataExport'
import { Button, useToast } from '@tumaet/prompt-ui-components'
import { formatFileSize } from './formatFileSize'
import { useState } from 'react'

interface PrivacyExportBannerProps {
  inProgress: boolean
  privacyExport: PrivacyExport
}

function getTotalFileSize(privacyExport: PrivacyExport): number | null {
  let total = 0
  for (const doc of privacyExport.documents) {
    if (doc.file_size == null) return null
    total += doc.file_size
  }
  return total
}

export function PrivacyExportBanner({ inProgress, privacyExport }: PrivacyExportBannerProps) {
  const [downloading, setDownloading] = useState(-1)
  const completeDocs = privacyExport.documents.filter((doc) => doc.status == ExportStatus.complete)
  const { toast } = useToast()
  const totalSize = !inProgress ? getTotalFileSize(privacyExport) : null

  const handleDownloadAll = async () => {
    setDownloading(1)
    let failed = 0

    try {
      for (let i = 0; i < completeDocs.length; i++) {
        const expDoc = completeDocs[i]

        try {
          const downloadURL = await getExportDocDownloadURL(privacyExport.id, expDoc.id)
          window.location.assign(downloadURL)
        } catch {
          failed++
        }

        setDownloading(i + 1)

        if (i < completeDocs.length - 1) {
          await new Promise((resolve) => setTimeout(resolve, 1000))
        }
      }
    } finally {
      setDownloading(-1)
      if (failed > 0) {
        toast({
          title: 'Some downloads failed',
          description: `${failed} of ${privacyExport.documents.length} files could not be downloaded.`,
          variant: 'destructive',
        })
      }
    }
  }

  const isDownloading = downloading != -1
  const isFailed = !inProgress && privacyExport.status === ExportStatus.failed

  return (
    <div className='rounded-lg border border-border bg-muted p-4 flex items-center justify-between'>
      <div className='flex items-center gap-3'>
        <div className='lg:mx-2'>
          {inProgress ? (
            <Loader2 className='animate-spin h-5 w-5 text-muted-foreground' />
          ) : isFailed ? (
            <AlertTriangle className='h-5 w-5 text-muted-foreground' />
          ) : (
            <ShieldCheck className='h-5 w-5 text-muted-foreground' />
          )}
        </div>
        <div>
          <p className='font-semibold text-foreground'>
            {inProgress ? 'Collecting your data…' : isFailed ? 'There was a problem with your Export' : 'Export ready'}
          </p>
          <p className='text-xs text-muted-foreground mt-0.5'>
            {isFailed ? 'Review below' : `Requested on ${new Date(privacyExport.date_created).toLocaleString()}`}
          </p>
          {!inProgress && !isFailed && (
            <>
              <p className='text-xs text-muted-foreground mt-0.5'>
                Files available until {new Date(privacyExport.valid_until).toLocaleString()}
              </p>
              {totalSize != null && (
                <p className='text-xs text-muted-foreground mt-0.5'>~{formatFileSize(totalSize)}</p>
              )}
            </>
          )}
        </div>
      </div>
      {!inProgress && !isFailed && (
        <Button onClick={handleDownloadAll} disabled={isDownloading}>
          {isDownloading ? (
            <>
              <Loader2 className='animate-spin h-5 w-5 text-muted-foreground' />
              Downloading {downloading}/{completeDocs.length}
            </>
          ) : (
            <>
              <Download className='mr-2 h-4 w-4' />
              Download All
            </>
          )}
        </Button>
      )}
    </div>
  )
}
