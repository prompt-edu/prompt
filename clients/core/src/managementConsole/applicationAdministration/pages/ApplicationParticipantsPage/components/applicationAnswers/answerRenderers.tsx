import type { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { getApplicationFileDownloadUrl } from '@core/network/queries/applicationFileDownloadUrl'
import { formatFileSize, openFileDownload } from '@tumaet/prompt-shared-state'
import { Badge, Button, useToast } from '@tumaet/prompt-ui-components'
import { Download, File } from 'lucide-react'
import { useState } from 'react'
import { AnswerBlock } from './AnswerBlock'

export const TextAnswer = ({ answer }: { answer: string }) => (
  <AnswerBlock isEmpty={!answer.trim()}>
    <p className='whitespace-pre-wrap wrap-break-word'>{answer}</p>
  </AnswerBlock>
)

export const MultiSelectAnswer = ({ answer }: { answer: string[] }) => (
  <AnswerBlock isEmpty={answer.length === 0}>
    <div className='flex flex-wrap gap-1.5'>
      {answer.map((item, index) => (
        <Badge key={`${item}-${index}`}>{item}</Badge>
      ))}
    </div>
  </AnswerBlock>
)

interface FileUploadAnswerProps {
  coursePhaseId: string
  answer?: ApplicationAnswerFileUpload
}

export const FileUploadAnswer = ({ coursePhaseId, answer }: FileUploadAnswerProps) => {
  const [isDownloading, setIsDownloading] = useState(false)
  const { toast } = useToast()

  if (!answer) {
    return <AnswerBlock isEmpty />
  }

  const handleDownload = async () => {
    if (!coursePhaseId || !answer.fileID) {
      return
    }

    try {
      setIsDownloading(true)
      const downloadUrl = await getApplicationFileDownloadUrl(coursePhaseId, answer.fileID)
      await openFileDownload({ downloadUrl, fileName: answer.fileName })
    } catch (error) {
      console.error('Failed to download file:', error)
      toast({
        title: 'Download failed',
        description: "Couldn't download the file. Please try again.",
        variant: 'destructive',
      })
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <AnswerBlock>
      <div className='flex items-center justify-between gap-3'>
        <div className='flex min-w-0 items-center gap-3'>
          <div className='flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-background'>
            <File className='h-4 w-4 text-muted-foreground' />
          </div>
          <div className='min-w-0'>
            <div className='truncate font-medium'>{answer.fileName}</div>
            <div className='text-xs text-muted-foreground'>{formatFileSize(answer.fileSize)}</div>
          </div>
        </div>
        <Button
          variant='outline'
          size='sm'
          className='shrink-0'
          onClick={handleDownload}
          disabled={!coursePhaseId || !answer.fileID || isDownloading}
        >
          <Download className='mr-2 h-4 w-4' />
          Download
        </Button>
      </div>
    </AnswerBlock>
  )
}
