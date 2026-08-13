import { useCourseStore } from '@tumaet/prompt-shared-state'
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'

import { useStudentEvaluationStore } from '../../zustand/useStudentEvaluationStore'
import { EvaluationHeader } from '../components/EvaluationHeader'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { useGetMyEvaluationCompletions } from '../hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from '../hooks/useGetMyEvaluations'
import { useGetMyParticipation } from '../hooks/useGetMyParticipation'

import { CategoryEvaluation } from './components/CategoryEvaluation'
import { EvaluationCompletionPage } from './components/EvaluationCompletionPage/EvaluationCompletionPage'

export const PeerEvaluationPage = () => {
  const { courseId, courseParticipationID } = useParams<{
    courseId: string
    courseParticipationID: string
  }>()
  const { isStudentOfCourse } = useCourseStore()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: myParticipation } = useGetMyParticipation({ enabled: isStudent })
  const { data: peerEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )
  const { peerEvaluations: evaluations } = useGetMyEvaluations({ enabled: isStudent })
  const { peerEvaluationCompletions } = useGetMyEvaluationCompletions({ enabled: isStudent })
  const completion = peerEvaluationCompletions.find(
    (c) => c.courseParticipationID === courseParticipationID,
  )

  const { data: teams } = useGetAllTeams()
  const { setStudentName } = useStudentEvaluationStore()

  const studentName = teams
    .flatMap((team) => team.members)
    .find((participant) => participant.id === courseParticipationID)?.firstName

  useEffect(() => {
    if (studentName) {
      setStudentName(studentName)
    }
  }, [studentName, setStudentName])

  return (
    <div className='flex flex-col gap-4'>
      <EvaluationHeader>Peer Evaluation for {studentName}</EvaluationHeader>

      <p className='text-sm text-gray-600 dark:text-gray-400'>
        Please fill out the Peer evaluation below to assess the performance and contributions of
        your team members.
      </p>

      {peerEvaluationCategories.map((category) => (
        <CategoryEvaluation
          key={category.id}
          type={AssessmentType.PEER}
          courseParticipationID={courseParticipationID ?? ''}
          category={category}
          evaluations={evaluations.filter(
            (evaluation) => evaluation.courseParticipationID === courseParticipationID,
          )}
          completed={(completion?.completed ?? false) || !isStudent}
        />
      ))}

      <EvaluationCompletionPage
        type={AssessmentType.PEER}
        deadline={coursePhaseConfig?.peerEvaluationDeadline ?? new Date()}
        courseParticipationID={courseParticipationID ?? ''}
        authorCourseParticipationID={myParticipation?.courseParticipationID ?? ''}
        completed={(completion?.completed ?? false) || !isStudent}
        completedAt={completion?.completedAt ? new Date(completion.completedAt) : undefined}
      />
    </div>
  )
}
