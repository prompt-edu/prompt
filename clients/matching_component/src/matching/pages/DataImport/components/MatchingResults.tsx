import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ScrollArea,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@tumaet/prompt-ui-components'
import { UploadedStudent } from '../../../interfaces/UploadedStudent'

interface MatchingResultsProps {
  matchedStudents: CoursePhaseParticipationWithStudent[]
  unmatchedStudents: UploadedStudent[]
}

function MatchingResults({ matchedStudents, unmatchedStudents }: MatchingResultsProps) {
  return (
    <div className='space-y-8 mx-auto'>
      {[
        {
          title: 'Successfully Matched Students',
          description: 'These students have been successfully matched by matriculation number.',
          data: matchedStudents,
        },
        {
          title: 'Unmatched Students',
          description: `These students have been assigned to the course, but cannot be find. 
          Please check if a student with this matriculation number (or name) is in this course and has passed the previous phases.`,
          data: unmatchedStudents,
        },
      ].map((section, index) => (
        <Card key={index}>
          <CardHeader>
            <CardTitle>{section.title}</CardTitle>
            <CardDescription className='text-gray-600 mb-4'>{section.description}</CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className='h-[300px] rounded-md border'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>First Name</TableHead>
                    <TableHead>Last Name</TableHead>
                    <TableHead>Matriculation Number</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {section.data.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={3} className='text-center'>
                        No entries
                      </TableCell>
                    </TableRow>
                  )}
                  {section.data.map(
                    (item: UploadedStudent | CoursePhaseParticipationWithStudent, idx: number) => (
                      <TableRow key={idx}>
                        <TableCell>
                          {'student' in item ? item.student?.firstName : item.firstName}
                        </TableCell>
                        <TableCell>
                          {'student' in item ? item.student?.lastName : item.lastName}
                        </TableCell>
                        <TableCell>
                          {'student' in item
                            ? item.student?.matriculationNumber
                            : item.matriculationNumber || 'N/A'}
                        </TableCell>
                      </TableRow>
                    ),
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

export default MatchingResults
