import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import { EvaluationHeader } from '../components/EvaluationHeader'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { useGetMyEvaluationCompletions } from '../hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from '../hooks/useGetMyEvaluations'
import { useGetMyParticipation } from '../hooks/useGetMyParticipation'

import { CategoryEvaluation } from './components/CategoryEvaluation'
import { EvaluationCompletionPage } from './components/EvaluationCompletionPage/EvaluationCompletionPage'

export const SelfEvaluationPage = () => {
  const { courseId } = useParams<{
    courseId: string
  }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: myParticipation } = useGetMyParticipation({ enabled: isStudent })
  const { data: selfEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )
  const { selfEvaluations: evaluations } = useGetMyEvaluations({ enabled: isStudent })
  const { selfEvaluationCompletion: completion } = useGetMyEvaluationCompletions({
    enabled: isStudent,
  })

  return (
    <div className='flex flex-col gap-4'>
      <EvaluationHeader>Self Evaluation</EvaluationHeader>

      <p className='text-sm text-gray-600 dark:text-gray-400'>
        Please fill out the self-evaluation below to reflect on your performance and contributions.
      </p>

      {selfEvaluationCategories.map((category) => (
        <CategoryEvaluation
          key={category.id}
          type={AssessmentType.SELF}
          courseParticipationID={myParticipation?.courseParticipationID ?? ''}
          category={category}
          evaluations={evaluations}
          completed={(completion?.completed ?? false) || !isStudent}
        />
      ))}

      <EvaluationCompletionPage
        type={AssessmentType.SELF}
        deadline={coursePhaseConfig?.selfEvaluationDeadline ?? new Date()}
        courseParticipationID={myParticipation?.courseParticipationID ?? ''}
        authorCourseParticipationID={myParticipation?.courseParticipationID ?? ''}
        completed={(completion?.completed ?? false) || !isStudent}
        completedAt={completion?.completedAt ? new Date(completion.completedAt) : undefined}
      />
    </div>
  )
}
