import {
  ExportStatus,
  PrivacyExportDocument as PrivacyExportDocumentType,
  getExportDocDownloadURL,
} from '@core/network/queries/privacyStudentDataExport'
import { Button, Card, CardContent, useToast } from '@tumaet/prompt-ui-components'
import { Download, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { PrivacyExportStatus } from './PrivacyExportStatusBadge'
import { formatFileSize } from './formatFileSize'

interface PrivacyExportDocumentProps {
  exportId: string
  privacy_export_document: PrivacyExportDocumentType
}

export function PrivacyExportDocument({
  exportId,
  privacy_export_document,
}: PrivacyExportDocumentProps) {
  const isComplete = privacy_export_document.status === ExportStatus.complete
  const [isDownloading, setIsDownloading] = useState(false)
  const { toast } = useToast()

  const handleDownload = async () => {
    setIsDownloading(true)
    try {
      const url = await getExportDocDownloadURL(exportId, privacy_export_document.id)
      window.location.assign(url)
    } catch {
      toast({
        title: 'Download failed',
        description: `Could not download ${privacy_export_document.source_name}. Please try again.`,
        variant: 'destructive',
      })
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <Card className='border border-border'>
      <CardContent className='p-4'>
        <div className='flex items-center gap-3'>
          <div className='shrink-0'>
            <PrivacyExportStatus privacy_export_status={privacy_export_document.status} />
          </div>
          <div className='flex-1 min-w-0'>
            <p className='font-semibold text-foreground truncate'>
              {privacy_export_document.source_name}
            </p>
            <p className='text-xs text-muted-foreground mt-0.5'>
              {new Date(privacy_export_document.date_created).toLocaleString('de-DE')}
            </p>
          </div>
          <Button
            size='sm'
            disabled={!isComplete || isDownloading}
            variant='outline'
            onClick={handleDownload}
            className='shrink-0'
          >
            {isDownloading ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : (
              <Download className='h-4 w-4' />
            )}
            {privacy_export_document.file_size != null && (
              <span className='text-xs'>{formatFileSize(privacy_export_document.file_size)}</span>
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
