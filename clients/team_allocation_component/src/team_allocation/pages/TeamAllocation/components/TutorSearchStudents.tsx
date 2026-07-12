import { useQuery } from '@tanstack/react-query'
import type { Student } from '@tumaet/prompt-shared-state'
import { Button, DialogFooter, Input, Label } from '@tumaet/prompt-ui-components'
import { Loader2, Search } from 'lucide-react'
import { useEffect, useState } from 'react'
import type { Tutor } from '../../../interfaces/tutor'
import { searchStudents } from '../../../network/queries/searchStudents'
import { TutorSelection } from './TutorSelection'

interface TutorSearchStudentsProps {
  phaseId: string
  teamOptions: { id: string; name: string }[]
  isImporting: boolean
  onImport: (tutors: Tutor[]) => void
}

const useDebouncedValue = <T,>(value: T, delayMs = 300): T => {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs)
    return () => clearTimeout(id)
  }, [value, delayMs])
  return debounced
}

export const TutorSearchStudents = ({
  phaseId,
  teamOptions,
  isImporting,
  onImport,
}: TutorSearchStudentsProps) => {
  const [query, setQuery] = useState('')
  const debouncedQuery = useDebouncedValue(query, 300)
  const [selectedStudents, setSelectedStudents] = useState<string[]>([])
  const [teamAssignments, setTeamAssignments] = useState<Record<string, string>>({})

  const trimmedQuery = debouncedQuery.trim()

  const {
    data: students,
    isFetching,
    isError,
  } = useQuery<Student[]>({
    queryKey: ['student-search', trimmedQuery],
    queryFn: () => searchStudents(trimmedQuery),
    enabled: trimmedQuery.length >= 2,
  })

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
    setTeamAssignments((prev) => ({ ...prev, [studentId]: teamId }))
  }

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

  return (
    <div className='grid gap-4 py-4'>
      <div className='grid gap-2'>
        <Label htmlFor='student-search'>Search Students</Label>
        <div className='relative w-full'>
          <Search className='pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
          <Input
            id='student-search'
            autoFocus
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder='Search by name, email, or login...'
            className='w-full pl-10'
          />
        </div>
      </div>

      {trimmedQuery.length > 0 && trimmedQuery.length < 2 && (
        <p className='text-sm text-muted-foreground'>Type at least 2 characters to search.</p>
      )}
      {trimmedQuery.length >= 2 && isFetching && (
        <div className='flex items-center justify-center py-4'>
          <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
        </div>
      )}
      {isError && <div className='text-red-500 text-sm'>Search failed. Please try again.</div>}
      {trimmedQuery.length >= 2 && !isFetching && students && students.length === 0 && (
        <p className='text-sm text-muted-foreground'>No students found.</p>
      )}
      {students && students.length > 0 && (
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

      <DialogFooter>
        <Button
          type='submit'
          onClick={handleImport}
          disabled={selectedStudents.length === 0 || isImporting}
        >
          {isImporting ? (
            <>
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              Adding...
            </>
          ) : (
            'Add Selected Tutors'
          )}
        </Button>
      </DialogFooter>
    </div>
  )
}
