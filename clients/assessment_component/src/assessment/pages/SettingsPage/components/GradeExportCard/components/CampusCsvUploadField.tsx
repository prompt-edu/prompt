import { Button, cn } from '@tumaet/prompt-ui-components'
import { FileText, Loader2, UploadCloud, X } from 'lucide-react'
import { type ChangeEvent, type DragEvent, useRef, useState } from 'react'

interface CampusCsvUploadFieldProps {
  onFileSelected: (file: File) => void
  isParsing: boolean
  fileName: string | null
  onClear: () => void
}

/**
 * Drag-and-drop CSV picker, modelled on the matching phase's upload button but
 * driven entirely by props so the parsed data can live wherever the caller wants
 * it.
 */
export const CampusCsvUploadField = ({
  onFileSelected,
  isParsing,
  fileName,
  onClear,
}: CampusCsvUploadFieldProps) => {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [dragActive, setDragActive] = useState(false)

  const handleDrag = (event: DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    setDragActive(event.type === 'dragenter' || event.type === 'dragover')
  }

  const handleDrop = (event: DragEvent) => {
    event.preventDefault()
    event.stopPropagation()
    setDragActive(false)

    const file = event.dataTransfer.files?.[0]
    if (file) {
      onFileSelected(file)
    }
  }

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    // Clear the input so selecting the same file again still fires a change event.
    event.target.value = ''
    if (file) {
      onFileSelected(file)
    }
  }

  return (
    <div className='space-y-3'>
      {/* The drop zone is the button: it is genuinely interactive, so it stays
          keyboard-operable and needs no nested control. */}
      <button
        type='button'
        onClick={() => fileInputRef.current?.click()}
        disabled={isParsing}
        className={cn(
          'w-full rounded-lg border-2 border-dashed p-6 text-center transition-colors',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          dragActive ? 'border-primary bg-primary/10' : 'border-muted-foreground/40',
          isParsing ? 'opacity-50' : 'hover:border-primary/60 hover:bg-muted/30',
        )}
        onDragEnter={isParsing ? undefined : handleDrag}
        onDragLeave={isParsing ? undefined : handleDrag}
        onDragOver={isParsing ? undefined : handleDrag}
        onDrop={isParsing ? undefined : handleDrop}
      >
        {isParsing ? (
          <Loader2 className='mx-auto mb-3 h-10 w-10 animate-spin text-muted-foreground' />
        ) : (
          <UploadCloud className='mx-auto mb-3 h-10 w-10 text-muted-foreground' />
        )}
        <span className='block text-sm text-muted-foreground'>
          {isParsing
            ? 'Reading file...'
            : 'Drag and drop the CampusOnline CSV here, or click to select'}
        </span>
        <span className='mt-1 block text-xs text-muted-foreground'>Supported file types: .csv</span>
      </button>
      <input
        type='file'
        ref={fileInputRef}
        onChange={handleChange}
        className='hidden'
        accept='.csv'
      />

      {fileName && (
        <div className='flex items-center justify-between gap-2 rounded-md border border-border bg-muted/20 px-3 py-2'>
          <span className='flex min-w-0 items-center gap-2 text-sm text-foreground'>
            <FileText className='h-4 w-4 shrink-0 text-muted-foreground' />
            <span className='truncate'>{fileName}</span>
          </span>
          <Button type='button' variant='ghost' size='sm' onClick={onClear} disabled={isParsing}>
            <X className='mr-1 h-4 w-4' />
            Remove
          </Button>
        </div>
      )}
    </div>
  )
}
