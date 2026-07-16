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
import { useState } from 'react'
import type { Tutor } from '../../../interfaces/tutor'
import { getStudentsOfCoursePhase } from '../../../network/queries/getStudentsOfCoursePhase'
import { TutorSelection } from './TutorSelection'

interface TutorImportFromCourseProps {
  phaseId: string
  teamOptions: { id: string; name: string }[]
  isImporting: boolean
  onImport: (tutors: Tutor[]) => void
}

export const TutorImportFromCourse = ({
  phaseId,
  teamOptions,
  isImporting,
  onImport,
}: TutorImportFromCourseProps) => {
  const { courses } = useCourseStore()

  const [selectedSourceCourse, setSelectedSourceCourse] = useState<string | null>(null)
  const [selectedSourcePhase, setSelectedSourcePhase] = useState<string | null>(null)
  const [selectedStudents, setSelectedStudents] = useState<string[]>([])
  const [teamAssignments, setTeamAssignments] = useState<Record<string, string>>({})

  const handleCourseChange = (courseID: string) => {
    setSelectedSourceCourse(courseID)
    setSelectedSourcePhase(null)
    setSelectedStudents([])
    setTeamAssignments({})
  }

  const handlePhaseChange = (phaseID: string) => {
    setSelectedSourcePhase(phaseID)
    setSelectedStudents([])
    setTeamAssignments({})
  }

  const handleStudentToggle = (studentId: string) => {
    setSelectedStudents((prev) =>
      prev.includes(studentId) ? prev.filter((id) => id !== studentId) : [...prev, studentId],
    )
  }

  const handleSelectAll = (allStudents: Student[]) => {
    const selectableIds = allStudents
      .filter((s) => s.universityLogin)
      .map((s) => s.id)
      .filter((id): id is string => !!id)
    if (selectedStudents.length === selectableIds.length) {
      setSelectedStudents([])
    } else {
      setSelectedStudents(selectableIds)
    }
  }

  const handleTeamChange = (studentId: string, teamId: string) => {
    setTeamAssignments((prev) => ({ ...prev, [studentId]: teamId }))
  }

  const {
    data: students,
    isLoading: isStudentsLoading,
    isError: isStudentsError,
  } = useQuery<Student[]>({
    queryKey: ['students', selectedSourcePhase],
    queryFn: () => {
      if (selectedSourceCourse && selectedSourcePhase) {
        return getStudentsOfCoursePhase(selectedSourcePhase)
      }
      return Promise.resolve([])
    },
    enabled: !!selectedSourceCourse && !!selectedSourcePhase,
  })

  const handleImport = () => {
    const selectedStudentData = (students || []).filter(
      (s) => s.id && selectedStudents.includes(s.id),
    )
    const studentsAsTutors: Tutor[] = selectedStudentData.map((s) => ({
      CoursePhaseID: phaseId,
      CourseParticipationID: s.id!,
      FirstName: s.firstName,
      LastName: s.lastName,
      TeamID: teamAssignments[s.id!] || '',
      UniversityLogin: s.universityLogin ?? '',
    }))
    onImport(studentsAsTutors)
  }

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
              tutors={students}
              selectedTutors={selectedStudents}
              onTutorToggle={handleStudentToggle}
              onSelectAll={() => handleSelectAll(students)}
              teamOptions={teamOptions}
              teamAssignments={teamAssignments}
              onTeamChange={handleTeamChange}
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
            selectedStudents.length === 0 ||
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
