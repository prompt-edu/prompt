import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { getInstructorNotes } from '../queries/getInstructorNotes'
import { deleteInstructorNote } from '../mutations/deleteInstructorNote'
import { postInstructorNote } from '../mutations/postInstructorNote'
import type { CreateInstructorNote } from '@core/managementConsole/shared/interfaces/InstructorNote'

export const useInstructorNotes = (studentId?: string) => {
  return useQuery({
    queryKey: ['instructorNotes', studentId],
    queryFn: () => getInstructorNotes(studentId!),
    enabled: !!studentId,
  })
}

export const useDeleteInstructorNote = (studentId?: string) => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (noteId: string) => deleteInstructorNote(noteId),
    onSuccess: () => {
      toast({
        title: 'Note deleted successfully',
      })
      queryClient.invalidateQueries({ queryKey: ['instructorNotes', studentId] })
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
    mutationFn: (note: CreateInstructorNote) => postInstructorNote(studentId, note),
    onSuccess: () => {
      toast({
        title: 'Note saved successfully',
      })
      queryClient.invalidateQueries({ queryKey: ['instructorNotes', studentId] })
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
