import { useQuery } from '@tanstack/react-query'
import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { getResolvedCoursePhaseParticipations } from '../../network/getResolvedCoursePhaseParticipations'

export const MatchingParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: participations,
    isPending,
    isError,
    refetch,
  } = useQuery<CoursePhaseParticipationWithStudent[]>({
    queryKey: ['participants', phaseId],
    queryFn: () => getResolvedCoursePhaseParticipations(phaseId ?? ''),
    enabled: !!phaseId,
  })

  // The interview score is resolved from the interview service into prevData.score.
  const interviewScoreColumn = useMemo<ExtraParticipantColumn<number | null>>(
    () => ({
      id: 'interviewScore',
      header: 'Interview Score',
      accessorFn: (row) => (row.prevData?.score as number | undefined) ?? null,
      cell: (info) => (info.getValue() as number | null) ?? 'N/A',
      enableSorting: true,
      extraData: (participations ?? []).map((participation) => {
        const score = participation.prevData?.score as number | undefined
        return {
          courseParticipationID: participation.courseParticipationID,
          value: score ?? null,
          stringValue: score !== undefined ? String(score) : '',
        }
      }),
    }),
    [participations],
  )

  return (
    <div>
      <ManagementPageHeader>Matching Participants</ManagementPageHeader>
      {isError ? (
        <ErrorPage onRetry={refetch} />
      ) : isPending ? (
        <div className='flex justify-center items-center h-64'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <CoursePhaseParticipationsTable
          phaseId={phaseId!}
          participants={participations ?? []}
          extraColumns={[interviewScoreColumn]}
          exportDeps={{ prevDataKeys: ['score'] }}
        />
      )}
    </div>
  )
}
