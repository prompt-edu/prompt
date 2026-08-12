import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  useToast,
} from '@tumaet/prompt-ui-components'
import { CircleDashed, Download, File, Loader2, Trash2, Upload } from 'lucide-react'
import { useRef, useState } from 'react'
import type { MaterialType, PresentationMaterial, PresentationSummary } from '../interfaces'
import {
  getMaterialTypeAccept,
  getMaterialTypeDefinition,
  type MaterialTypeDefinition,
  sortMaterialTypes,
} from '../materialTypes'
import { openMaterialDownload, presentationApi, uploadMaterial } from '../network'
import { formatFileSize, getErrorMessage } from '../utils'

// Only covers the moment before the config query resolves. The server enforces its own
// limit either way, so this never has to be exactly right.
const FALLBACK_MAX_FILE_SIZE_BYTES = 50 * 1024 * 1024

interface MaterialSlotProps {
  definition: MaterialTypeDefinition
  materials: PresentationMaterial[]
  required: boolean
  isUploading: boolean
  isDeleting: boolean
  downloadingId?: string
  onUpload?: (files: FileList | null) => void
  onDownload: (materialId: string) => void
  onDelete?: (materialId: string) => void
}

const MaterialSlot = ({
  definition,
  materials,
  required,
  isUploading,
  isDeleting,
  downloadingId,
  onUpload,
  onDownload,
  onDelete,
}: MaterialSlotProps) => {
  const inputRef = useRef<HTMLInputElement>(null)

  return (
    <div className='space-y-3 rounded-md border p-3'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div className='min-w-0'>
          <div className='flex flex-wrap items-center gap-2'>
            <p className='font-medium'>{definition.label}</p>
            {required ? (
              <Badge variant='secondary'>Required</Badge>
            ) : (
              <Badge variant='outline'>No longer requested</Badge>
            )}
            {required && materials.length === 0 ? (
              <span className='flex items-center gap-1 text-xs text-muted-foreground'>
                <CircleDashed className='h-3 w-3' />
                Missing
              </span>
            ) : null}
          </div>
          <p className='text-xs text-muted-foreground'>
            {definition.formats}
            {definition.note ? ` · ${definition.note}` : ''}
          </p>
        </div>
        {onUpload ? (
          <>
            <input
              ref={inputRef}
              type='file'
              multiple
              accept={getMaterialTypeAccept(definition)}
              className='hidden'
              onChange={(event) => onUpload(event.target.files)}
            />
            <Button
              size='sm'
              variant='outline'
              disabled={isUploading}
              onClick={() => inputRef.current?.click()}
            >
              {isUploading ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <Upload className='mr-2 h-4 w-4' />
              )}
              Upload
            </Button>
          </>
        ) : null}
      </div>
      {materials.map((material) => (
        <div
          key={material.id}
          className='flex flex-col gap-3 rounded-md bg-muted/30 p-2 sm:flex-row sm:items-center sm:justify-between'
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
              onClick={() => onDownload(material.id)}
            >
              {downloadingId === material.id ? (
                <Loader2 className='h-4 w-4 animate-spin' />
              ) : (
                <Download className='h-4 w-4' />
              )}
              <span className='sr-only'>Download {material.fileName}</span>
            </Button>
            {onDelete ? (
              <Button
                size='sm'
                variant='ghost'
                className='text-destructive hover:bg-destructive/10 hover:text-destructive'
                disabled={isDeleting}
                onClick={() => onDelete(material.id)}
              >
                <Trash2 className='h-4 w-4' />
                <span className='sr-only'>Delete {material.fileName}</span>
              </Button>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  )
}

interface MaterialsPanelProps {
  coursePhaseId: string
  presentation: PresentationSummary
  isStaff: boolean
  // Sample files for the staff preview of the student page. They replace the API data and
  // turn every action off, so the preview never touches a real presentation.
  previewMaterials?: PresentationMaterial[]
}

export const MaterialsPanel = ({
  coursePhaseId,
  presentation,
  isStaff,
  previewMaterials,
}: MaterialsPanelProps) => {
  const isPreview = previewMaterials !== undefined
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [uploadingType, setUploadingType] = useState<MaterialType>()
  const [downloadingId, setDownloadingId] = useState<string>()
  const canManage =
    !isPreview && (isStaff || new Date(presentation.startTime).getTime() > Date.now())

  const materialsQuery = useQuery({
    queryKey: ['presentation-materials', coursePhaseId, presentation.id],
    queryFn: () => presentationApi.getMaterials(coursePhaseId, presentation.id),
    enabled: !isPreview && Boolean(coursePhaseId && presentation.id),
  })

  const configQuery = useQuery({
    queryKey: ['presentation-config', coursePhaseId],
    queryFn: () => presentationApi.getConfig(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })
  const maxFileSizeBytes = configQuery.data?.maxUploadBytes || FALLBACK_MAX_FILE_SIZE_BYTES
  const requiredTypes = sortMaterialTypes(configQuery.data?.requiredMaterialTypes ?? [])
  const materials = previewMaterials ?? materialsQuery.data ?? []
  // Files whose type the lecturer has since removed from the requirements. They stay
  // visible and downloadable, because deleting a presenter's work is never implicit.
  const staleMaterials = materials.filter(
    (material) => !requiredTypes.includes(material.materialType),
  )

  const invalidateMaterials = () => {
    void queryClient.invalidateQueries({
      queryKey: ['presentation-materials', coursePhaseId, presentation.id],
    })
    void queryClient.invalidateQueries({ queryKey: ['presentations', coursePhaseId] })
  }

  const deleteMutation = useMutation({
    mutationFn: (materialId: string) =>
      presentationApi.deleteMaterial(coursePhaseId, presentation.id, materialId),
    onSuccess: () => {
      invalidateMaterials()
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

  const handleFiles = async (materialType: MaterialType, files: FileList | null) => {
    const selectedFiles = Array.from(files ?? [])
    if (selectedFiles.length === 0) return
    const oversized = selectedFiles.find((file) => file.size > maxFileSizeBytes)
    if (oversized) {
      toast({
        title: 'File is too large',
        description: `${oversized.name} exceeds the ${formatFileSize(maxFileSizeBytes)} limit.`,
        variant: 'destructive',
      })
      return
    }

    setUploadingType(materialType)
    const results = await Promise.allSettled(
      selectedFiles.map((file) =>
        uploadMaterial(coursePhaseId, presentation.id, materialType, file),
      ),
    )
    setUploadingType(undefined)

    const rejected = results.filter((result) => result.status === 'rejected')
    invalidateMaterials()
    if (rejected.length > 0) {
      toast({
        title: 'Some files could not be uploaded',
        description: getErrorMessage(
          rejected[0].reason,
          `${selectedFiles.length - rejected.length} of ${selectedFiles.length} files uploaded.`,
        ),
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
      <CardHeader>
        <CardTitle className='text-base'>Presentation materials</CardTitle>
        <CardDescription>
          {requiredTypes.length > 0
            ? `The teaching team asks for the uploads below. Each file may be up to ${formatFileSize(maxFileSizeBytes)}.`
            : 'The teaching team does not ask for any uploads in this phase.'}
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-3'>
        {!canManage && !isStaff && !isPreview ? (
          <Alert>
            <AlertDescription>
              Student uploads and deletions closed when the presentation started.
            </AlertDescription>
          </Alert>
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
        {requiredTypes.map((materialType) => (
          <MaterialSlot
            key={materialType}
            definition={getMaterialTypeDefinition(materialType)}
            materials={materials.filter((material) => material.materialType === materialType)}
            required
            isUploading={uploadingType === materialType}
            isDeleting={deleteMutation.isPending}
            downloadingId={downloadingId}
            onUpload={canManage ? (files) => void handleFiles(materialType, files) : undefined}
            onDownload={(materialId) => void handleDownload(materialId)}
            onDelete={canManage ? (materialId) => deleteMutation.mutate(materialId) : undefined}
          />
        ))}
        {staleMaterials.length > 0 ? (
          <MaterialSlot
            definition={{
              type: 'slides',
              label: 'Previously uploaded files',
              description: '',
              extensions: [],
              formats: 'These uploads are no longer requested by the teaching team',
            }}
            materials={staleMaterials}
            required={false}
            isUploading={false}
            isDeleting={deleteMutation.isPending}
            downloadingId={downloadingId}
            onDownload={(materialId) => void handleDownload(materialId)}
            onDelete={canManage ? (materialId) => deleteMutation.mutate(materialId) : undefined}
          />
        ) : null}
      </CardContent>
    </Card>
  )
}
