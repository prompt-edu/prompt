import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import { getAllTeams } from '../../network/queries/getAllTeams'

export const useGetAllTeams = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<Team[]>({
    queryKey: ['teams', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
  })
}
