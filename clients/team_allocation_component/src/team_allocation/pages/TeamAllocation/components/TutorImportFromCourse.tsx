import { useQuery } from '@tanstack/react-query'
import { type Student, useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  DialogFooter,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useMemo, useState } from 'react'
import type { Tutor } from '../../../interfaces/tutor'
import { teamAllocationKeys } from '../../../network/cache'
import { getStudentsOfCoursePhase } from '../../../network/queries/getStudentsOfCoursePhase'
import { useTutorSelection } from '../hooks/useTutorSelection'
import { studentToTutorCandidate } from '../utils/tutorImport'
import { TutorSelection } from './TutorSelection'

interface TutorImportFromCourseProps {
  teamOptions: { id: string; name: string }[]
  isImporting: boolean
  onImport: (tutors: Tutor[]) => void
}

export const TutorImportFromCourse = ({
  teamOptions,
  isImporting,
  onImport,
}: TutorImportFromCourseProps) => {
  const { courses } = useCourseStore()

  const [selectedSourceCourse, setSelectedSourceCourse] = useState<string | null>(null)
  const [selectedSourcePhase, setSelectedSourcePhase] = useState<string | null>(null)
  const selection = useTutorSelection()

  const handleCourseChange = (courseID: string) => {
    setSelectedSourceCourse(courseID)
    setSelectedSourcePhase(null)
    selection.reset()
  }

  const handlePhaseChange = (phaseID: string) => {
    setSelectedSourcePhase(phaseID)
    selection.reset()
  }

  const {
    data: students,
    isLoading: isStudentsLoading,
    isError: isStudentsError,
  } = useQuery<Student[]>({
    queryKey: teamAllocationKeys.tutorImport.studentsOfPhase(selectedSourcePhase ?? undefined),
    queryFn: () => {
      if (selectedSourceCourse && selectedSourcePhase) {
        return getStudentsOfCoursePhase(selectedSourcePhase)
      }
      return Promise.resolve([])
    },
    enabled: !!selectedSourceCourse && !!selectedSourcePhase,
  })

  const candidates = useMemo(() => (students ?? []).map(studentToTutorCandidate), [students])

  const handleImport = () => onImport(selection.buildImport(candidates))

  const currentSourceCourse = selectedSourceCourse
    ? courses.find((c) => c.id === selectedSourceCourse)
    : null

  return (
    <div className='grid gap-4 py-4'>
      <div className='grid gap-2'>
        <Label htmlFor='course'>Select Source Course</Label>
        <Select value={selectedSourceCourse || ''} onValueChange={handleCourseChange}>
          <SelectTrigger id='course'>
            <SelectValue placeholder='Select a course' />
          </SelectTrigger>
          <SelectContent>
            {courses.map((course) => (
              <SelectItem key={course.id} value={course.id}>
                {course.name}
                {' ('}
                {course.semesterTag}
                {')'}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {selectedSourceCourse && (
        <div className='grid gap-2'>
          <Label htmlFor='phase'>Select Source Course Phase</Label>
          <Select value={selectedSourcePhase || ''} onValueChange={handlePhaseChange}>
            <SelectTrigger id='phase'>
              <SelectValue placeholder='Select a phase' />
            </SelectTrigger>
            <SelectContent>
              {currentSourceCourse?.coursePhases.map((phase) => (
                <SelectItem key={phase.id} value={phase.id}>
                  {phase.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      {selectedSourcePhase && (
        <div className='grid gap-2'>
          {isStudentsLoading && <div>Loading students...</div>}
          {isStudentsError && (
            <div className='text-red-500'>Error loading students. Please try again.</div>
          )}
          {students && (
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
        </div>
      )}

      <DialogFooter>
        <Button
          type='submit'
          onClick={handleImport}
          disabled={
            !selectedSourceCourse ||
            !selectedSourcePhase ||
            selection.selectedTutors.length === 0 ||
            isImporting
          }
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
