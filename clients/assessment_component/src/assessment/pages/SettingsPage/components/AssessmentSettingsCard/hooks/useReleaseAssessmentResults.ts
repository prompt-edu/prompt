import { useState } from 'react'
import { useGetAllAssessmentCompletions } from '../../../../hooks/useGetAllAssessmentCompletions'
import { useGetCoursePhaseConfig } from '../../../../hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from '../../../../hooks/useGetCoursePhaseParticipations'
import { useReleaseResults } from '../../../hooks/useReleaseResults'
import { useUnreleaseResults } from '../../../hooks/useUnreleaseResults'

export interface ReleaseAssessmentResultsModel {
  assessmentEnabled: boolean
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

  const { data: participations } = useGetCoursePhaseParticipations()
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  // Default to enabled while the config loads so an existing phase is never briefly
  // treated as evaluation-only
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? true
  const { data: assessmentCompletions } = useGetAllAssessmentCompletions({
    enabled: assessmentEnabled,
  })
  const { mutate: releaseResults, isPending: isReleasing } = useReleaseResults()
  const { mutate: unreleaseResults, isPending: isUnreleasing } = useUnreleaseResults()

  const totalAssessments = participations.length
  const completedAssessments =
    assessmentCompletions?.filter((completion) => completion.completed).length ?? 0
  // Evaluation-only phases have no assessments to finalize, so nothing gates the release
  const allAssessmentsCompleted =
    !assessmentEnabled || (totalAssessments > 0 && completedAssessments === totalAssessments)

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
    assessmentEnabled,
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
