import { Button, Card, CardContent, CardHeader, CardTitle } from '@tumaet/prompt-ui-components'
import { Loader2, UploadCloud } from 'lucide-react'
import { type ReactNode, useRef, useState } from 'react'
import { useMatchingStore } from '../zustand/useMatchingStore'

interface UploadButtonProps {
  title: string
  description: string
  icon: ReactNode
  onUploadFinish: () => void
  onUploadFunction: (file: File) => Promise<void>
  acceptedFileTypes: string[]
}

export const UploadButton = ({
  title,
  description,
  icon,
  onUploadFinish,
  onUploadFunction,
  acceptedFileTypes,
}: UploadButtonProps) => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragActive, setDragActive] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null) // State to store error messages

  const { setUploadedFile } = useMatchingStore()

  const handleUpload = async (file: File) => {
    setIsUploading(true)
    setError(null)
    setUploadedFile(file)
    setTimeout(async () => {
      try {
        await onUploadFunction(file)
        onUploadFinish()
      } catch (err: any) {
        console.error('Failed to parse file:', err)
        setError(err?.message ?? 'An unknown error occurred while parsing the file.')
        setUploadedFile(null)
      } finally {
        fileInputRef.current!.value = '' // Clear the file input
        setIsUploading(false)
      }
    }, 0)
  }

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
      setIsUploading(true)
      handleUpload(e.dataTransfer.files[0])
    }
  }

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      setIsUploading(true)
      handleUpload(file)
    }
  }

  return (
    <Card
      className={`hover:shadow-lg transition-all duration-300 ${dragActive ? 'border-primary' : ''}`}
      onDragEnter={isUploading ? undefined : handleDrag}
      onDragLeave={isUploading ? undefined : handleDrag}
      onDragOver={isUploading ? undefined : handleDrag}
      onDrop={isUploading ? undefined : handleDrop}
    >
      <CardHeader>
        <CardTitle className='flex items-center text-2xl'>
          {icon}
          <span className='ml-3'>{title}</span>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className='mb-6 text-muted-foreground'>{description}</p>
        <div
          className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
            dragActive ? 'border-primary bg-primary/10' : 'border-muted-foreground'
          } ${isUploading ? 'opacity-50' : ''}`}
        >
          <UploadCloud className='mx-auto h-12 w-12 text-muted-foreground mb-4 mt-4' />
          <p className='text-sm text-muted-foreground mb-2'>
            Drag and drop your file here, or click to select
          </p>
          <p className='text-xs text-muted-foreground mb-4'>
            Supported file types: {acceptedFileTypes.join(', ')}
          </p>
          <Button
            onClick={() => fileInputRef.current?.click()}
            className='text-lg py-6 mb-4'
            disabled={isUploading}
          >
            {isUploading ? (
              <>
                <Loader2 className='mr-2 h-4 w-4 animate-spin!' />
                Uploading...
              </>
            ) : (
              'Select File'
            )}
          </Button>
        </div>
        <input
          type='file'
          ref={fileInputRef}
          onChange={handleChange}
          className='hidden'
          accept={acceptedFileTypes.join(',')}
        />
        {/* Display error message if any */}
        {error && <p className='text-red-500 mt-4'>{error}</p>}
      </CardContent>
    </Card>
  )
}
