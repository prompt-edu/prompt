import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'

import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useEvaluationStore } from '../../zustand/useEvaluationStore'
import { useMyParticipationStore } from '../../zustand/useMyParticipationStore'
import { useStudentEvaluationStore } from '../../zustand/useStudentEvaluationStore'
import { useTeamStore } from '../../zustand/useTeamStore'
import { useTutorEvaluationCategoryStore } from '../../zustand/useTutorEvaluationCategoryStore'

import { CategoryEvaluation } from './components/CategoryEvaluation'
import { EvaluationCompletionPage } from './components/EvaluationCompletionPage/EvaluationCompletionPage'

export const TutorEvaluationPage = () => {
  const { courseId, courseParticipationID } = useParams<{
    courseId: string
    courseParticipationID: string
  }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const { myParticipation } = useMyParticipationStore()
  const { tutorEvaluationCategories } = useTutorEvaluationCategoryStore()
  const { tutorEvaluations: evaluations, tutorEvaluationCompletions } = useEvaluationStore()
  const completion = tutorEvaluationCompletions.find(
    (c) => c.courseParticipationID === courseParticipationID,
  )

  const { teams } = useTeamStore()
  const { setStudentName } = useStudentEvaluationStore()

  const tutor = teams.flatMap((team) => team.tutors).find((t) => t.id === courseParticipationID)

  const tutorName = tutor ? `${tutor.firstName} ${tutor.lastName}` : undefined

  useEffect(() => {
    if (tutorName) {
      setStudentName(tutorName)
    }
  }, [tutorName, setStudentName])

  return (
    <div className='flex flex-col gap-4'>
      <ManagementPageHeader>Evaluation for {tutorName}</ManagementPageHeader>

      <p className='text-sm text-gray-600 dark:text-gray-400'>
        Please fill out the evaluation below to assess the performance and contributions of{' '}
        {tutorName}.
      </p>

      {tutorEvaluationCategories.map((category) => (
        <CategoryEvaluation
          key={category.id}
          type={AssessmentType.TUTOR}
          courseParticipationID={courseParticipationID ?? ''}
          category={category}
          evaluations={evaluations.filter(
            (evaluation) => evaluation.courseParticipationID === courseParticipationID,
          )}
          completed={(completion?.completed ?? false) || !isStudent}
        />
      ))}

      <EvaluationCompletionPage
        type={AssessmentType.TUTOR}
        deadline={coursePhaseConfig?.tutorEvaluationDeadline ?? new Date()}
        courseParticipationID={courseParticipationID ?? ''}
        authorCourseParticipationID={myParticipation?.courseParticipationID ?? ''}
        completed={(completion?.completed ?? false) || !isStudent}
        completedAt={completion?.completedAt ? new Date(completion.completedAt) : undefined}
      />
    </div>
  )
}
