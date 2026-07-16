import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Download, File, Loader2, Trash2, Upload } from 'lucide-react'
import { useRef, useState } from 'react'
import type { PresentationSummary } from '../interfaces'
import { openMaterialDownload, presentationApi, uploadMaterial } from '../network'
import { formatFileSize, getErrorMessage } from '../utils'

const MAX_FILE_SIZE_BYTES = 50 * 1024 * 1024

interface MaterialsPanelProps {
  coursePhaseId: string
  presentation: PresentationSummary
  isStaff: boolean
}

export const MaterialsPanel = ({ coursePhaseId, presentation, isStaff }: MaterialsPanelProps) => {
  const inputRef = useRef<HTMLInputElement>(null)
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [uploadingFiles, setUploadingFiles] = useState<string[]>([])
  const [downloadingId, setDownloadingId] = useState<string>()
  const canManage = isStaff || new Date(presentation.startTime).getTime() > Date.now()

  const materialsQuery = useQuery({
    queryKey: ['presentation-materials', coursePhaseId, presentation.id],
    queryFn: () => presentationApi.getMaterials(coursePhaseId, presentation.id),
    enabled: Boolean(coursePhaseId && presentation.id),
  })

  const deleteMutation = useMutation({
    mutationFn: (materialId: string) =>
      presentationApi.deleteMaterial(coursePhaseId, presentation.id, materialId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ['presentation-materials', coursePhaseId, presentation.id],
      })
      void queryClient.invalidateQueries({
        queryKey: ['presentations', coursePhaseId],
      })
      toast({ title: 'Material deleted' })
    },
    onError: (error) => {
      toast({
        title: 'Could not delete material',
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    },
  })

  const handleFiles = async (files: FileList | null) => {
    const selectedFiles = Array.from(files ?? [])
    if (selectedFiles.length === 0) return
    const oversized = selectedFiles.find((file) => file.size > MAX_FILE_SIZE_BYTES)
    if (oversized) {
      toast({
        title: 'File is too large',
        description: `${oversized.name} exceeds the 50 MB limit.`,
        variant: 'destructive',
      })
      return
    }

    setUploadingFiles(selectedFiles.map((file) => file.name))
    const results = await Promise.allSettled(
      selectedFiles.map((file) => uploadMaterial(coursePhaseId, presentation.id, file)),
    )
    setUploadingFiles([])
    if (inputRef.current) inputRef.current.value = ''

    const failed = results.filter((result) => result.status === 'rejected').length
    void queryClient.invalidateQueries({
      queryKey: ['presentation-materials', coursePhaseId, presentation.id],
    })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
    if (failed > 0) {
      toast({
        title: 'Some files could not be uploaded',
        description: `${selectedFiles.length - failed} of ${selectedFiles.length} files uploaded.`,
        variant: 'destructive',
      })
    } else {
      toast({
        title: 'Materials uploaded',
        description: `${selectedFiles.length} file${selectedFiles.length === 1 ? '' : 's'} added.`,
      })
    }
  }

  const handleDownload = async (materialId: string) => {
    setDownloadingId(materialId)
    try {
      await openMaterialDownload(coursePhaseId, presentation.id, materialId)
    } catch (error) {
      toast({
        title: 'Could not download material',
        description: getErrorMessage(error, 'Please try again.'),
        variant: 'destructive',
      })
    } finally {
      setDownloadingId(undefined)
    }
  }

  return (
    <Card>
      <CardHeader className='flex-row items-start justify-between gap-4'>
        <div>
          <CardTitle className='text-base'>Presentation materials</CardTitle>
          <CardDescription>
            Slides and supporting files. Each file may be up to 50 MB.
          </CardDescription>
        </div>
        {canManage ? (
          <>
            <input
              ref={inputRef}
              type='file'
              multiple
              className='hidden'
              onChange={(event) => void handleFiles(event.target.files)}
            />
            <Button
              size='sm'
              variant='outline'
              disabled={uploadingFiles.length > 0}
              onClick={() => inputRef.current?.click()}
            >
              {uploadingFiles.length > 0 ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <Upload className='mr-2 h-4 w-4' />
              )}
              Upload files
            </Button>
          </>
        ) : null}
      </CardHeader>
      <CardContent className='space-y-3'>
        {!canManage && !isStaff ? (
          <Alert>
            <AlertDescription>
              Student uploads and deletions closed when the presentation started.
            </AlertDescription>
          </Alert>
        ) : null}
        {uploadingFiles.length > 0 ? (
          <div className='rounded-md border bg-muted/30 p-3 text-sm text-muted-foreground'>
            Uploading {uploadingFiles.join(', ')}
          </div>
        ) : null}
        {materialsQuery.isLoading ? (
          <div className='flex items-center gap-2 text-sm text-muted-foreground'>
            <Loader2 className='h-4 w-4 animate-spin' />
            Loading materials…
          </div>
        ) : null}
        {materialsQuery.isError ? (
          <Alert variant='destructive'>
            <AlertDescription>Materials could not be loaded.</AlertDescription>
          </Alert>
        ) : null}
        {materialsQuery.data?.length === 0 ? (
          <p className='text-sm text-muted-foreground'>No materials have been uploaded yet.</p>
        ) : null}
        {materialsQuery.data?.map((material) => (
          <div
            key={material.id}
            className='flex flex-col gap-3 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between'
          >
            <div className='flex min-w-0 items-center gap-3'>
              <File className='h-5 w-5 shrink-0 text-muted-foreground' />
              <div className='min-w-0'>
                <p className='truncate text-sm font-medium'>{material.fileName}</p>
                <p className='text-xs text-muted-foreground'>
                  {formatFileSize(material.sizeBytes)}
                  {material.uploadedByName ? ` · ${material.uploadedByName}` : ''}
                </p>
              </div>
            </div>
            <div className='flex gap-2'>
              <Button
                size='sm'
                variant='outline'
                disabled={downloadingId === material.id}
                onClick={() => void handleDownload(material.id)}
              >
                {downloadingId === material.id ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <Download className='h-4 w-4' />
                )}
                <span className='sr-only'>Download {material.fileName}</span>
              </Button>
              {canManage ? (
                <Button
                  size='sm'
                  variant='ghost'
                  className='text-destructive hover:bg-destructive/10 hover:text-destructive'
                  disabled={deleteMutation.isPending}
                  onClick={() => deleteMutation.mutate(material.id)}
                >
                  <Trash2 className='h-4 w-4' />
                  <span className='sr-only'>Delete {material.fileName}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
