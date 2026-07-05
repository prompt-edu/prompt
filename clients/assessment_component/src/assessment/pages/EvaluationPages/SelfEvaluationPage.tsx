import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'

import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useEvaluationStore } from '../../zustand/useEvaluationStore'
import { useMyParticipationStore } from '../../zustand/useMyParticipationStore'
import { useSelfEvaluationCategoryStore } from '../../zustand/useSelfEvaluationCategoryStore'

import { CategoryEvaluation } from './components/CategoryEvaluation'
import { EvaluationCompletionPage } from './components/EvaluationCompletionPage/EvaluationCompletionPage'

export const SelfEvaluationPage = () => {
  const { courseId } = useParams<{
    courseId: string
  }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { myParticipation } = useMyParticipationStore()
  const { selfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { selfEvaluations: evaluations, selfEvaluationCompletion: completion } =
    useEvaluationStore()

  return (
    <div className='flex flex-col gap-4'>
      <ManagementPageHeader>Self Evaluation</ManagementPageHeader>

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
