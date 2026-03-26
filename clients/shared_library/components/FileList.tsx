import * as React from 'react'
import { useState } from 'react'
import { Download, Trash2, File } from 'lucide-react'
import { FileResponse } from '@/network/mutations/uploadFile'
import { Button, Card, CardContent, DeleteConfirmation } from '@tumaet/prompt-ui-components'
import { cn } from '../utils/cn'
import { formatFileSize } from '@/lib/formatFileSize'
import { openFileDownload } from '@/lib/openFileDownload'

export interface FileListProps {
  files: FileResponse[]
  onDelete?: (fileId: string) => void
  allowDelete?: boolean
  className?: string
}

export const FileList: React.FC<FileListProps> = ({
  files,
  onDelete,
  allowDelete = false,
  className,
}) => {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [fileToDelete, setFileToDelete] = useState<string | null>(null)

  const handleDownload = async (file: FileResponse) => {
    try {
      await openFileDownload({ downloadUrl: file.downloadUrl, fileName: file.originalFilename })
    } catch (error) {
      console.error('Download failed:', error)
    }
  }

  const promptDelete = (fileId: string) => {
    setFileToDelete(fileId)
    setDeleteDialogOpen(true)
  }

  const handleDeleteConfirmation = async (confirmed: boolean) => {
    if (confirmed && fileToDelete && onDelete) {
      try {
        await onDelete(fileToDelete)
      } catch (error) {
        console.error('Delete failed:', error)
      }
    }
    setFileToDelete(null)
  }

  const formatDate = (dateString: string): string => {
    const date = new Date(dateString)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
  }

  if (!files || files.length === 0) {
    return (
      <div className={cn('text-center py-8 text-gray-500', className)}>
        <File className='h-12 w-12 mx-auto mb-2 opacity-50' />
        <p className='text-sm'>No files uploaded yet</p>
      </div>
    )
  }

  return (
    <div className={cn('space-y-2', className)}>
      {files.map((file) => (
        <Card key={file.id}>
          <CardContent className='p-4'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-3 flex-1 min-w-0'>
                <div className='p-2 bg-primary/10 rounded'>
                  <File className='h-5 w-5 text-primary' />
                </div>
                <div className='flex-1 min-w-0'>
                  <p className='text-sm font-medium truncate'>{file.originalFilename}</p>
                  <div className='flex items-center gap-2 text-xs text-gray-500'>
                    <span>{formatFileSize(file.sizeBytes)}</span>
                    <span>•</span>
                    <span>{formatDate(file.createdAt)}</span>
                  </div>
                  {file.description && (
                    <p className='text-xs text-gray-600 mt-1'>{file.description}</p>
                  )}
                </div>
              </div>

              <div className='flex items-center gap-2'>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  onClick={() => handleDownload(file)}
                  disabled={!file.downloadUrl}
                >
                  <Download className='h-4 w-4' />
                </Button>

                {allowDelete && onDelete && (
                  <Button
                    type='button'
                    size='sm'
                    variant='outline'
                    onClick={() => promptDelete(file.id)}
                  >
                    <Trash2 className='h-4 w-4' />
                  </Button>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
      <DeleteConfirmation
        isOpen={deleteDialogOpen}
        setOpen={setDeleteDialogOpen}
        onClick={handleDeleteConfirmation}
        deleteMessage='Are you sure you want to delete this file?'
      />{' '}
    </div>
  )
}
