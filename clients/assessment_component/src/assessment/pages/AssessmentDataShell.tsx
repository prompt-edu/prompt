import { QueryGate } from '@tumaet/prompt-ui-components'
import { AssessmentType } from '../interfaces/assessmentType'
import { useGetAllCategoriesWithCompetencies } from './hooks/useGetAllCategoriesWithCompetencies'
import { useGetAllScoreLevels } from './hooks/useGetAllScoreLevels'
import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetCoursePhaseParticipations } from './hooks/useGetCoursePhaseParticipations'
import { useGetEvaluationCategoriesWithCompetencies } from './hooks/useGetEvaluationCategoriesWithCompetencies'

interface AssessmentDataShellProps {
  children: React.ReactNode
}

export const AssessmentDataShell = ({ children }: AssessmentDataShellProps) => {
  const coursePhaseParticipations = useGetCoursePhaseParticipations()
  const teams = useGetAllTeams()
  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data

  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? false
  const categories = useGetAllCategoriesWithCompetencies({ enabled: assessmentEnabled })
  const scoreLevels = useGetAllScoreLevels({ enabled: assessmentEnabled })

  const selfEvaluationCategories = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )
  const peerEvaluationCategories = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )
  const tutorEvaluationCategories = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
    coursePhaseConfig?.tutorEvaluationEnabled ?? false,
  )

  return (
    <QueryGate
      queries={[
        coursePhaseParticipations,
        teams,
        coursePhaseConfigQuery,
        categories,
        scoreLevels,
        selfEvaluationCategories,
        peerEvaluationCategories,
        tutorEvaluationCategories,
      ]}
    >
      {children}
    </QueryGate>
  )
}
