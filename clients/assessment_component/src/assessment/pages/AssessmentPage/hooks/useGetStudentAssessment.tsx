import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { StudentAssessment } from '../../../interfaces/studentAssessment'
import { assessmentApi } from '../../../network/api'
import { assessmentKeys } from '../../../network/cache'

export const useGetStudentAssessment = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { courseParticipationID } = useParams<{ courseParticipationID: string }>()
  const enabled = options?.enabled ?? true

  return useQuery<StudentAssessment>({
    queryKey: assessmentKeys.assessments.ofParticipant(phaseId, courseParticipationID),
    queryFn: () =>
      assessmentApi.assessments.ofParticipant(phaseId ?? '', courseParticipationID ?? ''),
    enabled,
    // Keep the previous student's data on screen while the next one loads so the
    // page does not collapse to a loader, which would reset the scroll position.
    placeholderData: keepPreviousData,
  })
}
