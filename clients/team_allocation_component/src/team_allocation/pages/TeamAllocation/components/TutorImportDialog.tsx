import { useState, useEffect } from 'react'
import { Loader2, UserPlus } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCourseStore, Student } from '@tumaet/prompt-shared-state'
import { getStudentsOfCoursePhase } from '../../../network/queries/getStudentsOfCoursePhase'
import { importTutors } from '../../../network/mutations/importTutors'
import { TutorSelection } from './TutorSelection'
import { Tutor } from '../../../interfaces/tutor'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Label,
} from '@tumaet/prompt-ui-components'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import type { Team } from '@tumaet/prompt-shared-state'

export function TutorImportDialog() {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { courses } = useCourseStore()
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [selectedSourceCourse, setSelectedSourceCourse] = useState<string | null>(null)
  const [selectedSourcePhase, setSelectedSourcePhase] = useState<string | null>(null)
  const [selectedStudents, setSelectedStudents] = useState<string[]>([])
  const [tutorTeamAssignments, setTutorTeamAssignments] = useState<Record<string, string>>({})
  const [isImporting, setIsImporting] = useState(false)
  const [importError, setImportError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) {
      setImportError(null)
      setSelectedSourceCourse(null)
      setSelectedSourcePhase(null)
      setSelectedStudents([])
      setTutorTeamAssignments({})
    }
  }, [open])

  const handleCourseChange = (courseID: string) => {
    setSelectedSourceCourse(courseID)
    setSelectedSourcePhase(null)
    setSelectedStudents([])
    setTutorTeamAssignments({})
  }

  const handlePhaseChange = (phaseID: string) => {
    setSelectedSourcePhase(phaseID)
    setSelectedStudents([])
    setTutorTeamAssignments({})
  }

  const handleStudentToggle = (studentId: string) => {
    setSelectedStudents((prev) =>
      prev.includes(studentId) ? prev.filter((id) => id !== studentId) : [...prev, studentId],
    )
  }

  const handleSelectAll = (students: Student[]) => {
    if (selectedStudents.length === students.length) {
      setSelectedStudents([])
    } else {
      setSelectedStudents(students.map((s) => s.id).filter((id): id is string => !!id))
    }
  }

  const handleTeamChange = (studentId: string, teamId: string) => {
    setTutorTeamAssignments((prev) => ({ ...prev, [studentId]: teamId }))
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

  const { data: fetchedTeams } = useQuery<Team[]>({
    queryKey: ['team_allocation_team', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { mutate: mutateImportTutors } = useMutation({
    mutationFn: (tutors: Tutor[]) => importTutors(phaseId ?? '', tutors),
    onSuccess: () => {
      setOpen(false)
      setIsImporting(false)
      setImportError(null)
      queryClient.invalidateQueries({ queryKey: ['tutors', phaseId] })
    },
    onError: (error: unknown) => {
      console.error('Error importing tutors:', error)
      setIsImporting(false)
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to import tutors. Please try again.'
      setImportError(`Import failed: ${errorMessage}`)
    },
  })

  const handleImport = () => {
    if (!selectedSourceCourse || !selectedSourcePhase || selectedStudents.length === 0) return

    const selectedStudentData = (students || []).filter(
      (s) => s.id && selectedStudents.includes(s.id),
    )
    const studentsAsTutors: Tutor[] = selectedStudentData.map((s) => ({
      CoursePhaseID: phaseId ?? '',
      CourseParticipationID: s.id!,
      FirstName: s.firstName,
      LastName: s.lastName,
      TeamID: tutorTeamAssignments[s.id!] || '',
      UniversityLogin: s.universityLogin ?? '',
    }))

    const allAssigned = studentsAsTutors.every((t) => t.TeamID !== '')
    if (!allAssigned) {
      setImportError('Please assign a team for each selected tutor.')
      setIsImporting(false)
      return
    }

    setIsImporting(true)
    mutateImportTutors(studentsAsTutors)
  }

  const currentSourceCourse = selectedSourceCourse
    ? courses.find((c) => c.id === selectedSourceCourse)
    : null

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <UserPlus className='mr-2 h-4 w-4' />
          Import Tutors
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[525px]'>
        <DialogHeader>
          <DialogTitle>Import Tutors</DialogTitle>
          <DialogDescription>
            Select students from another course and assign them to teams.
          </DialogDescription>
        </DialogHeader>

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

          {/* Student Selection */}
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
                  teamOptions={(fetchedTeams ?? []).map((t) => ({ id: t.id, name: t.name }))}
                  teamAssignments={tutorTeamAssignments}
                  onTeamChange={handleTeamChange}
                />
              )}
            </div>
          )}

          {importError && <div className='text-red-500 text-sm'>{importError}</div>}
        </div>

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
      </DialogContent>
    </Dialog>
  )
}
