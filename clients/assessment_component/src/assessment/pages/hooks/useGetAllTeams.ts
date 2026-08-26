import { useQuery } from '@tanstack/react-query'
import type { Team } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_TEAMS: Team[] = []

export const useGetAllTeams = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Team[]>({
    queryKey: assessmentKeys.teams(phaseId),
    queryFn: () => assessmentApi.config.teams(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: data ?? EMPTY_TEAMS }
}
