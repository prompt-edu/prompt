import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
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
import { selfTeamAllocationCache, selfTeamAllocationKeys } from '../../../network/cache'
import { importTutors } from '../../../network/mutations/importTutors'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import { getTutorAssignments } from '../../../network/queries/getTutorAssignments'
import { currentTutorTeams } from '../utils/tutorImport'
import { TutorImportFromCourse } from './TutorImportFromCourse'
import { TutorImportFromStaff } from './TutorImportFromStaff'

export function TutorImportDialog() {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const queryClient = useQueryClient()

  const [open, setOpen] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importError, setImportError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) {
      setImportError(null)
      setIsImporting(false)
      queryClient.removeQueries({ queryKey: selfTeamAllocationKeys.tutors(phaseId) })
    }
  }, [open, queryClient, phaseId])

  const { data: fetchedTeams } = useQuery<Team[]>({
    queryKey: selfTeamAllocationKeys.teams(phaseId),
    queryFn: () => getAllTeams(phaseId ?? ''),
    enabled: !!phaseId,
  })

  // The team pickers take their starting value once, so each opening waits for a fresh fetch.
  const {
    data: tutorAssignments,
    isPending: isTutorAssignmentsPending,
    isError: isTutorAssignmentsError,
  } = useQuery<TutorAssignment[]>({
    queryKey: selfTeamAllocationKeys.tutors(phaseId),
    queryFn: () => getTutorAssignments(phaseId ?? ''),
    enabled: !!phaseId && open,
  })

  const { mutate: mutateImportTutors } = useMutation({
    mutationFn: (tutors: Tutor[]) => importTutors(phaseId ?? '', tutors),
    onSuccess: () => {
      setOpen(false)
      setIsImporting(false)
      setImportError(null)
      selfTeamAllocationCache.tutorsImported(queryClient, phaseId)
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
        <Button variant='outline'>
          <UserPlus className='h-4 w-4' />
          Import Tutors
        </Button>
      </DialogTrigger>
      <DialogContent className='max-w-[600px] max-h-[90vh] flex flex-col [&>*]:min-w-0'>
        <DialogHeader>
          <DialogTitle>Import Tutors</DialogTitle>
          <DialogDescription>
            Import tutors from another course phase or from the lecturers and editors of this
            course. A tutor can lead several teams. Importing a tutor replaces the teams they
            already lead, so each person starts with their current teams.
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
          // Scroll the body, not the dialog, so the close button stays pinned.
          <Tabs
            defaultValue='course'
            onValueChange={() => setImportError(null)}
            className='min-h-0 overflow-y-auto'
          >
            <TabsList className='grid w-full grid-cols-2'>
              <TabsTrigger value='course'>From Course</TabsTrigger>
              <TabsTrigger value='staff'>Course Staff</TabsTrigger>
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
          </Tabs>
        )}

        {importError && <div className='text-red-500 text-sm'>{importError}</div>}
      </DialogContent>
    </Dialog>
  )
}
