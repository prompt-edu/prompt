import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import { getAllScoreLevels } from '../../network/queries/getAllScoreLevels'

export const useGetAllScoreLevels = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<ScoreLevelWithParticipation[]>({
    queryKey: ['scoreLevels', phaseId],
    queryFn: () => getAllScoreLevels(phaseId ?? ''),
  })
}
