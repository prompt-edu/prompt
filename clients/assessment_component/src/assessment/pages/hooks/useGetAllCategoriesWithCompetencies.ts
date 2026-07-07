import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../interfaces/category'

import { getAllCategoriesWithCompetencies } from '../../network/queries/getAllCategoriesWithCompetencies'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_CATEGORIES: CategoryWithCompetencies[] = []

export const useGetAllCategoriesWithCompetencies = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = options?.enabled ?? true

  const { data, ...queryInfo } = useQuery<CategoryWithCompetencies[]>({
    queryKey: ['categories', phaseId],
    queryFn: () => getAllCategoriesWithCompetencies(phaseId ?? '', AssessmentType.ASSESSMENT),
    enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: enabled ? (data ?? EMPTY_CATEGORIES) : EMPTY_CATEGORIES }
}
