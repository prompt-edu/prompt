import { useState } from 'react'

import { useParticipationStore } from '../../../../../zustand/useParticipationStore'
import { useCoursePhaseConfigStore } from '../../../../../zustand/useCoursePhaseConfigStore'
import { useGetAllAssessmentCompletions } from '../../../../hooks/useGetAllAssessmentCompletions'
import { useReleaseResults } from '../../../hooks/useReleaseResults'
import { useUnreleaseResults } from '../../../hooks/useUnreleaseResults'

export interface ReleaseAssessmentResultsModel {
  resultsReleased: boolean
  showReleaseDialog: boolean
  setShowReleaseDialog: (open: boolean) => void
  showUnreleaseDialog: boolean
  setShowUnreleaseDialog: (open: boolean) => void
  confirmRelease: () => void
  confirmUnrelease: () => void
  isReleasing: boolean
  isUnreleasing: boolean
  releaseError: string | null
  unreleaseError: string | null
  completedAssessments: number
  totalAssessments: number
  allAssessmentsCompleted: boolean
}

export const useReleaseAssessmentResults = (): ReleaseAssessmentResultsModel => {
  const [showReleaseDialog, setShowReleaseDialog] = useState(false)
  const [showUnreleaseDialog, setShowUnreleaseDialog] = useState(false)
  const [releaseError, setReleaseError] = useState<string | null>(null)
  const [unreleaseError, setUnreleaseError] = useState<string | null>(null)

  const { participations } = useParticipationStore()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { data: assessmentCompletions } = useGetAllAssessmentCompletions()
  const { mutate: releaseResults, isPending: isReleasing } = useReleaseResults()
  const { mutate: unreleaseResults, isPending: isUnreleasing } = useUnreleaseResults()

  const totalAssessments = participations.length
  const completedAssessments =
    assessmentCompletions?.filter((completion) => completion.completed).length ?? 0
  const allAssessmentsCompleted = totalAssessments > 0 && completedAssessments === totalAssessments

  const getErrorMessage = (error: unknown, fallback: string): string => {
    const responseError = (error as { response?: { data?: { error?: string } } })?.response?.data
      ?.error

    return responseError || fallback
  }

  const confirmRelease = () => {
    setReleaseError(null)
    releaseResults(undefined, {
      onSuccess: () => setShowReleaseDialog(false),
      onError: (error) =>
        setReleaseError(getErrorMessage(error, 'Releasing results failed. Please try again.')),
    })
  }

  const confirmUnrelease = () => {
    setUnreleaseError(null)
    unreleaseResults(undefined, {
      onSuccess: () => setShowUnreleaseDialog(false),
      onError: (error) =>
        setUnreleaseError(getErrorMessage(error, 'Unreleasing results failed. Please try again.')),
    })
  }

  return {
    resultsReleased: Boolean(coursePhaseConfig?.resultsReleased),
    showReleaseDialog,
    setShowReleaseDialog,
    showUnreleaseDialog,
    setShowUnreleaseDialog,
    confirmRelease,
    confirmUnrelease,
    isReleasing,
    isUnreleasing,
    releaseError,
    unreleaseError,
    completedAssessments,
    totalAssessments,
    allAssessmentsCompleted,
  }
}
