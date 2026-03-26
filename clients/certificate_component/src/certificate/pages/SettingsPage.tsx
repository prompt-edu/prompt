import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Loader2,
  Upload,
  CheckCircle2,
  Eye,
  AlertCircle,
  X,
  HelpCircle,
  AlertTriangle,
  CalendarCheck,
} from 'lucide-react'

import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Textarea,
  DatePicker,
  Popover,
  PopoverContent,
  PopoverTrigger,
  ManagementPageHeader,
  ErrorPage,
  Alert,
  AlertDescription,
  AlertTitle,
  useToast,
} from '@tumaet/prompt-ui-components'

import { getConfig, updateConfig, updateReleaseDate } from '../network/queries/getConfig'
import { previewCertificate, type PreviewError } from '../network/queries/previewCertificate'

/**
 * Format a date string to European format: DD.MM.YYYY HH:mm
 */
const formatEuropeanDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleString('de-DE', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export const SettingsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [templateContent, setTemplateContent] = useState('')
  const [hasChanges, setHasChanges] = useState(false)
  const [isPreviewing, setIsPreviewing] = useState(false)
  const [compilerError, setCompilerError] = useState<string | null>(null)
  const [selectedReleaseDate, setSelectedReleaseDate] = useState<Date | undefined>(undefined)
  const [releaseDateDirty, setReleaseDateDirty] = useState(false)

  const {
    data: config,
    isPending,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['config', phaseId],
    queryFn: () => getConfig(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const updateMutation = useMutation({
    mutationFn: (content: string) => updateConfig(phaseId ?? '', content),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', phaseId] })
      toast({
        title: 'Success',
        description: 'Certificate template updated successfully',
      })
      setHasChanges(false)
    },
    onError: () => {
      toast({
        title: 'Error',
        description: 'Failed to update certificate template',
        variant: 'destructive',
      })
    },
  })

  const releaseDateMutation = useMutation({
    mutationFn: (releaseDate: string | null) => updateReleaseDate(phaseId ?? '', releaseDate),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', phaseId] })
      setReleaseDateDirty(false)
      toast({
        title: 'Success',
        description: 'Release date updated successfully',
      })
    },
    onError: () => {
      toast({
        title: 'Error',
        description: 'Failed to update release date',
        variant: 'destructive',
      })
    },
  })

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (!file.name.endsWith('.typ')) {
      toast({
        title: 'Invalid file type',
        description: 'Please upload a Typst (.typ) file',
        variant: 'destructive',
      })
      return
    }

    const reader = new FileReader()
    reader.onload = (event) => {
      const content = event.target?.result as string
      setTemplateContent(content)
      setHasChanges(true)
    }
    reader.readAsText(file)
  }

  const handleSave = () => {
    if (templateContent) {
      updateMutation.mutate(templateContent)
    }
  }

  const handlePreview = async () => {
    if (!phaseId || !templateContent) return

    setIsPreviewing(true)
    setCompilerError(null)
    try {
      // Save first if there are unsaved changes
      if (hasChanges) {
        try {
          await updateConfig(phaseId, templateContent)
          queryClient.invalidateQueries({ queryKey: ['config', phaseId] })
          setHasChanges(false)
          toast({
            title: 'Template saved',
            description: 'Template saved before generating preview',
          })
        } catch {
          toast({
            title: 'Save failed',
            description: 'Failed to save template before preview.',
            variant: 'destructive',
          })
          return
        }
      }

      const blob = await previewCertificate(phaseId)
      const url = window.URL.createObjectURL(blob)
      window.open(url, '_blank')
      setTimeout(() => window.URL.revokeObjectURL(url), 10000)
    } catch (error: unknown) {
      console.error('Failed to generate preview:', error)
      const previewError = error as PreviewError
      if (previewError?.compilerOutput) {
        setCompilerError(previewError.compilerOutput)
      } else {
        toast({
          title: 'Preview failed',
          description: 'Failed to generate certificate preview. Make sure the template is valid.',
          variant: 'destructive',
        })
      }
    } finally {
      setIsPreviewing(false)
    }
  }

  const handleSaveReleaseDate = () => {
    if (selectedReleaseDate) {
      releaseDateMutation.mutate(selectedReleaseDate.toISOString())
    } else {
      releaseDateMutation.mutate(null)
    }
  }

  const handleReleaseNow = () => {
    const now = new Date()
    setSelectedReleaseDate(now)
    releaseDateMutation.mutate(now.toISOString())
  }

  const handleClearReleaseDate = () => {
    setSelectedReleaseDate(undefined)
    releaseDateMutation.mutate(null)
  }

  // Initialize template content from config
  const initialized = useRef(false)
  useEffect(() => {
    if (config?.templateContent && !initialized.current) {
      setTemplateContent(config.templateContent)
      initialized.current = true
    }
  }, [config?.templateContent])

  if (isError) {
    return <ErrorPage message='Error loading configuration' onRetry={refetch} />
  }

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>Certificate Settings</ManagementPageHeader>
      <p className='text-muted-foreground'>
        Configure the certificate template for this course phase.
      </p>

      {/* Warning: template changed after students downloaded */}
      {config?.hasDownloads && hasChanges && (
        <Alert variant='destructive'>
          <AlertTriangle className='h-4 w-4' />
          <AlertTitle>Students have already downloaded certificates</AlertTitle>
          <AlertDescription>
            Changing the template will affect all future downloads. Students who already downloaded
            their certificate will not automatically receive the updated version.
          </AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <div className='flex items-center justify-between'>
            <div>
              <CardTitle>Certificate Template</CardTitle>
              <CardDescription>
                Upload a Typst (.typ) file or paste the template content directly.
              </CardDescription>
            </div>
            <div className='flex items-center gap-3'>
              {config?.hasTemplate && (
                <span className='flex items-center gap-2 text-sm text-green-600'>
                  <CheckCircle2 className='h-4 w-4' />
                  Configured
                  {config.updatedAt && (
                    <span className='text-muted-foreground'>
                      · {formatEuropeanDate(config.updatedAt)}
                      {config.updatedBy && ` by ${config.updatedBy}`}
                    </span>
                  )}
                </span>
              )}
              <Popover>
                <PopoverTrigger asChild>
                  <Button
                    variant='ghost'
                    size='sm'
                    className='h-8 w-8 p-0'
                    aria-label='Template requirements help'
                  >
                    <HelpCircle className='h-4 w-4 text-muted-foreground' />
                  </Button>
                </PopoverTrigger>
                <PopoverContent className='w-80' align='end'>
                  <div className='space-y-3 p-1'>
                    <h4 className='font-medium text-sm'>Template Requirements</h4>
                    <ul className='list-disc pl-4 space-y-2 text-sm text-muted-foreground'>
                      <li>File must be a valid Typst (.typ) file</li>
                      <li>
                        Use{' '}
                        <code className='bg-muted px-1 py-0.5 rounded text-xs'>
                          json(&quot;data.json&quot;)
                        </code>{' '}
                        to access certificate data
                      </li>
                      <li>
                        Available fields:{' '}
                        <code className='bg-muted px-1 py-0.5 rounded text-xs'>studentName</code>,{' '}
                        <code className='bg-muted px-1 py-0.5 rounded text-xs'>courseName</code>,{' '}
                        <code className='bg-muted px-1 py-0.5 rounded text-xs'>teamName</code>,{' '}
                        <code className='bg-muted px-1 py-0.5 rounded text-xs'>date</code>
                      </li>
                      <li>Template should be in A4 format</li>
                      <li>Ensure all used fonts are included or are system fonts</li>
                    </ul>
                  </div>
                </PopoverContent>
              </Popover>
            </div>
          </div>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='flex items-center gap-4'>
            <Button variant='outline' asChild>
              <label className='cursor-pointer'>
                <Upload className='mr-2 h-4 w-4' />
                Upload .typ file
                <input type='file' accept='.typ' onChange={handleFileUpload} className='hidden' />
              </label>
            </Button>
            <span className='text-sm text-muted-foreground'>or paste content below</span>
          </div>

          <Textarea
            placeholder='Paste your Typst template content here...'
            value={templateContent}
            onChange={(e) => {
              setTemplateContent(e.target.value)
              setHasChanges(true)
            }}
            rows={15}
            className='font-mono text-sm'
          />

          <div className='flex items-center gap-3'>
            <Button
              onClick={handleSave}
              disabled={!hasChanges || !templateContent || updateMutation.isPending}
            >
              {updateMutation.isPending ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Saving...
                </>
              ) : (
                'Save Template'
              )}
            </Button>

            <Button
              variant='outline'
              onClick={handlePreview}
              disabled={!templateContent || isPreviewing || updateMutation.isPending}
            >
              {isPreviewing ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Generating...
                </>
              ) : (
                <>
                  <Eye className='mr-2 h-4 w-4' />
                  Test Certificate
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {compilerError && (
        <Alert variant='destructive'>
          <AlertCircle className='h-4 w-4' />
          <AlertTitle className='flex items-center justify-between'>
            Template Compilation Error
            <Button
              variant='ghost'
              size='sm'
              className='h-6 w-6 p-0'
              onClick={() => setCompilerError(null)}
            >
              <X className='h-4 w-4' />
            </Button>
          </AlertTitle>
          <AlertDescription>
            <pre className='mt-2 whitespace-pre-wrap break-words rounded bg-destructive/10 p-3 font-mono text-xs'>
              {compilerError}
            </pre>
          </AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <div className='flex items-center justify-between'>
            <div>
              <CardTitle>Release Date</CardTitle>
              <CardDescription>
                Set a date after which students can download their certificates. Leave empty to
                allow downloads immediately.
              </CardDescription>
            </div>
            {config?.releaseDate && (
              <span className='flex items-center gap-2 text-sm text-muted-foreground'>
                {formatEuropeanDate(config.releaseDate)}
                {config.updatedBy && ` by ${config.updatedBy}`}
              </span>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <div className='flex items-center gap-4'>
            <DatePicker
              date={
                selectedReleaseDate ??
                (config?.releaseDate ? new Date(config.releaseDate) : undefined)
              }
              onSelect={(date) => {
                setSelectedReleaseDate(date)
                setReleaseDateDirty(true)
              }}
            />
            <Button
              onClick={handleSaveReleaseDate}
              disabled={!releaseDateDirty || releaseDateMutation.isPending}
            >
              {releaseDateMutation.isPending ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Saving...
                </>
              ) : (
                'Save'
              )}
            </Button>
            <Button
              variant='outline'
              onClick={handleReleaseNow}
              disabled={releaseDateMutation.isPending}
            >
              <CalendarCheck className='mr-2 h-4 w-4' />
              Release Now
            </Button>
            {(selectedReleaseDate || config?.releaseDate) && (
              <Button
                variant='ghost'
                size='sm'
                onClick={handleClearReleaseDate}
                disabled={releaseDateMutation.isPending}
              >
                <X className='mr-1 h-4 w-4' />
                Clear
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
