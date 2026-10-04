import type {
  CreateStudyProgram,
  StudyProgram,
  UpdateStudyProgram,
} from '@core/managementConsole/shared/interfaces/StudyProgram'
import { coreApi } from '@core/network/api'
import { coreCache, coreKeys } from '@core/network/cache'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'

interface UpdateStudyProgramArgs {
  previous: StudyProgram
  studyProgram: UpdateStudyProgram
}

const serverMessage = (error: unknown): string =>
  (isAxiosError<{ error?: string }>(error) && error.response?.data?.error) ||
  'Please try again later'

export const useStudyPrograms = () => {
  return useQuery({
    queryKey: coreKeys.studyPrograms.all(),
    queryFn: coreApi.studyPrograms.list,
  })
}

export const useStudyProgramStudentCounts = () => {
  return useQuery({
    queryKey: coreKeys.studyPrograms.studentCounts(),
    queryFn: coreApi.studyPrograms.studentCounts,
  })
}

export const useCreateStudyProgram = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (studyProgram: CreateStudyProgram) => coreApi.studyPrograms.create(studyProgram),
    onSuccess: () => {
      toast({ title: 'Study program created successfully' })
      coreCache.studyProgramsChanged(queryClient)
    },
    onError: (error: unknown) => {
      toast({
        title: 'Failed to create study program',
        description: serverMessage(error),
        variant: 'destructive',
      })
    },
  })
}

export const useUpdateStudyProgram = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ previous, studyProgram }: UpdateStudyProgramArgs) =>
      coreApi.studyPrograms.update(previous.id, studyProgram),
    onSuccess: (updated, { previous }) => {
      toast({ title: 'Study program updated successfully' })
      if (updated.name === previous.name) {
        coreCache.studyProgramsChanged(queryClient)
      } else {
        coreCache.studyProgramRenamed(queryClient)
      }
    },
    onError: (error: unknown) => {
      toast({
        title: 'Failed to update study program',
        description: serverMessage(error),
        variant: 'destructive',
      })
    },
  })
}

export const useDeleteStudyProgram = () => {
  const { toast } = useToast()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (studyProgramIDs: string[]) => {
      await Promise.all(studyProgramIDs.map((id) => coreApi.studyPrograms.remove(id)))
    },
    onSuccess: (_, studyProgramIDs) => {
      toast({
        title:
          studyProgramIDs.length === 1
            ? 'Study program deleted successfully'
            : `${studyProgramIDs.length} study programs deleted successfully`,
      })
    },
    onError: (error: unknown) => {
      toast({
        title: 'Failed to delete study program',
        description: serverMessage(error),
        variant: 'destructive',
      })
    },
    onSettled: () => {
      coreCache.studyProgramsChanged(queryClient)
    },
  })
}
