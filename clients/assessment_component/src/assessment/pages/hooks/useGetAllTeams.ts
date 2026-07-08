import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import { getAllTeams } from '../../network/queries/getAllTeams'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_TEAMS: Team[] = []

export const useGetAllTeams = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Team[]>({
    queryKey: ['teams', phaseId],
    queryFn: () => getAllTeams(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: data ?? EMPTY_TEAMS }
}
