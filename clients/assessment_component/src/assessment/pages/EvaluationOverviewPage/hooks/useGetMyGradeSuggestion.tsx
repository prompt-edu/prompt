import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetMyGradeSuggestion = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<number | undefined>({
    queryKey: assessmentKeys.results.myGradeSuggestion(phaseId),
    queryFn: () => assessmentApi.completions.myGradeSuggestion(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
