import { useParams } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Download, Loader2, FileCheck2 } from 'lucide-react'
import { useState } from 'react'

import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ManagementPageHeader,
  ErrorPage,
  useToast,
} from '@tumaet/prompt-ui-components'

import { getCertificateStatus } from '../network/queries/getCertificateStatus'
import { downloadOwnCertificate, triggerBlobDownload } from '../network/queries/downloadCertificate'

export const StudentOverviewPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [isDownloading, setIsDownloading] = useState(false)

  const {
    data: status,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['certificateStatus', phaseId],
    queryFn: () => getCertificateStatus(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const handleDownload = async () => {
    if (!phaseId) return

    setIsDownloading(true)
    try {
      const blob = await downloadOwnCertificate(phaseId)
      triggerBlobDownload(blob, 'certificate.pdf')
      // Refresh status after download
      queryClient.invalidateQueries({ queryKey: ['certificateStatus', phaseId] })
    } catch (error) {
      console.error('Failed to download certificate:', error)
      toast({
        title: 'Download failed',
        description: 'Failed to download your certificate. Please try again.',
        variant: 'destructive',
      })
    } finally {
      setIsDownloading(false)
    }
  }

  if (isError) {
    return <ErrorPage message='Error loading certificate status' onRetry={refetch} />
  }

  if (isLoading) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Course Certificate</ManagementPageHeader>
      <p className='text-muted-foreground'>Download your course completion certificate.</p>

      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2'>
            <FileCheck2 className='h-5 w-5' />
            Certificate Status
          </CardTitle>
          <CardDescription>
            {status?.available
              ? 'Your certificate is ready for download.'
              : 'Your certificate is not available yet.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {status?.available ? (
            <div className='space-y-4'>
              {status.hasDownloaded && status.lastDownload && (
                <p className='text-sm text-muted-foreground'>
                  Last downloaded: {new Date(status.lastDownload).toLocaleDateString()}
                  {status.downloadCount && status.downloadCount > 1 && (
                    <span className='ml-2'>({status.downloadCount} total downloads)</span>
                  )}
                </p>
              )}
              <Button onClick={handleDownload} disabled={isDownloading}>
                {isDownloading ? (
                  <>
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                    Downloading...
                  </>
                ) : (
                  <>
                    <Download className='mr-2 h-4 w-4' />
                    Download Certificate
                  </>
                )}
              </Button>
            </div>
          ) : (
            <p className='text-amber-600'>
              {status?.message ||
                'Your certificate is not available yet. Please wait for your instructor to configure the certificate template.'}
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
