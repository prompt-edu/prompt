import type { CreateInstructorNote } from '@core/managementConsole/shared/interfaces/InstructorNote'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'

export const useInstructorNotes = (studentId?: string) => {
  return useQuery({
    queryKey: coreKeys.instructorNotes.ofStudent(studentId),
    queryFn: () => coreApi.instructorNotes.ofStudent(studentId!),
    enabled: !!studentId,
  })
}

export const useDeleteInstructorNote = (studentId?: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (noteId: string) => coreApi.instructorNotes.remove(noteId),
    onSuccess: () => {
      coreCache.instructorNotesChanged(queryClient, studentId)
    },
    onError: () => {
      toast({
        title: 'Failed to delete note',
        description: 'Are you sure you have the right permissions?',
        variant: 'destructive',
      })
    },
  })
}

export const useCreateInstructorNote = (studentId: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (note: CreateInstructorNote) => coreApi.instructorNotes.create(studentId, note),
    onSuccess: () => {
      coreCache.instructorNotesChanged(queryClient, studentId)
    },
    onError: () => {
      toast({
        title: 'Failed to save note',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}
