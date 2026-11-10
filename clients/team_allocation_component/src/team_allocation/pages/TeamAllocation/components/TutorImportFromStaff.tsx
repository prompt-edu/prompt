import { useQuery } from '@tanstack/react-query'
import { Button, DialogFooter } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useMemo } from 'react'
import type { CourseStaff } from '../../../interfaces/courseStaff'
import type { Tutor } from '../../../interfaces/tutor'
import { teamAllocationKeys } from '../../../network/cache'
import { getCourseStaff } from '../../../network/queries/getCourseStaff'
import { useTutorSelection } from '../hooks/useTutorSelection'
import { staffMemberToTutorCandidate } from '../utils/tutorImport'
import { TutorSelection } from './TutorSelection'

interface TutorImportFromStaffProps {
  courseId: string
  teamOptions: { id: string; name: string }[]
  isImporting: boolean
  onImport: (tutors: Tutor[]) => void
}

// Imports course editors, who have no student record (the project leads of the iPraktikum, for
// example), as tutors of one or more teams.
export const TutorImportFromStaff = ({
  courseId,
  teamOptions,
  isImporting,
  onImport,
}: TutorImportFromStaffProps) => {
  const selection = useTutorSelection()

  const {
    data: staff,
    isLoading,
    isError,
  } = useQuery<CourseStaff>({
    queryKey: teamAllocationKeys.tutorImport.courseStaff(courseId),
    queryFn: () => getCourseStaff(courseId),
    enabled: !!courseId,
  })

  const candidates = useMemo(() => (staff?.editors ?? []).map(staffMemberToTutorCandidate), [staff])

  return (
    <div className='grid gap-4 py-4'>
      <p className='text-sm text-muted-foreground'>
        Course editors of this course. Assign each one the teams they lead.
      </p>
      {isLoading && (
        <div className='flex items-center justify-center py-4'>
          <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
        </div>
      )}
      {isError && (
        <div className='text-red-500 text-sm'>
          Error loading the course staff. Please try again.
        </div>
      )}
      {staff && candidates.length === 0 && (
        <p className='text-sm text-muted-foreground'>This course has no course editors.</p>
      )}
      {candidates.length > 0 && (
        <TutorSelection
          tutors={candidates}
          selectedTutors={selection.selectedTutors}
          onTutorToggle={selection.toggleTutor}
          onSelectAll={() => selection.toggleAll(candidates)}
          teamOptions={teamOptions}
          teamAssignments={selection.teamAssignments}
          onTeamChange={selection.setTeams}
        />
      )}

      <DialogFooter>
        <Button
          type='submit'
          onClick={() => onImport(selection.buildImport(candidates))}
          disabled={selection.selectedTutors.length === 0 || isImporting}
        >
          {isImporting ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Importing...
            </>
          ) : (
            'Import Selected Tutors'
          )}
        </Button>
      </DialogFooter>
    </div>
  )
}
