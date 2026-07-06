import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../interfaces/category'

import { getAllCategoriesWithCompetencies } from '../../network/queries/getAllCategoriesWithCompetencies'

export const useGetPeerEvaluationCategoriesWithCompetencies = (enabled: boolean = true) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CategoryWithCompetencies[]>({
    queryKey: ['peerEvaluationCategories', phaseId],
    queryFn: () => getAllCategoriesWithCompetencies(phaseId ?? '', AssessmentType.PEER),
    enabled: enabled,
  })
}
