import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Badge,
  Button,
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
import { AlertTriangle, Download } from 'lucide-react'
import { useDataDownload } from '../../../hooks/useDataDownload'
import type { UploadedStudent } from '../../../interfaces/UploadedStudent'

interface MatchingResultsProps {
  matchedByMatriculation: UploadedStudent[]
  matchedByName: UploadedStudent[]
  unmatchedApplications: CoursePhaseParticipationWithStudent[]
  unmatchedStudents: UploadedStudent[]
  useScoreAsRank: boolean
}

function MatchingResults({
  matchedByMatriculation,
  matchedByName,
  unmatchedApplications,
  unmatchedStudents,
  useScoreAsRank,
}: MatchingResultsProps) {
  const { generateAndDownloadFile } = useDataDownload()

  const matchedRankMissing =
    matchedByMatriculation.some((student) => !student.rank) ||
    matchedByName.some((student) => !student.rank)
  return (
    <div className='space-y-8 mx-auto'>
      {matchedRankMissing && (
        <Alert variant='destructive'>
          <AlertTriangle className='h-4 w-4' />
          <AlertTitle>Matched Application is missing application/interview score</AlertTitle>
          <AlertDescription>
            At least one application that has been matched has no application/interview store. This
            will result in the student being unranked. Please make sure to assign every accepted
            application an application score.
          </AlertDescription>
        </Alert>
      )}
      {[
        {
          title: 'Matched by Matriculation Number',
          description: 'These students have been successfully matched by matriculation number.',
          data: matchedByMatriculation,
          showRank: true,
        },
        {
          title: 'Matched by Name',
          description:
            'These students could not be matched by matriculation number, but by first and last name.',
          data: matchedByName,
          showRank: true,
        },
        {
          title: 'Unmatched Accepted Applications',
          description:
            'These students have an accepted application, but could not be matched with an entry in the uploaded file.',
          data: unmatchedApplications,
          showRank: false,
        },
        {
          title: 'Entries Without Applications',
          description: `These entries could not be matched to any application. 
          They might be students who have not applied, have been rejected, or have a typo in their name or matriculation number.`,
          data: unmatchedStudents,
          showRank: false,
        },
      ].map((section, index) => (
        <Card key={index}>
          <CardHeader>
            <CardTitle>{section.title}</CardTitle>
            <CardDescription className='text-gray-600 mb-4'>{section.description}</CardDescription>
            <div className='flex items-center gap-2'>
              <span className='text-sm text-gray-700'>Number of Students:</span>
              <Badge variant='secondary' className='text-xs'>
                {section.data.length}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <ScrollArea className='h-[300px] rounded-md border'>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>First Name</TableHead>
                    <TableHead>Last Name</TableHead>
                    <TableHead>Matriculation Number</TableHead>
                    {section.showRank && <TableHead>Rank</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {section.data.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={section.showRank ? 4 : 3} className='text-center'>
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
                        {section.showRank && (
                          <TableCell>
                            {useScoreAsRank ? (
                              'rank' in item && item.rank ? (
                                item.rank
                              ) : (
                                <span className='text-destructive'>{'No Rank'}</span>
                              )
                            ) : (
                              1
                            )}
                          </TableCell>
                        )}
                      </TableRow>
                    ),
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          </CardContent>
        </Card>
      ))}

      <div className='mt-8'>
        <Button
          onClick={() =>
            generateAndDownloadFile([...matchedByMatriculation, ...matchedByName], useScoreAsRank)
          }
          className='w-full'
        >
          <Download />
          Download Filled Out Excel Sheet
        </Button>
      </div>
    </div>
  )
}

export default MatchingResults
