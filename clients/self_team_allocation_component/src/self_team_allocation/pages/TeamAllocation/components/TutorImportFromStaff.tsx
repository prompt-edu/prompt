import { useQuery } from '@tanstack/react-query'
import { Button, DialogFooter } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useMemo } from 'react'
import type { CourseStaff } from '../../../interfaces/courseStaff'
import type { Tutor } from '../../../interfaces/tutor'
import { selfTeamAllocationKeys } from '../../../network/cache'
import { getCourseStaff } from '../../../network/queries/getCourseStaff'
import { useTutorSelection } from '../hooks/useTutorSelection'
import { type CurrentTutorTeams, staffToTutorCandidates } from '../utils/tutorImport'
import { TutorSelection } from './TutorSelection'

interface TutorImportFromStaffProps {
  courseId: string
  teamOptions: { id: string; name: string }[]
  currentTeams: CurrentTutorTeams
  isImporting: boolean
  onImport: (tutors: Tutor[]) => void
}

// Imports course lecturers and editors, who have no student record (the project leads of the
// iPraktikum, for example), as tutors of one or more teams.
export const TutorImportFromStaff = ({
  courseId,
  teamOptions,
  currentTeams,
  isImporting,
  onImport,
}: TutorImportFromStaffProps) => {
  const selection = useTutorSelection(currentTeams)

  const {
    data: staff,
    isLoading,
    isError,
  } = useQuery<CourseStaff>({
    queryKey: selfTeamAllocationKeys.tutorImport.courseStaff(courseId),
    queryFn: () => getCourseStaff(courseId),
    enabled: !!courseId,
  })

  const candidates = useMemo(() => (staff ? staffToTutorCandidates(staff) : []), [staff])

  return (
    <div className='grid gap-4 py-4'>
      <p className='text-sm text-muted-foreground'>
        Course lecturers and editors of this course. Assign each one the teams they lead.
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
        <p className='text-sm text-muted-foreground'>This course has no lecturers or editors.</p>
      )}
      {candidates.length > 0 && (
        <TutorSelection
          tutors={candidates}
          selectedTutors={selection.selectedTutors}
          onTutorToggle={selection.toggleTutor}
          onSelectAll={() => selection.toggleAll(candidates)}
          teamOptions={teamOptions}
          teamsOf={selection.teamsOf}
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
