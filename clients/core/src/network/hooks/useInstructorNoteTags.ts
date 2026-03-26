import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { getNoteTags } from '../queries/getNoteTags'
import { postNoteTag } from '../mutations/postNoteTag'
import { putNoteTag } from '../mutations/putNoteTag'
import { deleteNoteTag } from '../mutations/deleteNoteTag'
import type {
  CreateNoteTag,
  UpdateNoteTag,
} from '@core/managementConsole/shared/interfaces/InstructorNote'

export const useNoteTags = () => {
  return useQuery({
    queryKey: ['noteTags'],
    queryFn: getNoteTags,
  })
}

export const useCreateNoteTag = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (tag: CreateNoteTag) => postNoteTag(tag),
    onSuccess: () => {
      toast({ title: 'Tag created successfully' })
      queryClient.invalidateQueries({ queryKey: ['noteTags'] })
    },
    onError: () => {
      toast({
        title: 'Failed to create tag',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}

export const useUpdateNoteTag = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ tagId, tag }: { tagId: string; tag: UpdateNoteTag }) => putNoteTag(tagId, tag),
    onSuccess: () => {
      toast({ title: 'Tag updated successfully' })
      queryClient.invalidateQueries({ queryKey: ['noteTags'] })
    },
    onError: () => {
      toast({
        title: 'Failed to update tag',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}

export const useDeleteNoteTag = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (tagId: string) => deleteNoteTag(tagId),
    onSuccess: () => {
      toast({ title: 'Tag deleted successfully' })
      queryClient.invalidateQueries({ queryKey: ['noteTags'] })
    },
    onError: () => {
      toast({
        title: 'Failed to delete tag',
        description: 'Please try again later',
        variant: 'destructive',
      })
    },
  })
}
