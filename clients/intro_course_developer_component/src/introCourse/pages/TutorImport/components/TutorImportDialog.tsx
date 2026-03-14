import { useState, useEffect } from 'react'
import { Loader2, UserPlus } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCourseStore, Student } from '@tumaet/prompt-shared-state'
import { getStudentsOfCoursePhase } from '../../../network/queries/getStudentsOfCoursePhase'
import { importTutors } from '../../../network/mutations/importTutors'
import { StudentSelection } from './StudentSelection'
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

export function TutorImportDialog() {
  // Destination course/phase come from URL parameters.
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { courses } = useCourseStore()
  const queryClient = useQueryClient()

  // Local state for dialog open, selections, loading & error states.
  const [open, setOpen] = useState(false)
  const [selectedSourceCourse, setSelectedSourceCourse] = useState<string | null>(null)
  const [selectedSourcePhase, setSelectedSourcePhase] = useState<string | null>(null)
  const [selectedStudents, setSelectedStudents] = useState<string[]>([])
  const [isImporting, setIsImporting] = useState(false)
  const [importError, setImportError] = useState<string | null>(null)

  // Reset error and selections when the dialog closes.
  useEffect(() => {
    if (!open) {
      setImportError(null)
      setSelectedSourceCourse(null)
      setSelectedSourcePhase(null)
      setSelectedStudents([])
    }
  }, [open])

  // Handle course and phase changes.
  const handleCourseChange = (courseID: string) => {
    setSelectedSourceCourse(courseID)
    setSelectedSourcePhase(null)
    setSelectedStudents([])
  }

  const handlePhaseChange = (phaseID: string) => {
    setSelectedSourcePhase(phaseID)
    setSelectedStudents([])
  }

  // Toggle individual student selection.
  const handleStudentToggle = (studentId: string) => {
    setSelectedStudents((prev) =>
      prev.includes(studentId) ? prev.filter((id) => id !== studentId) : [...prev, studentId],
    )
  }

  // (De)select all students.
  const handleSelectAll = (students: Student[]) => {
    if (selectedStudents.length === students.length) {
      setSelectedStudents([])
    } else {
      function isString(value: string | undefined): value is string {
        return value !== undefined
      }

      setSelectedStudents(students.map((s) => s.id).filter(isString))
    }
  }

  // Fetch students based on the selected source course and phase.
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

  // Mutation for importing tutors into the destination course/phase.
  const { mutate: mutateImportTutors } = useMutation({
    mutationFn: (tutors: Student[]) => importTutors(phaseId ?? '', courseId ?? '', tutors),
    onSuccess: () => {
      setOpen(false)
      setIsImporting(false)
      setImportError(null)
      queryClient.invalidateQueries({ queryKey: ['tutors', phaseId] })
    },
    onError: (error: unknown) => {
      console.error('Error importing tutors:', error)
      setIsImporting(false)
      setImportError('Failed to import tutors. Please try again.')
    },
  })

  // Handle the import action.
  const handleImport = () => {
    if (!selectedSourceCourse || !selectedSourcePhase || selectedStudents.length === 0) return

    setIsImporting(true)
    const selectedStudentData = (students || []).filter(
      (s) => s.id && selectedStudents.includes(s.id),
    )
    mutateImportTutors(selectedStudentData)
  }

  // Get the selected source course details to list its phases.
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
            Select students from another course to import as tutors.
          </DialogDescription>
        </DialogHeader>

        <div className='grid gap-4 py-4'>
          {/* Course Selection */}
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
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Phase Selection */}
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
                <StudentSelection
                  students={students}
                  selectedStudents={selectedStudents}
                  onStudentToggle={handleStudentToggle}
                  onSelectAll={() => handleSelectAll(students)}
                />
              )}
            </div>
          )}

          {/* Display any import error */}
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
              'Import Selected Students'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
