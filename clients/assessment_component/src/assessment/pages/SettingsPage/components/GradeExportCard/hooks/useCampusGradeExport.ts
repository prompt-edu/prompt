import { useToast } from '@tumaet/prompt-ui-components'
import { useCallback, useMemo, useState } from 'react'

import { useGetAllAssessmentCompletions } from '../../../../hooks/useGetAllAssessmentCompletions'
import { useGetCoursePhaseParticipations } from '../../../../hooks/useGetCoursePhaseParticipations'
import { triggerBlobDataDownload } from '../../../../utils/triggerBlobDataDownload'
import type { GradeFillReport, ParsedCampusCsv } from '../interfaces/campusGradeExport'
import { buildStudentGradeEntries } from '../utils/buildStudentGradeEntries'
import { decodeCsvFile, encodeCsvText, getCsvMimeType } from '../utils/campusCsvEncoding'
import {
  buildFilledCsvFileName,
  MAX_CSV_BYTES,
  parseCampusCsv,
  serializeCampusCsv,
} from '../utils/campusCsvFile'
import { fillCampusGrades } from '../utils/fillCampusGrades'

export type CampusGradeExportStatus = 'idle' | 'parsing' | 'ready' | 'error'

interface UseCampusGradeExportResult {
  status: CampusGradeExportStatus
  fileName: string | null
  report: GradeFillReport | null
  errorMessage: string | null
  hasGradesToExport: boolean
  isDataPending: boolean
  isDataError: boolean
  handleFileSelected: (file: File) => Promise<void>
  handleDownload: () => void
  reset: () => void
}

/**
 * Drives the CampusOnline round trip: decode and parse the uploaded file, fill
 * the grades PROMPT knows about, and hand the result back as a download.
 *
 * Only the parsed file is held in state; the filled result is derived from it,
 * so a grade edited elsewhere is reflected without re-uploading. The uploaded
 * `File` itself is deliberately not retained, which makes producing the download
 * a pure function of state with no second read from disk.
 */
export const useCampusGradeExport = (): UseCampusGradeExportResult => {
  const { toast } = useToast()

  const {
    data: participations,
    isPending: isParticipationsPending,
    isError: isParticipationsError,
  } = useGetCoursePhaseParticipations()
  const {
    data: assessmentCompletions,
    isPending: isCompletionsPending,
    isError: isCompletionsError,
  } = useGetAllAssessmentCompletions()

  const [status, setStatus] = useState<CampusGradeExportStatus>('idle')
  const [fileName, setFileName] = useState<string | null>(null)
  const [parsedCsv, setParsedCsv] = useState<ParsedCampusCsv | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const gradeEntries = useMemo(
    () => buildStudentGradeEntries(participations, assessmentCompletions ?? []),
    [participations, assessmentCompletions],
  )

  const filledCsv = useMemo(
    () => (parsedCsv ? fillCampusGrades(parsedCsv, gradeEntries) : null),
    [parsedCsv, gradeEntries],
  )

  const reset = useCallback(() => {
    setStatus('idle')
    setFileName(null)
    setParsedCsv(null)
    setErrorMessage(null)
  }, [])

  const handleFileSelected = useCallback(
    async (file: File) => {
      setStatus('parsing')
      setFileName(file.name)
      setParsedCsv(null)
      setErrorMessage(null)

      try {
        if (file.size > MAX_CSV_BYTES) {
          throw new Error(
            `The file is larger than ${MAX_CSV_BYTES / (1024 * 1024)} MB and is probably not a CampusOnline grade export.`,
          )
        }

        const decoded = decodeCsvFile(await file.arrayBuffer())

        setParsedCsv(parseCampusCsv(decoded, file.name))
        setStatus('ready')
      } catch (error) {
        const message =
          error instanceof Error ? error.message : 'The CSV file could not be processed.'
        setErrorMessage(message)
        setStatus('error')
        toast({
          title: 'Could not read the CSV file',
          description: message,
          variant: 'destructive',
        })
      }
    },
    [toast],
  )

  const handleDownload = useCallback(() => {
    if (!filledCsv) return

    const csvText = serializeCampusCsv(filledCsv.csv)
    triggerBlobDataDownload(
      encodeCsvText(csvText, filledCsv.csv.encoding),
      buildFilledCsvFileName(filledCsv.csv.fileName),
      getCsvMimeType(filledCsv.csv.encoding),
    )
  }, [filledCsv])

  return {
    status,
    fileName,
    report: filledCsv?.report ?? null,
    errorMessage,
    hasGradesToExport: gradeEntries.some((entry) => entry.grade !== null),
    isDataPending: isParticipationsPending || isCompletionsPending,
    isDataError: isParticipationsError || isCompletionsError,
    handleFileSelected,
    handleDownload,
    reset,
  }
}
