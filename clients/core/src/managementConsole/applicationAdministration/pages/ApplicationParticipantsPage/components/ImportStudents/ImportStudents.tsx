import type { ImportApplicationRequest } from '@core/managementConsole/applicationAdministration/interfaces/import/importApplicationRequest'
import type { ImportResult } from '@core/managementConsole/applicationAdministration/interfaces/import/importResult'
import { postApplicationImport } from '@core/network/mutations/postApplicationImport'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { PassStatus } from '@tumaet/prompt-shared-state'
import {
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  useToast,
} from '@tumaet/prompt-ui-components'
import { Loader2, Upload } from 'lucide-react'
import { type ReactNode, useCallback, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { ApplicationParticipation } from '../../../../interfaces/applicationParticipation'
import { ImportMappingStep } from './components/ImportMappingStep'
import { ImportPreviewStep } from './components/ImportPreviewStep'
import { ImportResultSummary } from './components/ImportResultSummary'
import { ImportUploadStep } from './components/ImportUploadStep'
import { buildImportRequest } from './utils/buildImportRequest'
import { buildInitialMapping, type ColumnTarget, validateMapping } from './utils/matchImportColumns'
import { parseImportCsv } from './utils/parseImportCsv'

interface ImportStudentsProps {
  existingApplications: ApplicationParticipation[]
}

const UPLOAD_PAGE = 1
const MAPPING_PAGE = 2
const PREVIEW_PAGE = 3
const RESULT_PAGE = 4

export const ImportStudents = ({ existingApplications }: ImportStudentsProps): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const [open, setOpen] = useState(false)
  const [page, setPage] = useState(UPLOAD_PAGE)
  const [fileName, setFileName] = useState<string | null>(null)
  const [headers, setHeaders] = useState<string[]>([])
  const [rows, setRows] = useState<Record<string, string>[]>([])
  const [mapping, setMapping] = useState<Record<string, ColumnTarget>>({})
  const [passStatus, setPassStatus] = useState<PassStatus>(PassStatus.PASSED)
  const [uploadError, setUploadError] = useState<string | null>(null)
  const [mappingError, setMappingError] = useState<string | null>(null)
  const [result, setResult] = useState<ImportResult | null>(null)

  const resetStates = useCallback(() => {
    setPage(UPLOAD_PAGE)
    setFileName(null)
    setHeaders([])
    setRows([])
    setMapping({})
    setPassStatus(PassStatus.PASSED)
    setUploadError(null)
    setMappingError(null)
    setResult(null)
  }, [])

  const request = useMemo<ImportApplicationRequest | null>(
    () => (headers.length > 0 ? buildImportRequest(headers, rows, mapping, passStatus) : null),
    [headers, rows, mapping, passStatus],
  )

  const existingLogins = useMemo(() => {
    const logins = existingApplications
      .map((application) => (application.student.universityLogin ?? '').toLowerCase())
      .filter((login) => login.length > 0)
    return new Set(logins)
  }, [existingApplications])

  const updateCount = request
    ? request.rows.filter((row) => existingLogins.has(row.student.universityLogin)).length
    : 0
  const newCount = request ? request.rows.length - updateCount : 0

  const { mutate, isPending } = useMutation({
    mutationFn: (importRequest: ImportApplicationRequest) =>
      postApplicationImport(phaseId ?? '', importRequest),
    onSuccess: (importResult) => {
      queryClient.invalidateQueries({ queryKey: ['application_participations', phaseId] })
      queryClient.invalidateQueries({
        queryKey: ['application_participations', 'students', phaseId],
      })
      setResult(importResult)
      setPage(RESULT_PAGE)
      toast({
        title: `Imported ${importResult.created + importResult.updated} students`,
        variant: 'default',
      })
    },
    onError: () => {
      toast({
        title: 'Import failed',
        description: 'Please check the file and try again.',
        variant: 'destructive',
      })
    },
  })

  const handleFile = (file: File) => {
    setUploadError(null)
    parseImportCsv(file)
      .then((parsed) => {
        setFileName(file.name)
        setHeaders(parsed.headers)
        setRows(parsed.rows)
        setMapping(buildInitialMapping(parsed.headers))
      })
      .catch((error: unknown) => {
        setUploadError(error instanceof Error ? error.message : 'Could not parse the CSV file.')
        setFileName(null)
        setHeaders([])
        setRows([])
      })
  }

  const handleMappingChange = (header: string, target: ColumnTarget) => {
    setMapping((previous) => ({ ...previous, [header]: target }))
  }

  const goNext = () => {
    if (page === UPLOAD_PAGE) {
      if (headers.length === 0) {
        setUploadError('Please select a CSV file.')
        return
      }
      setPage(MAPPING_PAGE)
    } else if (page === MAPPING_PAGE) {
      const validationError = validateMapping(mapping)
      if (validationError) {
        setMappingError(validationError.message)
        return
      }
      setMappingError(null)
      setPage(PREVIEW_PAGE)
    } else if (page === PREVIEW_PAGE && request) {
      mutate(request)
    }
  }

  const goBack = () => {
    setPage((previous) => Math.max(UPLOAD_PAGE, previous - 1))
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(newOpen) => {
        if (!newOpen) {
          resetStates()
        }
        setOpen(newOpen)
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <Upload className='h-4 w-4 mr-2' />
          Import Students
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[900px] w-[90vw]'>
        <DialogHeader>
          <DialogTitle>Import Students</DialogTitle>
        </DialogHeader>

        <div className='mt-4 min-w-0'>
          {page === UPLOAD_PAGE && (
            <ImportUploadStep
              fileName={fileName}
              headers={headers}
              rows={rows}
              error={uploadError}
              onFile={handleFile}
            />
          )}
          {page === MAPPING_PAGE && (
            <ImportMappingStep
              headers={headers}
              mapping={mapping}
              error={mappingError}
              onChange={handleMappingChange}
            />
          )}
          {page === PREVIEW_PAGE && request && (
            <ImportPreviewStep
              totalRows={request.rows.length}
              newCount={newCount}
              updateCount={updateCount}
              questionCount={request.newQuestions.length}
              passStatus={passStatus}
              onPassStatusChange={setPassStatus}
            />
          )}
          {page === RESULT_PAGE && result && <ImportResultSummary result={result} />}
        </div>

        <div className='mt-4 flex justify-between'>
          {page !== UPLOAD_PAGE && page !== RESULT_PAGE ? (
            <Button variant='outline' onClick={goBack} disabled={isPending}>
              Previous
            </Button>
          ) : (
            <span />
          )}
          {page === RESULT_PAGE ? (
            <Button
              className='ml-auto'
              onClick={() => {
                resetStates()
                setOpen(false)
              }}
            >
              Done
            </Button>
          ) : (
            <Button className='ml-auto' onClick={goNext} disabled={isPending}>
              {isPending && <Loader2 className='h-4 w-4 mr-2 animate-spin' />}
              {page === PREVIEW_PAGE ? 'Import' : 'Next'}
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
