import { useQuery } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { Loader2, TriangleAlert } from 'lucide-react'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { getResolvedCoursePhaseParticipations } from '../../network/getResolvedCoursePhaseParticipations'
import type { ResolvedParticipations } from '../../network/resolveParticipations'

export const MatchingParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: resolvedParticipations,
    isPending,
    isError,
    refetch,
  } = useQuery<ResolvedParticipations>({
    queryKey: ['participants', phaseId],
    queryFn: () => getResolvedCoursePhaseParticipations(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const participations = resolvedParticipations?.participations
  const failedResolutions = resolvedParticipations?.failedResolutions ?? []

  // The interview score is resolved from the interview service into prevData.score.
  const interviewScoreColumn = useMemo<ExtraParticipantColumn<number | null>>(
    () => ({
      id: 'interviewScore',
      header: 'Interview Score',
      accessorFn: (row) => (row.prevData?.score as number | undefined) ?? null,
      cell: (info) => (info.getValue() as number | null) ?? 'N/A',
      enableSorting: true,
      extraData: (participations ?? []).map((participation) => {
        const score = participation.prevData?.score as number | null | undefined
        return {
          courseParticipationID: participation.courseParticipationID,
          value: score ?? null,
          stringValue: score !== undefined && score !== null ? String(score) : '',
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
        <>
          {failedResolutions.length > 0 && (
            <Alert variant='destructive' className='mb-4'>
              <TriangleAlert className='h-4 w-4' />
              <AlertTitle>Predecessor data could not be loaded</AlertTitle>
              <AlertDescription>
                {failedResolutions.join(', ')} could not be fetched from the providing phase. Empty
                cells below mean "not loaded", not "no score", so do not export this table until it
                loads.
              </AlertDescription>
            </Alert>
          )}
          <CoursePhaseParticipationsTable
            phaseId={phaseId!}
            participants={participations ?? []}
            extraColumns={[interviewScoreColumn]}
            exportDeps={{ prevDataKeys: ['score'] }}
          />
        </>
      )}
    </div>
  )
}
