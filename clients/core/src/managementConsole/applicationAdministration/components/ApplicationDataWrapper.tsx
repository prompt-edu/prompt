import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { ErrorPage } from '@tumaet/prompt-ui-components'
import { AdditionalScore } from '../interfaces/additionalScore/additionalScore'
import { getAdditionalScoreNames } from '@core/network/queries/additionalScoreNames'
import { useGetApplicationParticipations } from '../hooks/useGetApplicationParticipations'
import { useApplicationStore } from '../zustand/useApplicationStore'
import { useGetCoursePhase } from '@tumaet/prompt-shared-state'

interface ApplicationDataWrapperProps {
  children: React.ReactNode
}

export const ApplicationDataWrapper = ({ children }: ApplicationDataWrapperProps) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { setAdditionalScores, setParticipations, setCoursePhase } = useApplicationStore()

  const {
    data: fetchedAdditionalScores,
    isPending: isAdditionalScoresPending,
    isError: isAdditionalScoresError,
    refetch: refetchScores,
  } = useQuery<AdditionalScore[]>({
    queryKey: ['application_participations', phaseId],
    queryFn: () => getAdditionalScoreNames(phaseId ?? ''),
  })

  const {
    data: fetchedParticipations,
    isPending: isParticipationsPending,
    isError: isParticipantsError,
    refetch: refetchParticipations,
  } = useGetApplicationParticipations()

  const {
    data: fetchedCoursePhase,
    isPending: isCoursePhasePending,
    isError: isCoursePhaseError,
    refetch: refetchCoursePhase,
  } = useGetCoursePhase()

  const isError = isParticipantsError || isAdditionalScoresError || isCoursePhaseError
  const isPending = isParticipationsPending || isAdditionalScoresPending || isCoursePhasePending

  const refetch = () => {
    refetchParticipations()
    refetchScores()
    refetchCoursePhase()
  }

  useEffect(() => {
    if (fetchedAdditionalScores) {
      setAdditionalScores(fetchedAdditionalScores)
    }
  }, [fetchedAdditionalScores, setAdditionalScores])

  useEffect(() => {
    if (fetchedParticipations) {
      setParticipations(fetchedParticipations)
    }
  }, [fetchedParticipations, setParticipations])

  useEffect(() => {
    if (fetchedCoursePhase) {
      setCoursePhase(fetchedCoursePhase)
    }
  }, [fetchedCoursePhase, setCoursePhase])

  return (
    <>
      {isError ? (
        <>
          <ErrorPage onRetry={refetch} />
        </>
      ) : isPending ? (
        <div className='flex justify-center items-center flex-grow'>
          <Loader2 className='h-12 w-12 animate-spin text-primary' />
        </div>
      ) : (
        <>{children}</>
      )}
    </>
  )
}
