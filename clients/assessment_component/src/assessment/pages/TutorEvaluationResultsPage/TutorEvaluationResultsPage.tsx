import {
  Badge,
  Card,
  CardContent,
  ErrorPage,
  ParticipantNavigation,
  QueryGate,
  StickyHeader,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { DownloadPdfButton } from '../components/DownloadPdfButton'
import { ScoreBadges } from '../components/EvaluationScore'
import { FeedbackItemDisplayPanel } from '../components/FeedbackItemDisplayPanel/FeedbackItemDisplayPanel'
import { ParticipantProfile } from '../components/ParticipantProfile'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { getTutorLabel } from '../hooks/useTutorLabel'
import { personFileBaseName } from '../utils/bulkExport/zipFiles'
import { getScoreLevelsFromEvaluations } from '../utils/getScoreLevelsFromEvaluations'
import { getTeamNamesOfTutor } from '../utils/getTutorRows'
import { CategoryEvaluation } from './components/CategoryEvaluation'
import { TutorEvaluationReport } from './components/TutorEvaluationReport'
import { useGetEvaluationsForTutorInPhase } from './hooks/useGetEvaluationsForTutorInPhase'
import { useGetFeedbackItemsForTutorInPhase } from './hooks/useGetFeedbackItemsForTutorInPhase'
import { useTutorNavigation } from './hooks/useTutorNavigation'

export const TutorEvaluationResultsPage = () => {
  const { tutorId } = useParams<{ tutorId: string }>()
  const { orderedTutors, navigateToTutor } = useTutorNavigation()

  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const tutorLabel = getTutorLabel(coursePhaseConfig)
  const { data: teams } = useGetAllTeams()
  const { data: tutorEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
    coursePhaseConfig?.tutorEvaluationEnabled ?? false,
  )

  const evaluationsQuery = useGetEvaluationsForTutorInPhase(tutorId ?? '', {
    enabled: !!tutorId,
  })
  const feedbackItemsQuery = useGetFeedbackItemsForTutorInPhase(tutorId ?? '', {
    enabled: !!tutorId,
  })

  const tutorEvaluations = evaluationsQuery.data ?? []
  const feedbackItems = feedbackItemsQuery.data ?? []

  // A tutor who leads several teams is listed under each of them.
  const tutor = useMemo(() => {
    const foundTutor = teams
      .flatMap((team) => team.tutors)
      .find((teamTutor) => teamTutor.id === tutorId)
    if (!foundTutor) return null
    return {
      id: foundTutor.id,
      firstName: foundTutor.firstName,
      lastName: foundTutor.lastName,
      teamNames: getTeamNamesOfTutor(teams, tutorId ?? ''),
    }
  }, [teams, tutorId])

  const positiveFeedbackItems = useMemo(() => {
    return feedbackItems.filter((item) => item.feedbackType === 'positive')
  }, [feedbackItems])

  const negativeFeedbackItems = useMemo(() => {
    return feedbackItems.filter((item) => item.feedbackType === 'negative')
  }, [feedbackItems])

  return (
    <QueryGate queries={[evaluationsQuery, feedbackItemsQuery]}>
      {() => {
        if (!tutor) {
          return <ErrorPage message={`The requested ${tutorLabel.text} could not be found.`} />
        }

        const badges = (
          <>
            {tutor.teamNames.map((teamName) => (
              <Badge key={teamName} variant='outline'>
                {teamName}
              </Badge>
            ))}
            <ScoreBadges
              score={getScoreLevelsFromEvaluations(tutorEvaluations, AssessmentType.TUTOR)[0]}
            />
          </>
        )

        return (
          <>
            <div className='space-y-4 print:hidden'>
              <StickyHeader
                floating
                expandedContent={
                  <ParticipantProfile
                    courseParticipationId={tutor.id}
                    person={tutor}
                    label={`${tutorLabel.title} Evaluation`}
                  >
                    {badges}
                  </ParticipantProfile>
                }
              >
                {(docked) => (
                  <ParticipantNavigation
                    participants={orderedTutors}
                    currentId={tutorId}
                    onNavigate={navigateToTutor}
                    compact={docked}
                  >
                    {badges}
                  </ParticipantNavigation>
                )}
              </StickyHeader>

              {tutorEvaluationCategories.length === 0 ? (
                <Card>
                  <CardContent className='p-6'>
                    <p className='text-center text-muted-foreground'>
                      No evaluation categories configured yet.
                    </p>
                  </CardContent>
                </Card>
              ) : (
                <div className='space-y-6'>
                  <div className='space-y-4'>
                    {tutorEvaluationCategories.map((category) => {
                      return (
                        <CategoryEvaluation
                          key={category.id}
                          category={category}
                          evaluations={tutorEvaluations.filter((evaluation) =>
                            category.competencies
                              .map((competency) => competency.id)
                              .includes(evaluation.competencyID),
                          )}
                        />
                      )
                    })}
                  </div>

                  <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
                    <FeedbackItemDisplayPanel
                      feedbackItems={negativeFeedbackItems}
                      feedbackType='negative'
                      studentName={tutor.firstName}
                    />
                    <FeedbackItemDisplayPanel
                      feedbackItems={positiveFeedbackItems}
                      feedbackType='positive'
                      studentName={tutor.firstName}
                    />
                  </div>
                </div>
              )}

              {tutorEvaluationCategories.length > 0 && (
                <div className='flex justify-end pt-4'>
                  <DownloadPdfButton
                    fileName={`tutor_evaluation_${personFileBaseName(tutor)}.pdf`}
                  />
                </div>
              )}
            </div>

            <TutorEvaluationReport
              tutor={tutor}
              subtitle={tutor.teamNames.join(', ')}
              tutorTitle={tutorLabel.title}
              categories={tutorEvaluationCategories}
              evaluations={tutorEvaluations}
              feedbackItems={feedbackItems}
              teams={teams}
            />
          </>
        )
      }}
    </QueryGate>
  )
}
