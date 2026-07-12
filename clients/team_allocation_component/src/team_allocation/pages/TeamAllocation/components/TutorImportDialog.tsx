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
import { UserPlus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import type { Tutor } from '../../../interfaces/tutor'
import { importTutors } from '../../../network/mutations/importTutors'
import { getAllTeams } from '../../../network/queries/getAllTeams'
import { TutorImportFromCourse } from './TutorImportFromCourse'
import { TutorSearchStudents } from './TutorSearchStudents'

export function TutorImportDialog() {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

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

  const handleImport = (tutors: Tutor[]) => {
    if (tutors.length === 0) return

    const allAssigned = tutors.every((t) => t.TeamID !== '')
    if (!allAssigned) {
      setImportError('Please assign a team for each selected tutor.')
      return
    }

    setImportError(null)
    setIsImporting(true)
    mutateImportTutors(tutors)
  }

  const teamOptions = (fetchedTeams ?? []).map((t) => ({ id: t.id, name: t.name }))

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
            Import tutors from another course phase or search all students to add them directly.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue='course' onValueChange={() => setImportError(null)}>
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='course'>From Course</TabsTrigger>
            <TabsTrigger value='search'>Search Students</TabsTrigger>
          </TabsList>

          <TabsContent value='course'>
            <TutorImportFromCourse
              phaseId={phaseId ?? ''}
              teamOptions={teamOptions}
              isImporting={isImporting}
              onImport={handleImport}
            />
          </TabsContent>

          <TabsContent value='search'>
            <TutorSearchStudents
              phaseId={phaseId ?? ''}
              teamOptions={teamOptions}
              isImporting={isImporting}
              onImport={handleImport}
            />
          </TabsContent>
        </Tabs>

        {importError && <div className='text-red-500 text-sm'>{importError}</div>}
      </DialogContent>
    </Dialog>
  )
}
