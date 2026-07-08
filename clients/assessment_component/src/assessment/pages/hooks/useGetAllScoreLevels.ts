import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { getAllScoreLevels } from '../../network/queries/getAllScoreLevels'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_SCORE_LEVELS: ScoreLevelWithParticipation[] = []

export const useGetAllScoreLevels = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<ScoreLevelWithParticipation[]>({
    queryKey: ['scoreLevels', phaseId],
    queryFn: () => getAllScoreLevels(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: data ?? EMPTY_SCORE_LEVELS }
}
