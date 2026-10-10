import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'
import { useNavigate, useParams } from 'react-router-dom'
import { AssessmentType } from '../../../../../interfaces/assessmentType'
import type {
  CategoryWithCompetencies,
  UpdateSchemaOrderResponse,
} from '../../../../../interfaces/category'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache, assessmentKeys } from '../../../../../network/cache'
import { toSchemaOrderRequest } from '../../../utils/schemaOrder'

interface UpdateSchemaOrderContext {
  previous: CategoryWithCompetencies[] | undefined
}

const categoriesKeyFor = (assessmentType: AssessmentType, phaseId: string | undefined) =>
  assessmentType === AssessmentType.ASSESSMENT
    ? assessmentKeys.categories(phaseId)
    : assessmentKeys.evaluationCategories(assessmentType, phaseId)

// Shows the new order right away and rolls it back if the server rejects it
export const useUpdateSchemaOrder = (
  assessmentSchemaID: string,
  assessmentType: AssessmentType,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const queryKey = categoriesKeyFor(assessmentType, phaseId)

  return useMutation<
    UpdateSchemaOrderResponse,
    Error,
    CategoryWithCompetencies[],
    UpdateSchemaOrderContext
  >({
    mutationFn: (categories) =>
      assessmentApi.categories.updateOrder(phaseId ?? '', toSchemaOrderRequest(categories)),
    onMutate: async (categories) => {
      await queryClient.cancelQueries({ queryKey })
      const previous = queryClient.getQueryData<CategoryWithCompetencies[]>(queryKey)
      queryClient.setQueryData(queryKey, categories)
      return { previous }
    },
    onSuccess: async (response) => {
      // Reordering a shared schema gave the phase its own copy. Follow it once the config points
      // there, so the page never sees a schema that is no longer configured.
      if (response.assessmentSchemaID !== assessmentSchemaID) {
        await queryClient.refetchQueries({ queryKey: assessmentKeys.coursePhaseConfig(phaseId) })
        assessmentCache.schemaListChanged(queryClient, phaseId)
        navigate(`../${response.assessmentSchemaID}`, { relative: 'path', replace: true })
      }
    },
    onError: (error, _categories, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous)
      }
      const serverError = isAxiosError<{ error?: string }>(error)
        ? error.response?.data?.error
        : undefined
      toast({
        title: 'Could not save the new order',
        description: serverError ?? 'An unexpected error occurred.',
        variant: 'destructive',
      })
    },
    onSettled: () => {
      assessmentCache.schemaOrderChanged(queryClient, phaseId)
    },
  })
}
