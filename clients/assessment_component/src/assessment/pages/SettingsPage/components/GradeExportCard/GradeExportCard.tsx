import { Alert, AlertDescription, AlertTitle, Button, Card } from '@tumaet/prompt-ui-components'
import { Download } from 'lucide-react'

import { CampusCsvUploadField } from './components/CampusCsvUploadField'
import { CampusGradeFillSummary } from './components/CampusGradeFillSummary'
import { useCampusGradeExport } from './hooks/useCampusGradeExport'

/**
 * Fills the grades of this phase into a CampusOnline exam CSV so the lecturer
 * can upload the result back into CampusOnline instead of typing every grade by
 * hand.
 */
export const GradeExportCard = () => {
  const {
    status,
    fileName,
    report,
    errorMessage,
    hasGradesToExport,
    isDataPending,
    isDataError,
    handleFileSelected,
    handleDownload,
    reset,
  } = useCampusGradeExport()

  return (
    <Card className='border-border shadow-xs'>
      <div className='space-y-6 p-6'>
        <div className='space-y-2'>
          <h2 className='text-xl font-semibold text-foreground'>CampusOnline Grade Export</h2>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
            Transfer the grades of this phase into the exam list from CampusOnline. PROMPT only
            fills in the grade and assessment date columns of the file you upload, so the file stays
            importable.
          </p>
          <ol className='max-w-3xl list-decimal space-y-1 pl-5 text-sm leading-6 text-muted-foreground'>
            <li>Download the exam list as CSV from CampusOnline.</li>
            <li>Upload that file here.</li>
            <li>Check which students were matched and graded.</li>
            <li>Download the filled file.</li>
            <li>Upload the filled file back into CampusOnline.</li>
          </ol>
        </div>

        {isDataError && (
          <Alert variant='destructive'>
            <AlertTitle>Failed to load grades</AlertTitle>
            <AlertDescription>
              The participants or assessment grades of this phase could not be loaded. Please reload
              the page before exporting.
            </AlertDescription>
          </Alert>
        )}

        {!isDataPending && !isDataError && !hasGradesToExport && (
          <Alert>
            <AlertTitle>No grades to export yet</AlertTitle>
            <AlertDescription>
              No assessment in this phase is marked as complete, so there is nothing to fill in.
            </AlertDescription>
          </Alert>
        )}

        <CampusCsvUploadField
          onFileSelected={handleFileSelected}
          isParsing={status === 'parsing' || isDataPending}
          fileName={fileName}
          onClear={reset}
        />

        {status === 'error' && errorMessage && (
          <Alert variant='destructive'>
            <AlertTitle>Could not read the CSV file</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        {status === 'ready' && report && <CampusGradeFillSummary report={report} />}

        <div className='flex flex-wrap items-center gap-3'>
          <Button
            type='button'
            onClick={handleDownload}
            disabled={status !== 'ready' || !report || report.filled.length === 0}
          >
            <Download className='mr-2 h-4 w-4' />
            Download Filled CSV
          </Button>
        </div>
      </div>
    </Card>
  )
}
