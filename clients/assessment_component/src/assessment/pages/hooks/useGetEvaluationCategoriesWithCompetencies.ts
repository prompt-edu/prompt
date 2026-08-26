import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { EvaluationType } from '../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../interfaces/category'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_CATEGORIES: CategoryWithCompetencies[] = []

export const useGetEvaluationCategoriesWithCompetencies = (
  assessmentType: EvaluationType,
  enabled = true,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<CategoryWithCompetencies[]>({
    queryKey: assessmentKeys.evaluationCategories(assessmentType, phaseId),
    queryFn: () => assessmentApi.categories.listWithCompetencies(phaseId ?? '', assessmentType),
    enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: enabled ? (data ?? EMPTY_CATEGORIES) : EMPTY_CATEGORIES }
}
