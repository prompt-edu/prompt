import * as React from 'react'
import { Upload, X, File, Loader2 } from 'lucide-react'
import { useFileUpload } from '@/hooks/useFileUpload'
import { FileResponse } from '@/network/mutations/uploadFile'
import { Button, Progress, Card, CardContent } from '@tumaet/prompt-ui-components'
import { cn } from '../utils/cn'
import { formatFileSize } from '@/lib/formatFileSize'

export interface FileUploadProps {
  coursePhaseId?: string
  description?: string
  tags?: string
  accept?: string
  maxSizeMB?: number
  onSuccess?: (file: FileResponse) => void
  onError?: (error: Error) => void
  className?: string
}

export const FileUpload: React.FC<FileUploadProps> = ({
  coursePhaseId,
  description,
  tags,
  accept,
  maxSizeMB = 50,
  onSuccess,
  onError,
  className,
}) => {
  const [selectedFile, setSelectedFile] = React.useState<File | null>(null)
  const [dragActive, setDragActive] = React.useState(false)
  const fileInputRef = React.useRef<HTMLInputElement>(null)

  const { upload, uploadProgress, isUploading } = useFileUpload({
    onSuccess: (data) => {
      setSelectedFile(null)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
      onSuccess?.(data)
    },
    onError: (error) => {
      console.error('Upload failed:', error)
      onError?.(error)
    },
  })

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true)
    } else if (e.type === 'dragleave') {
      setDragActive(false)
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setDragActive(false)

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0])
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    e.preventDefault()
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0])
    }
  }

  const handleFile = (file: File) => {
    // Check file size
    const fileSizeMB = file.size / (1024 * 1024)
    if (fileSizeMB > maxSizeMB) {
      setSelectedFile(null)
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
      onError?.(new Error(`File size exceeds ${maxSizeMB}MB limit`))
      return
    }

    setSelectedFile(file)
    upload.mutate({
      file,
      coursePhaseId,
      description,
      tags,
    })
  }

  const handleRemove = () => {
    setSelectedFile(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className={cn('w-full', className)}>
      {!selectedFile ? (
        <div
          className={cn(
            'relative border-2 border-dashed rounded-lg p-8 text-center transition-colors',
            dragActive ? 'border-primary bg-primary/5' : 'border-gray-300 hover:border-primary',
            isUploading && 'pointer-events-none opacity-50',
          )}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
        >
          <input
            ref={fileInputRef}
            type='file'
            className='hidden'
            onChange={handleChange}
            accept={accept}
            disabled={isUploading}
          />

          <div className='flex flex-col items-center gap-4'>
            <div className='p-4 bg-primary/10 rounded-full'>
              <Upload className='h-8 w-8 text-primary' />
            </div>

            <div className='space-y-2'>
              <p className='text-sm font-medium'>
                Drag and drop your file here, or click to browse
              </p>
              <p className='text-xs text-gray-500'>Maximum file size: {maxSizeMB}MB</p>
            </div>

            <Button
              type='button'
              variant='outline'
              onClick={() => fileInputRef.current?.click()}
              disabled={isUploading}
            >
              Select File
            </Button>
          </div>
        </div>
      ) : (
        <Card>
          <CardContent className='p-4'>
            <div className='flex items-center justify-between'>
              <div className='flex items-center gap-3 flex-1'>
                <div className='p-2 bg-primary/10 rounded'>
                  <File className='h-5 w-5 text-primary' />
                </div>
                <div className='flex-1 min-w-0'>
                  <p className='text-sm font-medium truncate'>{selectedFile.name}</p>
                  <p className='text-xs text-gray-500'>{formatFileSize(selectedFile.size)}</p>
                </div>
              </div>

              <div className='flex items-center gap-2'>
                {!isUploading && (
                  <Button
                    type='button'
                    size='sm'
                    variant='ghost'
                    onClick={handleRemove}
                    disabled={isUploading}
                  >
                    <X className='h-4 w-4' />
                  </Button>
                )}
                {isUploading && <Loader2 className='h-5 w-5 animate-spin text-primary' />}
              </div>
            </div>

            {isUploading && uploadProgress > 0 && (
              <div className='mt-4'>
                <Progress value={uploadProgress} className='h-2' />
                <p className='text-xs text-gray-500 mt-1 text-right'>{uploadProgress}%</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
