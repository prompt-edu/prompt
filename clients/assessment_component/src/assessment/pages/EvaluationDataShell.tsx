import { useCourseStore } from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  ErrorPage,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../interfaces/assessmentType'
import { useGetAllTeams } from './hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from './hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from './hooks/useGetEvaluationCategoriesWithCompetencies'
import { useGetMyEvaluationCompletions } from './hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from './hooks/useGetMyEvaluations'
import { useGetMyParticipation } from './hooks/useGetMyParticipation'

interface EvaluationDataShellProps {
  children: React.ReactNode
}

export const EvaluationDataShell = ({ children }: EvaluationDataShellProps) => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId } = useParams<{ courseId: string; phaseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data

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

  const teams = useGetAllTeams()
  const myParticipation = useGetMyParticipation({ enabled: isStudent })
  const myEvaluationCompletions = useGetMyEvaluationCompletions({ enabled: isStudent })
  const myEvaluations = useGetMyEvaluations({ enabled: isStudent })

  return (
    <QueryGate
      queries={[
        coursePhaseConfigQuery,
        selfEvaluationCategories,
        peerEvaluationCategories,
        tutorEvaluationCategories,
        teams,
        myParticipation,
        myEvaluationCompletions,
        myEvaluations,
      ]}
      errorFallback={({ refetch }) => (
        <ErrorPage
          onRetry={refetch}
          description='Could not fetch self, peer, or tutor evaluation categories'
        />
      )}
    >
      {!isStudent && (
        <Alert>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>Your are not a student of this course.</AlertTitle>
          <AlertDescription>
            The following components are disabled because you are not a student of this course.
            Evaluations for self and peer assessments are currently only available for students. The
            platform will show a random team regardless, to demonstrate the functionality.
          </AlertDescription>
        </Alert>
      )}
      {children}
    </QueryGate>
  )
}
