import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { EvaluationReminderStatus } from '../../interfaces/evaluationReminder'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

export const useGetEvaluationReminderStatus = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<EvaluationReminderStatus>({
    queryKey: assessmentKeys.evaluationReminderStatus(phaseId),
    queryFn: () => assessmentApi.config.reminderStatus(phaseId ?? ''),
    enabled: !!phaseId,
  })
}
