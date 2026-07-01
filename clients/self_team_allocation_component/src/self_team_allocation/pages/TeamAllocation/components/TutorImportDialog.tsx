import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type Student, type Team, useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { Loader2, UserPlus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Tutor } from '../../../interfaces/tutor'
import { importTutors } from '../../../network/mutations/importTutors'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import { getStudentsOfCoursePhase } from '../../../network/queries/getStudentsOfCoursePhase'

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

  const handleSelectAll = (allStudents: Student[]) => {
    if (selectedStudents.length === allStudents.length) {
      setSelectedStudents([])
    } else {
      setSelectedStudents(allStudents.map((s) => s.id).filter((id): id is string => !!id))
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
    queryKey: ['self_team_allocations', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const { mutate: mutateImportTutors } = useMutation({
    mutationFn: (tutors: Tutor[]) => importTutors(phaseId ?? '', tutors),
    onSuccess: () => {
      setOpen(false)
      setIsImporting(false)
      setImportError(null)
      queryClient.invalidateQueries({ queryKey: ['self_team_allocations', phaseId] })
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
      coursePhaseID: phaseId ?? '',
      courseParticipationID: s.id!,
      firstName: s.firstName,
      lastName: s.lastName,
      teamID: tutorTeamAssignments[s.id!] || '',
    }))

    const allAssigned = studentsAsTutors.every((t) => t.teamID !== '')
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
        <Button variant='outline'>
          <UserPlus className='mr-2 h-4 w-4' />
          Import Tutors
        </Button>
      </DialogTrigger>
      <DialogContent className='sm:max-w-[600px] max-h-[90vh] overflow-y-auto'>
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
                    {course.name} ({course.semesterTag})
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
              {students && students.length > 0 && (
                <>
                  <div className='flex items-center justify-between'>
                    <Label>Select Students as Tutors</Label>
                    <Button variant='ghost' size='sm' onClick={() => handleSelectAll(students)}>
                      {selectedStudents.length === students.length ? 'Deselect All' : 'Select All'}
                    </Button>
                  </div>
                  <div className='border rounded-md p-4 max-h-[300px] overflow-y-auto space-y-3'>
                    {students.map((student) => (
                      <div key={student.id} className='space-y-2'>
                        <div className='flex items-center space-x-2'>
                          <Checkbox
                            id={`student-${student.id}`}
                            checked={selectedStudents.includes(student.id!)}
                            onCheckedChange={() => handleStudentToggle(student.id!)}
                          />
                          <label
                            htmlFor={`student-${student.id}`}
                            className='text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70'
                          >
                            {student.firstName} {student.lastName}
                          </label>
                        </div>
                        {selectedStudents.includes(student.id!) && (
                          <div className='ml-6 grid gap-2'>
                            <Label htmlFor={`team-${student.id}`} className='text-xs'>
                              Assign to Team
                            </Label>
                            <Select
                              value={tutorTeamAssignments[student.id!] || ''}
                              onValueChange={(teamId) => handleTeamChange(student.id!, teamId)}
                            >
                              <SelectTrigger id={`team-${student.id}`} className='h-8'>
                                <SelectValue placeholder='Select team' />
                              </SelectTrigger>
                              <SelectContent>
                                {(fetchedTeams ?? []).map((team) => (
                                  <SelectItem key={team.id} value={team.id}>
                                    {team.name}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </>
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
