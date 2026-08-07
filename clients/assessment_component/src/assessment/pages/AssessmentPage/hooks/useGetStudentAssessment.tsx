import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'
import { getStudentAssessment } from '../../../network/queries/getStudentAssessment'

export const useGetStudentAssessment = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { courseParticipationID } = useParams<{ courseParticipationID: string }>()
  const enabled = options?.enabled ?? true

  const query = useQuery<StudentAssessment>({
    queryKey: ['assessments', phaseId, courseParticipationID],
    queryFn: () => getStudentAssessment(phaseId ?? '', courseParticipationID ?? ''),
    enabled,
    // Keep the previous student's data on screen while the next one loads so the
    // page does not collapse to a loader, which would reset the scroll position.
    placeholderData: keepPreviousData,
  })

  // A disabled query stays pending forever, which would hang the loading gates above it.
  // refetch() ignores `enabled`, so it has to be neutralized as well.
  if (!enabled) {
    return {
      ...query,
      data: undefined,
      isPending: false,
      isError: false,
      refetch: async () => query,
    }
  }

  return query
}
