import { useState } from 'react'

import { useParticipationStore } from '../../../../../zustand/useParticipationStore'
import { useCoursePhaseConfigStore } from '../../../../../zustand/useCoursePhaseConfigStore'
import { useGetAllAssessmentCompletions } from '../../../../hooks/useGetAllAssessmentCompletions'
import { useReleaseResults } from '../../../hooks/useReleaseResults'

export interface ReleaseAssessmentResultsModel {
  resultsReleased: boolean
  showReleaseDialog: boolean
  setShowReleaseDialog: (open: boolean) => void
  confirmRelease: () => void
  isReleasing: boolean
  releaseError: string | null
  completedAssessments: number
  totalAssessments: number
  allAssessmentsCompleted: boolean
}

export const useReleaseAssessmentResults = (): ReleaseAssessmentResultsModel => {
  const [showReleaseDialog, setShowReleaseDialog] = useState(false)
  const [releaseError, setReleaseError] = useState<string | null>(null)

  const { participations } = useParticipationStore()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { data: assessmentCompletions } = useGetAllAssessmentCompletions()
  const { mutate: releaseResults, isPending: isReleasing } = useReleaseResults()

  const totalAssessments = participations.length
  const completedAssessments =
    assessmentCompletions?.filter((completion) => completion.completed).length ?? 0
  const allAssessmentsCompleted = totalAssessments > 0 && completedAssessments === totalAssessments

  const getErrorMessage = (error: unknown): string => {
    const responseError = (error as { response?: { data?: { error?: string } } })?.response?.data
      ?.error

    return responseError || 'Releasing results failed. Please try again.'
  }

  const confirmRelease = () => {
    setReleaseError(null)
    releaseResults(undefined, {
      onSuccess: () => setShowReleaseDialog(false),
      onError: (error) => setReleaseError(getErrorMessage(error)),
    })
  }

  return {
    resultsReleased: Boolean(coursePhaseConfig?.resultsReleased),
    showReleaseDialog,
    setShowReleaseDialog,
    confirmRelease,
    isReleasing,
    releaseError,
    completedAssessments,
    totalAssessments,
    allAssessmentsCompleted,
  }
}
