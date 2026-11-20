import { useQuery } from '@tanstack/react-query'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { ClipboardCheck, Loader2, Users } from 'lucide-react'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import type { TeaseStudent } from '../../../../interfaces/tease/student'
import { teamAllocationKeys } from '../../../../network/cache'
import { getAllTeaseStudents } from '../../../../network/queries/getAllTeaseStudents'
import { runStudentDataChecks } from '../../../../utils/studentDataChecks'
import { CheckItem } from './CheckItem'
import { DataCompletionSummary } from './DataCompletionSummary'

export const PreviousPhaseDataCheck = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const studentsQuery = useQuery<TeaseStudent[]>({
    queryKey: teamAllocationKeys.tease.students(phaseId),
    queryFn: () => getAllTeaseStudents(phaseId ?? ''),
  })

  const students = studentsQuery.data

  const checks = useMemo(
    () => (students && students.length > 0 ? runStudentDataChecks(students) : []),
    [students],
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <ClipboardCheck className='h-5 w-5' />
          Previous Phase Data
        </CardTitle>
        <CardDescription>
          TEASE matches students to teams with data from the previous phases. These checks show if
          this data is present for all students.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <QueryGate
          queries={[studentsQuery]}
          loadingFallback={
            <div className='flex justify-center p-6'>
              <Loader2 className='h-8 w-8 animate-spin text-primary' />
            </div>
          }
        >
          {checks.length === 0 ? (
            <div className='p-6 text-center'>
              <Users className='h-12 w-12 text-muted-foreground mx-auto mb-4' />
              <h3 className='text-lg font-medium mb-2'>No student data available</h3>
              <p className='text-muted-foreground'>
                There are no students enrolled in this phase yet.
              </p>
            </div>
          ) : (
            <div className='space-y-4'>
              <DataCompletionSummary checks={checks} />
              {checks.map((check) => (
                <CheckItem key={check.label} check={check} />
              ))}
            </div>
          )}
        </QueryGate>
      </CardContent>
    </Card>
  )
}
