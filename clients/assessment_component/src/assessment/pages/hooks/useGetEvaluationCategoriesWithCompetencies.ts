import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../interfaces/category'
import { getAllCategoriesWithCompetencies } from '../../network/queries/getAllCategoriesWithCompetencies'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export type EvaluationType = AssessmentType.SELF | AssessmentType.PEER | AssessmentType.TUTOR

const EMPTY_CATEGORIES: CategoryWithCompetencies[] = []

const QUERY_KEY_PREFIX: Record<EvaluationType, string> = {
  [AssessmentType.SELF]: 'selfEvaluationCategories',
  [AssessmentType.PEER]: 'peerEvaluationCategories',
  [AssessmentType.TUTOR]: 'tutorEvaluationCategories',
}

export const useGetEvaluationCategoriesWithCompetencies = (
  assessmentType: EvaluationType,
  enabled = true,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<CategoryWithCompetencies[]>({
    queryKey: [QUERY_KEY_PREFIX[assessmentType], phaseId],
    queryFn: () => getAllCategoriesWithCompetencies(phaseId ?? '', assessmentType),
    enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: enabled ? (data ?? EMPTY_CATEGORIES) : EMPTY_CATEGORIES }
}
