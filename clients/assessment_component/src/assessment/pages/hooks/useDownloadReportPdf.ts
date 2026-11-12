import { useToast } from '@tumaet/prompt-ui-components'
import { useState } from 'react'
import { downloadPrintReportsPdf } from '../utils/bulkExport/renderReportToPdf'

export const useDownloadReportPdf = () => {
  const { toast } = useToast()
  const [isDownloading, setIsDownloading] = useState(false)

  const downloadPdf = async (fileName: string) => {
    setIsDownloading(true)
    try {
      await downloadPrintReportsPdf(fileName)
    } catch {
      toast({
        title: 'Export failed',
        description: 'The PDF could not be created. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsDownloading(false)
    }
  }

  return { downloadPdf, isDownloading }
}
