import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  ScrollArea,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'

import type {
  GradeFillReport,
  SkippedCsvRow,
  SkippedRowReason,
} from '../interfaces/campusGradeExport'

interface CampusGradeFillSummaryProps {
  report: GradeFillReport
}

const SKIP_REASON_LABEL: Record<SkippedRowReason, string> = {
  noGradeInPrompt: 'Not graded in PROMPT yet',
  unmatched: 'No matching student in this phase',
  ambiguous: 'Matches more than one student',
}

const groupSkippedRows = (
  skipped: SkippedCsvRow[],
): { reason: SkippedRowReason; rows: SkippedCsvRow[] }[] =>
  (['ambiguous', 'noGradeInPrompt', 'unmatched'] as SkippedRowReason[])
    .map((reason) => ({ reason, rows: skipped.filter((row) => row.reason === reason) }))
    .filter((group) => group.rows.length > 0)

export const CampusGradeFillSummary = ({ report }: CampusGradeFillSummaryProps) => {
  const skippedGroups = groupSkippedRows(report.skipped)
  const hasBlockingWarning =
    report.gradedStudentsMissingFromCsv.length > 0 ||
    report.skipped.some((row) => row.reason === 'ambiguous')

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center gap-2'>
        <Badge variant='secondary'>
          {report.filled.length} of {report.totalDataRows}{' '}
          {report.totalDataRows === 1 ? 'row' : 'rows'} filled
        </Badge>
        {report.skipped.length > 0 && (
          <Badge variant='outline'>{report.skipped.length} skipped</Badge>
        )}
        {report.gradedStudentsMissingFromCsv.length > 0 && (
          <Badge variant='destructive'>
            {report.gradedStudentsMissingFromCsv.length} graded students missing from the CSV
          </Badge>
        )}
        {report.nameFallbackCount > 0 && (
          <Badge variant='outline'>{report.nameFallbackCount} matched by name only</Badge>
        )}
        {report.overwrittenCount > 0 && (
          <Badge variant='outline'>{report.overwrittenCount} existing grades replaced</Badge>
        )}
      </div>

      {report.filled.length === 0 && (
        <Alert variant='destructive'>
          <AlertTitle>No grades were filled in</AlertTitle>
          <AlertDescription>
            None of the rows in this file could be matched to a graded student of this phase. Check
            that the file belongs to this course and that the assessments are marked as complete.
          </AlertDescription>
        </Alert>
      )}

      {report.gradedStudentsMissingFromCsv.length > 0 && (
        <Alert variant='destructive'>
          <AlertTitle>Some grades will not reach CampusOnline</AlertTitle>
          <AlertDescription>
            These students are graded in PROMPT but have no row in the uploaded file, so their grade
            is not exported. This usually means the file was downloaded for a different exam date or
            course group.
          </AlertDescription>
        </Alert>
      )}

      {report.nameFallbackCount > 0 && (
        <Alert>
          <AlertTitle>Some rows were matched by name</AlertTitle>
          <AlertDescription>
            {report.nameFallbackCount}{' '}
            {report.nameFallbackCount === 1 ? 'row had no' : 'rows had no'} usable registration
            number and {report.nameFallbackCount === 1 ? 'was matched' : 'were matched'} by family
            and first name instead. Please spot-check those grades.
          </AlertDescription>
        </Alert>
      )}

      {report.passStatusMismatchCount > 0 && (
        <Alert>
          <AlertTitle>Grade does not match the pass status</AlertTitle>
          <AlertDescription>
            {report.passStatusMismatchCount}{' '}
            {report.passStatusMismatchCount === 1 ? 'student is' : 'students are'} marked as failed
            in PROMPT but {report.passStatusMismatchCount === 1 ? 'has' : 'have'} a passing grade.
            The grade from the assessment is exported as it is.
          </AlertDescription>
        </Alert>
      )}

      {report.filled.length > 0 && (
        <div className='space-y-2'>
          <h3 className='text-sm font-semibold text-foreground'>Filled grades</h3>
          <ScrollArea className='h-[300px] rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Row</TableHead>
                  <TableHead>Registration Number</TableHead>
                  <TableHead>Student</TableHead>
                  <TableHead>Grade</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Matched By</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {report.filled.map((row) => (
                  <TableRow key={row.csvRowNumber}>
                    <TableCell>{row.csvRowNumber}</TableCell>
                    <TableCell>{row.registrationNumber || '-'}</TableCell>
                    <TableCell>{row.studentName || '-'}</TableCell>
                    <TableCell>
                      {row.grade}
                      {row.overwrittenGrade && (
                        <span className='ml-2 text-xs text-muted-foreground'>
                          was {row.overwrittenGrade}
                        </span>
                      )}
                    </TableCell>
                    <TableCell>{row.dateOfAssessment || '-'}</TableCell>
                    <TableCell>
                      {row.matchedBy === 'name' ? (
                        <Badge variant='outline'>Name</Badge>
                      ) : (
                        'Registration number'
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </div>
      )}

      {skippedGroups.map((group) => (
        <div className='space-y-2' key={group.reason}>
          <h3 className='text-sm font-semibold text-foreground'>
            {SKIP_REASON_LABEL[group.reason]} ({group.rows.length})
          </h3>
          <ScrollArea className='h-[220px] rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Row</TableHead>
                  <TableHead>Registration Number</TableHead>
                  <TableHead>Student</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {group.rows.map((row) => (
                  <TableRow key={row.csvRowNumber}>
                    <TableCell>{row.csvRowNumber}</TableCell>
                    <TableCell>{row.registrationNumber || '-'}</TableCell>
                    <TableCell>{row.studentName || '-'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </div>
      ))}

      {report.gradedStudentsMissingFromCsv.length > 0 && (
        <div className='space-y-2'>
          <h3 className='text-sm font-semibold text-foreground'>
            Graded students missing from the CSV ({report.gradedStudentsMissingFromCsv.length})
          </h3>
          <ScrollArea className='h-[220px] rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Student</TableHead>
                  <TableHead>Matriculation Number</TableHead>
                  <TableHead>Grade</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {report.gradedStudentsMissingFromCsv.map((row) => (
                  <TableRow key={row.courseParticipationID}>
                    <TableCell>{row.studentName}</TableCell>
                    <TableCell>{row.matriculationNumber || '-'}</TableCell>
                    <TableCell>{row.grade.toFixed(1)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </div>
      )}

      {!hasBlockingWarning && report.filled.length > 0 && (
        <p className='text-sm leading-6 text-muted-foreground'>
          Only the grade and assessment date columns were changed. Every other column is exported
          exactly as CampusOnline delivered it.
        </p>
      )}
    </div>
  )
}
