import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentKeys } from '../../../network/cache'
import { getMyGradeSuggestion } from '../../../network/queries/getMyGradeSuggestion'

export const useGetMyGradeSuggestion = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<number | undefined>({
    queryKey: assessmentKeys.results.myGradeSuggestion(phaseId),
    queryFn: () => getMyGradeSuggestion(phaseId ?? ''),
    enabled: options?.enabled ?? true,
  })
}
