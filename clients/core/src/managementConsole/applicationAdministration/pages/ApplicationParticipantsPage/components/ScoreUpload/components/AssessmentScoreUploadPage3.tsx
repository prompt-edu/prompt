import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import { ApplicationParticipation } from '../../../../../interfaces/applicationParticipation'
import { translations } from '@tumaet/prompt-shared-state'

interface Page3Props {
  matchedCount: number
  unmatchedApplications: ApplicationParticipation[]
  belowThreshold: number | null
  rowsWithError: string[][]
}

export const AssessmentScoreUploadPage3 = ({
  matchedCount,
  unmatchedApplications,
  belowThreshold,
  rowsWithError,
}: Page3Props) => {
  return (
    <div className='space-y-6'>
      <div>
        <h3 className='text-lg font-medium'>Matching Results</h3>
        <p className='text-sm text-muted-foreground'>
          Successfully matched students: {matchedCount}
        </p>
        {belowThreshold !== null && (
          <p className='text-sm text-muted-foreground'>
            Students below threshold: {belowThreshold}
          </p>
        )}
        <p className='text-sm text-muted-foreground mt-2'>
          Note: All numeric values, including the threshold, are rounded to two decimal places.
        </p>
      </div>

      {unmatchedApplications.length > 1 && (
        <div className='mt-4 h-[300px] sm:max-w-[850px] w-[85vw] overflow-hidden flex flex-col'>
          <h4 className='text-md font-medium mb-2'>Unmatched Applications</h4>
          <div className='overflow-x-auto overflow-y-auto'>
            <Table>
              <TableHeader className='min-w-[150px] bg-muted'>
                <TableRow>
                  <TableHead>First Name</TableHead>
                  <TableHead>Last Name</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Matriculation Number</TableHead>
                  <TableHead>{translations.university['login-name']}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {unmatchedApplications.map((app) => (
                  <TableRow key={app.courseParticipationID}>
                    <TableCell>{app.student.firstName}</TableCell>
                    <TableCell>{app.student.lastName}</TableCell>
                    <TableCell>{app.student.email}</TableCell>
                    <TableCell>{app.student.matriculationNumber || 'N/A'}</TableCell>
                    <TableCell>{app.student.universityLogin || 'N/A'}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}

      {rowsWithError.length > 1 && (
        <div className='mt-4 h-[300px] sm:max-w-[850px] w-[85vw] overflow-hidden flex flex-col'>
          <h4 className='text-md font-medium mb-2'>Rows with Errors</h4>
          <div className='overflow-x-auto overflow-y-auto'>
            <Table>
              <TableHeader>
                <TableRow>
                  {rowsWithError[0].map((header, index) => (
                    <TableHead key={index} className='min-w-[150px] bg-muted'>
                      {header}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {rowsWithError.slice(1).map((row, rowIndex) => (
                  <TableRow key={rowIndex}>
                    {row.map((cell, cellIndex) => (
                      <TableCell key={cellIndex}>{cell}</TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}
    </div>
  )
}
