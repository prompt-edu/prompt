import { getApplicationParticipations } from '@core/network/queries/applicationParticipations'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { ApplicationParticipation } from '../interfaces/applicationParticipation'

export const useGetApplicationParticipations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<ApplicationParticipation[]>({
    queryKey: ['application_participations', 'students', phaseId],
    queryFn: () => getApplicationParticipations(phaseId ?? ''),
  })
}
