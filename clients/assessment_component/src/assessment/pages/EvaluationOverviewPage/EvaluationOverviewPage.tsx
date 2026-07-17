import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { GraduationCap, User, Users } from 'lucide-react'
import { useMemo } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { useGetMyEvaluationCompletions } from '../hooks/useGetMyEvaluationCompletions'
import { useGetMyEvaluations } from '../hooks/useGetMyEvaluations'
import { useGetMyParticipation } from '../hooks/useGetMyParticipation'
import { EvaluationInfoHeader } from './components/EvaluationInfoHeader'
import { EvaluationSection } from './components/EvaluationSection'

export const EvaluationOverviewPage = () => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId } = useParams<{ courseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const path = useLocation().pathname
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { selfEvaluations, peerEvaluations, tutorEvaluations } = useGetMyEvaluations({
    enabled: isStudent,
  })
  const { selfEvaluationCompletion, peerEvaluationCompletions, tutorEvaluationCompletions } =
    useGetMyEvaluationCompletions({ enabled: isStudent })
  const { data: selfEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )
  const { data: peerEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )
  const { data: tutorEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
    coursePhaseConfig?.tutorEvaluationEnabled ?? false,
  )
  const { data: myParticipation } = useGetMyParticipation({ enabled: isStudent })
  const { data: teams } = useGetAllTeams()

  const team = teams.find((t) =>
    t.members.some((member) => member.id === myParticipation?.courseParticipationID || !isStudent),
  )

  const now = new Date()
  const selfEvaluationStarted =
    coursePhaseConfig?.selfEvaluationEnabled &&
    (!coursePhaseConfig?.selfEvaluationStart ||
      now >= new Date(coursePhaseConfig.selfEvaluationStart))
  const peerEvaluationStarted =
    coursePhaseConfig?.peerEvaluationEnabled &&
    (!coursePhaseConfig?.peerEvaluationStart ||
      now >= new Date(coursePhaseConfig.peerEvaluationStart))
  const tutorEvaluationStarted =
    coursePhaseConfig?.tutorEvaluationEnabled &&
    (!coursePhaseConfig?.tutorEvaluationStart ||
      now >= new Date(coursePhaseConfig.tutorEvaluationStart))

  const selfEvaluationCompetencyCount = useMemo(
    () =>
      selfEvaluationCategories.reduce((count, category) => count + category.competencies.length, 0),
    [selfEvaluationCategories],
  )

  const peerEvaluationCompetencyCount = useMemo(
    () =>
      peerEvaluationCategories.reduce((count, category) => count + category.competencies.length, 0),
    [peerEvaluationCategories],
  )

  const tutorEvaluationCompetencyCount = useMemo(
    () =>
      tutorEvaluationCategories.reduce(
        (count, category) => count + category.competencies.length,
        0,
      ),
    [tutorEvaluationCategories],
  )

  const isSelfEvaluationCompleted = selfEvaluationCompletion?.completed ?? false

  const teamMembers =
    team?.members.filter((member) => member.id !== myParticipation?.courseParticipationID) ?? []
  const teamTutors =
    team?.tutors.filter((tutor) => tutor.id !== myParticipation?.courseParticipationID) ?? []

  const completedPeerEvaluations = teamMembers.filter(
    (member) =>
      peerEvaluationCompletions.find((c) => c.courseParticipationID === member.id)?.completed,
  ).length

  const isPeerEvaluationCompleted = completedPeerEvaluations === teamMembers.length

  const completedTutorEvaluations = teamTutors.filter(
    (tutor) =>
      tutorEvaluationCompletions.find((c) => c.courseParticipationID === tutor.id)?.completed,
  ).length

  const isTutorEvaluationCompleted = completedTutorEvaluations === teamTutors.length

  const allEvaluationsCompleted =
    isSelfEvaluationCompleted && isPeerEvaluationCompleted && isTutorEvaluationCompleted

  return (
    <div className='px-4 py-6'>
      <ManagementPageHeader>Assessment Results & Evaluation</ManagementPageHeader>

      <div className='space-y-4'>
        <EvaluationInfoHeader
          allEvaluationsCompleted={allEvaluationsCompleted}
          resultsLink={`${path}/results`}
        />

        {selfEvaluationStarted && (
          <EvaluationSection
            title='Self Evaluation'
            icon={<User className='h-5 w-5 text-blue-600 dark:text-blue-400' />}
            assessmentType={AssessmentType.SELF}
            deadline={coursePhaseConfig?.selfEvaluationDeadline}
            competencyCount={selfEvaluationCompetencyCount}
            targets={[
              {
                id: 'self',
                name: 'Evaluate yourself',
                navigationPath: `${path}/self-evaluation`,
                completed: isSelfEvaluationCompleted,
                evaluationCount: selfEvaluations.length,
              },
            ]}
          />
        )}

        {peerEvaluationStarted && team && (
          <EvaluationSection
            title='Peer Evaluation'
            icon={<Users className='h-5 w-5 text-green-600 dark:text-green-400' />}
            assessmentType={AssessmentType.PEER}
            deadline={coursePhaseConfig?.peerEvaluationDeadline}
            teamName={team.name}
            competencyCount={peerEvaluationCompetencyCount}
            targets={teamMembers.map((member) => {
              const memberId = member.id ?? `${member.firstName}-${member.lastName}`
              return {
                id: memberId,
                name: `${member.firstName} ${member.lastName}`,
                navigationPath: `${path}/peer-evaluation/${memberId}`,
                completed:
                  peerEvaluationCompletions.find((c) => c.courseParticipationID === member.id)
                    ?.completed ?? false,
                evaluationCount: peerEvaluations.filter(
                  (evaluation) => evaluation.courseParticipationID === member.id,
                ).length,
              }
            })}
          />
        )}

        {tutorEvaluationStarted && team && (
          <EvaluationSection
            title='Tutor Evaluation'
            icon={<GraduationCap className='h-5 w-5 text-purple-600 dark:text-purple-400' />}
            assessmentType={AssessmentType.TUTOR}
            deadline={coursePhaseConfig?.tutorEvaluationDeadline}
            teamName={team.name}
            competencyCount={tutorEvaluationCompetencyCount}
            targets={teamTutors.map((tutor) => {
              const tutorId = tutor.id ?? `${tutor.firstName}-${tutor.lastName}`
              return {
                id: tutorId,
                name: `${tutor.firstName} ${tutor.lastName}`,
                navigationPath: `${path}/tutor-evaluation/${tutorId}`,
                completed:
                  tutorEvaluationCompletions.find((c) => c.courseParticipationID === tutor.id)
                    ?.completed ?? false,
                evaluationCount: tutorEvaluations.filter(
                  (evaluation) => evaluation.courseParticipationID === tutor.id,
                ).length,
              }
            })}
          />
        )}
      </div>
    </div>
  )
}
