import { Button } from '@tumaet/prompt-ui-components'
import { FileDown, Loader2 } from 'lucide-react'
import { useDownloadReportPdf } from '../hooks/useDownloadReportPdf'

export const DownloadPdfButton = ({ fileName }: { fileName: string }) => {
  const { downloadPdf, isDownloading } = useDownloadReportPdf()

  return (
    <Button
      variant='outline'
      className='gap-2'
      disabled={isDownloading}
      onClick={() => downloadPdf(fileName)}
    >
      {isDownloading ? (
        <Loader2 className='h-4 w-4 animate-spin' />
      ) : (
        <FileDown className='h-4 w-4' />
      )}
      Download PDF
    </Button>
  )
}
