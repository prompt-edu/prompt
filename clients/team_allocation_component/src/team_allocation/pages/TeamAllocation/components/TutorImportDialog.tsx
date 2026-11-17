import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Role, type Team, useAuthStore } from '@tumaet/prompt-shared-state'
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'
import { Loader2, UserPlus } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Tutor, TutorAssignment } from '../../../interfaces/tutor'
import { teamAllocationCache, teamAllocationKeys } from '../../../network/cache'
import { importTutors } from '../../../network/mutations/importTutors'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import { getTutorAssignments } from '../../../network/queries/getTutorAssignments'
import { currentTutorTeams } from '../utils/tutorImport'
import { TutorImportFromCourse } from './TutorImportFromCourse'
import { TutorImportFromStaff } from './TutorImportFromStaff'
import { TutorSearchStudents } from './TutorSearchStudents'

export function TutorImportDialog() {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const queryClient = useQueryClient()
  const { permissions } = useAuthStore()
  const canSearchStudents = permissions.some(
    (p) => p === Role.PROMPT_ADMIN || p === Role.PROMPT_LECTURER,
  )

  const [open, setOpen] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importError, setImportError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) {
      setImportError(null)
      setIsImporting(false)
    }
  }, [open])

  const { data: fetchedTeams } = useQuery<Team[]>({
    queryKey: teamAllocationKeys.teams(phaseId),
    queryFn: () => getAllTeams(phaseId ?? ''),
    enabled: !!phaseId,
  })

  // The team pickers take their starting value once, so the sources wait for the current tutors.
  const {
    data: tutorAssignments,
    isPending: isTutorAssignmentsPending,
    isError: isTutorAssignmentsError,
  } = useQuery<TutorAssignment[]>({
    queryKey: teamAllocationKeys.tutors(phaseId),
    queryFn: () => getTutorAssignments(phaseId ?? ''),
    enabled: !!phaseId && open,
  })

  const { mutate: mutateImportTutors } = useMutation({
    mutationFn: (tutors: Tutor[]) => importTutors(phaseId ?? '', tutors),
    onSuccess: () => {
      setOpen(false)
      setIsImporting(false)
      setImportError(null)
      teamAllocationCache.tutorsImported(queryClient, phaseId)
    },
    onError: (error: unknown) => {
      console.error('Error importing tutors:', error)
      setIsImporting(false)
      const backendError = isAxiosError(error) ? error.response?.data?.error : undefined
      const errorMessage =
        backendError ??
        (error instanceof Error ? error.message : 'Failed to import tutors. Please try again.')
      setImportError(`Import failed: ${errorMessage}`)
    },
  })

  const handleImport = (tutors: Tutor[]) => {
    if (tutors.length === 0) return

    const allAssigned = tutors.every((t) => t.teamIDs.length > 0)
    if (!allAssigned) {
      setImportError('Please assign at least one team to each selected tutor.')
      return
    }

    setImportError(null)
    setIsImporting(true)
    mutateImportTutors(tutors)
  }

  const teamOptions = (fetchedTeams ?? []).map((t) => ({ id: t.id, name: t.name }))
  const currentTeams = useMemo(() => currentTutorTeams(tutorAssignments ?? []), [tutorAssignments])

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <UserPlus className='mr-2 h-4 w-4' />
          Import Tutors
        </Button>
      </DialogTrigger>
      <DialogContent className='max-w-[525px] [&>*]:min-w-0'>
        <DialogHeader>
          <DialogTitle>Import Tutors</DialogTitle>
          <DialogDescription>
            Import tutors from another course phase, from the editors of this course, or search all
            students to add them directly. A tutor can lead several teams. Importing a tutor
            replaces the teams they already lead, so each person starts with their current teams.
          </DialogDescription>
        </DialogHeader>

        {isTutorAssignmentsError && (
          <div className='text-red-500 text-sm'>
            Error loading the current tutors. Please try again.
          </div>
        )}
        {isTutorAssignmentsPending && !isTutorAssignmentsError && (
          <div className='flex items-center justify-center py-4'>
            <Loader2 className='h-5 w-5 animate-spin text-muted-foreground' />
          </div>
        )}
        {tutorAssignments && (
          <Tabs defaultValue='course' onValueChange={() => setImportError(null)}>
            <TabsList
              className={`grid w-full ${canSearchStudents ? 'grid-cols-3' : 'grid-cols-2'}`}
            >
              <TabsTrigger value='course'>From Course</TabsTrigger>
              <TabsTrigger value='staff'>Course Staff</TabsTrigger>
              {canSearchStudents && <TabsTrigger value='search'>Search Students</TabsTrigger>}
            </TabsList>

            <TabsContent value='course'>
              <TutorImportFromCourse
                teamOptions={teamOptions}
                currentTeams={currentTeams}
                isImporting={isImporting}
                onImport={handleImport}
              />
            </TabsContent>

            <TabsContent value='staff'>
              <TutorImportFromStaff
                courseId={courseId ?? ''}
                teamOptions={teamOptions}
                currentTeams={currentTeams}
                isImporting={isImporting}
                onImport={handleImport}
              />
            </TabsContent>

            {canSearchStudents && (
              <TabsContent value='search'>
                <TutorSearchStudents
                  teamOptions={teamOptions}
                  currentTeams={currentTeams}
                  isImporting={isImporting}
                  onImport={handleImport}
                />
              </TabsContent>
            )}
          </Tabs>
        )}

        {importError && <div className='text-red-500 text-sm'>{importError}</div>}
      </DialogContent>
    </Dialog>
  )
}
