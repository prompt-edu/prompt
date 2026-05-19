import { useMemo } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { Users, User, GraduationCap } from 'lucide-react'

import { useCourseStore } from '@tumaet/prompt-shared-state'
import { ManagementPageHeader, Card } from '@tumaet/prompt-ui-components'

import { useEvaluationStore } from '../../zustand/useEvaluationStore'
import { useSelfEvaluationCategoryStore } from '../../zustand/useSelfEvaluationCategoryStore'
import { usePeerEvaluationCategoryStore } from '../../zustand/usePeerEvaluationCategoryStore'
import { useTutorEvaluationCategoryStore } from '../../zustand/useTutorEvaluationCategoryStore'
import { useTeamStore } from '../../zustand/useTeamStore'
import { useMyParticipationStore } from '../../zustand/useMyParticipationStore'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'

import { EvaluationInfoCard } from './components/EvaluationInfoCard'
import { EvaluationInfoHeader } from './components/EvaluationInfoHeader'
import { SelfEvaluationStatusCard } from './components/SelfEvaluationStatusCard'
import { PeerEvaluationStatusCard } from './components/PeerEvaluationStatusCard'
import { TutorEvaluationStatusCard } from './components/TutorEvaluationStatusCard'
import { TeamBadge } from '../components/badges'

export const EvaluationOverviewPage = () => {
  const { isStudentOfCourse } = useCourseStore()
  const { courseId } = useParams<{ courseId: string }>()
  const isStudent = isStudentOfCourse(courseId ?? '')

  const path = useLocation().pathname
  const {
    selfEvaluations,
    peerEvaluations,
    tutorEvaluations,
    selfEvaluationCompletion,
    peerEvaluationCompletions,
    tutorEvaluationCompletions,
  } = useEvaluationStore()
  const { selfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { peerEvaluationCategories } = usePeerEvaluationCategoryStore()
  const { tutorEvaluationCategories } = useTutorEvaluationCategoryStore()
  const { myParticipation } = useMyParticipationStore()
  const { teams } = useTeamStore()
  const { coursePhaseConfig } = useCoursePhaseConfigStore()

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

  const completedPeerEvaluations =
    team?.members
      .filter((member) => member.id !== myParticipation?.courseParticipationID)
      .filter(
        (member) =>
          peerEvaluationCompletions.find((c) => c.courseParticipationID === member.id)?.completed,
      ).length ?? 0

  const totalPeerEvaluations =
    team?.members.filter((member) => member.id !== myParticipation?.courseParticipationID).length ??
    0

  const isPeerEvaluationCompleted = completedPeerEvaluations === totalPeerEvaluations

  const completedTutorEvaluations =
    team?.tutors.filter(
      (tutor) =>
        tutorEvaluationCompletions.find((c) => c.courseParticipationID === tutor.id)?.completed,
    ).length ?? 0

  const totalTutorEvaluations = team?.tutors.length ?? 0

  const isTutorEvaluationCompleted = completedTutorEvaluations === totalTutorEvaluations

  const allEvaluationsCompleted =
    isSelfEvaluationCompleted && isPeerEvaluationCompleted && isTutorEvaluationCompleted

  return (
    <div className=''>
      <div className='mx-auto px-4 py-6'>
        <ManagementPageHeader>Assessment Results & Evaluation</ManagementPageHeader>

        <EvaluationInfoHeader
          allEvaluationsCompleted={allEvaluationsCompleted}
          resultsLink={`${path}/results`}
        />

        <div>
          <div className='grid grid-cols-1 xl:grid-cols-3 gap-6 mb-8'>
            {selfEvaluationStarted && (
              <SelfEvaluationStatusCard isCompleted={isSelfEvaluationCompleted} />
            )}

            {peerEvaluationStarted && team && (
              <PeerEvaluationStatusCard
                completedEvaluations={completedPeerEvaluations}
                totalEvaluations={totalPeerEvaluations}
                isCompleted={isPeerEvaluationCompleted}
              />
            )}

            {tutorEvaluationStarted && team && (
              <TutorEvaluationStatusCard
                completedEvaluations={completedTutorEvaluations}
                totalEvaluations={totalTutorEvaluations}
                isCompleted={isTutorEvaluationCompleted}
              />
            )}
          </div>

          {selfEvaluationStarted && (
            <div className='mb-8'>
              <div className='flex items-center gap-3 mb-6'>
                <div className='flex items-center gap-2'>
                  <User className='h-6 w-6 text-blue-600 dark:text-blue-400' />
                  <h1 className='text-2xl font-bold text-gray-900 dark:text-gray-100'>
                    Self Evaluation
                  </h1>
                </div>
              </div>
              <Card className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'>
                <EvaluationInfoCard
                  name='Self Evaluation'
                  navigationPath={`${path}/self-evaluation`}
                  competencyCount={selfEvaluationCompetencyCount}
                  completed={isSelfEvaluationCompleted}
                  evaluations={selfEvaluations}
                />
              </Card>
            </div>
          )}

          {peerEvaluationStarted && team && (
            <div className='mb-8'>
              <div className='flex items-center gap-3 mb-6'>
                <div className='flex items-center gap-2'>
                  <Users className='h-6 w-6 text-green-600 dark:text-green-400' />
                  <h1 className='text-2xl font-bold text-gray-900 dark:text-gray-100'>
                    Peer Evaluation
                  </h1>
                </div>
                <TeamBadge teamName={team.name} />
              </div>
              <div className='space-y-4'>
                {team.members
                  .filter((member) => member.id !== myParticipation?.courseParticipationID)
                  .map((member) => (
                    <Card
                      key={member.id}
                      className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'
                    >
                      <EvaluationInfoCard
                        name={member.firstName + ' ' + member.lastName}
                        navigationPath={`${path}/peer-evaluation/${member?.id}`}
                        competencyCount={peerEvaluationCompetencyCount}
                        completed={
                          peerEvaluationCompletions.find(
                            (c) => c.courseParticipationID === member.id,
                          )?.completed ?? false
                        }
                        evaluations={peerEvaluations.filter(
                          (evaluation) => evaluation.courseParticipationID === member.id,
                        )}
                      />
                    </Card>
                  ))}
              </div>
            </div>
          )}

          {tutorEvaluationStarted && team && (
            <div className='mb-8'>
              <div className='flex items-center gap-3 mb-6'>
                <div className='flex items-center gap-2'>
                  <GraduationCap className='h-6 w-6 text-purple-600 dark:text-purple-400' />
                  <h1 className='text-2xl font-bold text-gray-900 dark:text-gray-100'>
                    Tutor Evaluation
                  </h1>
                </div>
                <TeamBadge teamName={team.name} />
              </div>
              <div className='space-y-4'>
                {team.tutors
                  .filter((tutor) => tutor.id !== myParticipation?.courseParticipationID)
                  .map((tutor) => (
                    <Card
                      key={tutor.id}
                      className='border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-xs'
                    >
                      <EvaluationInfoCard
                        name={tutor.firstName + ' ' + tutor.lastName}
                        navigationPath={`${path}/tutor-evaluation/${tutor?.id}`}
                        competencyCount={tutorEvaluationCompetencyCount}
                        completed={
                          tutorEvaluationCompletions.find(
                            (c) => c.courseParticipationID === tutor.id,
                          )?.completed ?? false
                        }
                        evaluations={tutorEvaluations.filter(
                          (evaluation) => evaluation.courseParticipationID === tutor.id,
                        )}
                      />
                    </Card>
                  ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
