import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { getAllScoreLevels } from '../../network/queries/getAllScoreLevels'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_SCORE_LEVELS: ScoreLevelWithParticipation[] = []

export const useGetAllScoreLevels = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = options?.enabled ?? true

  const { data, ...queryInfo } = useQuery<ScoreLevelWithParticipation[]>({
    queryKey: ['scoreLevels', phaseId],
    queryFn: () => getAllScoreLevels(phaseId ?? ''),
    enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  // A disabled query stays pending forever, which would hang the loading gates above it
  if (!enabled) {
    return { ...queryInfo, data: EMPTY_SCORE_LEVELS, isPending: false, isError: false }
  }

  return { ...queryInfo, data: data ?? EMPTY_SCORE_LEVELS }
}
