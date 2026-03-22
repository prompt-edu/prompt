import { Download, Loader2, ShieldCheck } from 'lucide-react'
import {
  getExportDocDownloadURL,
  PrivacyExport,
} from '@core/network/queries/privacyStudentDataExport'
import { Button } from '@tumaet/prompt-ui-components'
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
  const totalSize = !inProgress ? getTotalFileSize(privacyExport) : null

  const handleDownloadAll = async () => {
    setDownloading(1)
    for (const expDoc of privacyExport.documents) {
      const downloadURL = await getExportDocDownloadURL(privacyExport.id, expDoc.id)
      window.open(downloadURL, '_blank')
      setDownloading((prev) => prev + 1)
      await new Promise((resolve) => setTimeout(resolve, 1000))
    }
    setDownloading(-1)
  }

  const isDownloading = downloading != -1

  return (
    <div className='rounded-lg border border-gray-200 bg-gray-50 p-4 flex items-center justify-between'>
      <div className='flex items-center gap-3'>
        <div className='lg:mx-2'>
          {inProgress ? (
            <Loader2 className='animate-spin h-5 w-5 text-gray-500' />
          ) : (
            <ShieldCheck className='h-5 w-5 text-gray-500' />
          )}
        </div>
        <div>
          <p className='font-semibold text-gray-900'>
            {inProgress ? 'Collecting your data…' : 'Export ready'}
          </p>
          <p className='text-xs text-gray-500 mt-0.5'>
            Requested on {new Date(privacyExport.date_created).toLocaleString()}
          </p>
          {!inProgress && (
            <>
              <p className='text-xs text-gray-500 mt-0.5'>
                Files available until {new Date(privacyExport.valid_until).toLocaleString()}
              </p>
              {totalSize != null && (
                <p className='text-xs text-gray-500 mt-0.5'>~{formatFileSize(totalSize)}</p>
              )}
            </>
          )}
        </div>
      </div>
      {!inProgress && (
        <Button onClick={handleDownloadAll} disabled={isDownloading}>
          {isDownloading ? (
            <>
              <Loader2 className='animate-spin h-5 w-5 text-gray-500' />
              Downloading {downloading}/{privacyExport.documents.length}
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
