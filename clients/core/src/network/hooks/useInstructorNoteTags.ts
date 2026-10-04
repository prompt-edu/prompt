import type {
  CreateNoteTag,
  UpdateNoteTag,
} from '@core/managementConsole/shared/interfaces/InstructorNote'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'

export const useNoteTags = () => {
  return useQuery({
    queryKey: coreKeys.instructorNotes.tags(),
    queryFn: coreApi.instructorNotes.listTags,
  })
}

export const useCreateNoteTag = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (tag: CreateNoteTag) => coreApi.instructorNotes.createTag(tag),
    onSuccess: () => {
      toast({ title: 'Tag created successfully' })
      coreCache.noteTagsChanged(queryClient)
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
    mutationFn: ({ tagId, tag }: { tagId: string; tag: UpdateNoteTag }) =>
      coreApi.instructorNotes.updateTag(tagId, tag),
    onSuccess: () => {
      toast({ title: 'Tag updated successfully' })
      coreCache.noteTagsChanged(queryClient)
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
    mutationFn: (tagId: string) => coreApi.instructorNotes.removeTag(tagId),
    onSuccess: () => {
      toast({ title: 'Tag deleted successfully' })
      coreCache.noteTagsChanged(queryClient)
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
