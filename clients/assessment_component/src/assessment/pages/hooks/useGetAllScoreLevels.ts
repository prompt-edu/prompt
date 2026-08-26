import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { assessmentKeys } from '../../network/cache'
import { getAllScoreLevels } from '../../network/queries/getAllScoreLevels'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_SCORE_LEVELS: ScoreLevelWithParticipation[] = []

export const useGetAllScoreLevels = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = options?.enabled ?? true

  const { data, ...queryInfo } = useQuery<ScoreLevelWithParticipation[]>({
    queryKey: assessmentKeys.scoreLevels(phaseId),
    queryFn: () => getAllScoreLevels(phaseId ?? ''),
    enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: enabled ? (data ?? EMPTY_SCORE_LEVELS) : EMPTY_SCORE_LEVELS }
}
