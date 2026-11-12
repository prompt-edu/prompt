import { useQuery } from '@tanstack/react-query'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@tumaet/prompt-ui-components'
import { AlertCircle, ArrowRight, CheckCircle2, Loader2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import type { TeaseStudent } from '../../../interfaces/tease/student'
import { teamAllocationKeys } from '../../../network/cache'
import { getAllTeaseStudents } from '../../../network/queries/getAllTeaseStudents'
import { getLaunchWarnings } from '../../../utils/studentDataChecks'
import { NOT_SUBMITTED_FILTER_SEARCH } from '../../../utils/surveyStatus'

export const LaunchTeaseDialog = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const [open, setOpen] = useState(false)

  const studentsQuery = useQuery<TeaseStudent[]>({
    queryKey: teamAllocationKeys.tease.students(phaseId),
    queryFn: () => getAllTeaseStudents(phaseId ?? ''),
    enabled: open && !!phaseId,
  })

  const warnings = useMemo(
    () => (studentsQuery.data ? getLaunchWarnings(studentsQuery.data) : null),
    [studentsQuery.data],
  )

  const phasePath = `/management/course/${courseId ?? ''}/${phaseId ?? ''}`
  const hasWarnings =
    warnings !== null &&
    (warnings.incompleteChecks.length > 0 || warnings.studentsWithoutSurvey > 0)

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          Launch Tease to Matchmake
          <ArrowRight className='ml-2 h-4 w-4' />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Launch TEASE</DialogTitle>
          <DialogDescription>
            TEASE opens in a new tab. It matches the students to the teams with their survey answers
            and the data from the previous phases.
          </DialogDescription>
        </DialogHeader>

        {studentsQuery.isPending && (
          <div className='flex items-center gap-2 text-sm text-muted-foreground'>
            <Loader2 className='h-4 w-4 animate-spin' />
            Checking the student data...
          </div>
        )}

        {studentsQuery.isError && (
          <div className='flex items-start gap-2 text-sm text-amber-700'>
            <AlertCircle className='h-4 w-4 mt-0.5 shrink-0' />
            <p>Could not check the student data. You can still launch TEASE.</p>
          </div>
        )}

        {warnings !== null && !hasWarnings && (
          <div className='flex items-start gap-2 text-sm text-green-700'>
            <CheckCircle2 className='h-4 w-4 mt-0.5 shrink-0' />
            <p>All student data is complete.</p>
          </div>
        )}

        {warnings !== null && warnings.incompleteChecks.length > 0 && (
          <div className='flex items-start gap-2 text-sm text-amber-700'>
            <AlertCircle className='h-4 w-4 mt-0.5 shrink-0' />
            <p>
              Data from previous phases is missing: {warnings.incompleteChecks.join(', ')}.{' '}
              <Link to={`${phasePath}/settings`} className='underline font-medium'>
                Open the settings
              </Link>
            </p>
          </div>
        )}

        {warnings !== null && warnings.studentsWithoutSurvey > 0 && (
          <div className='flex items-start gap-2 text-sm text-amber-700'>
            <AlertCircle className='h-4 w-4 mt-0.5 shrink-0' />
            <p>
              {warnings.studentsWithoutSurvey} of {warnings.totalStudents} students have not
              submitted the survey.{' '}
              <Link
                to={`${phasePath}/participants${NOT_SUBMITTED_FILTER_SEARCH}`}
                className='underline font-medium'
              >
                Show these students
              </Link>
            </p>
          </div>
        )}

        <DialogFooter>
          <Button variant='outline' onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button asChild>
            <a
              href={`/tease?coursePhaseId=${phaseId ?? ''}`}
              target='_blank'
              rel='noopener noreferrer'
              onClick={() => setOpen(false)}
            >
              {hasWarnings ? 'Launch TEASE anyway' : 'Launch TEASE'}
              <ArrowRight className='ml-2 h-4 w-4' />
            </a>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
