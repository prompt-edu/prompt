import {
  Button,
  Card,
  CardContent,
  ErrorPage,
  getStudentName,
  ManagementPageHeader,
  ParticipantNavigation,
  QueryGate,
  StickyHeader,
} from '@tumaet/prompt-ui-components'
import { Printer } from 'lucide-react'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import { FeedbackItemDisplayPanel } from '../components/FeedbackItemDisplayPanel/FeedbackItemDisplayPanel'
import { PrintReport } from '../components/PrintReport/PrintReport'
import { useGetAllTeams } from '../hooks/useGetAllTeams'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { getTutorLabel } from '../hooks/useTutorLabel'
import { getTeamMemberName } from '../utils/getTeamMemberName'
import { printPage } from '../utils/printPage'
import { CategoryEvaluation } from './components/CategoryEvaluation'
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

  const tutor = useMemo(() => {
    for (const team of teams) {
      const foundTutor = team.tutors.find((t) => t.id === tutorId)
      if (foundTutor) {
        return {
          id: foundTutor.id,
          firstName: foundTutor.firstName,
          lastName: foundTutor.lastName,
          teamName: team.name,
          teamId: team.id,
        }
      }
    }
    return null
  }, [teams, tutorId])

  const positiveFeedbackItems = useMemo(() => {
    return feedbackItems.filter((item) => item.feedbackType === 'positive')
  }, [feedbackItems])

  const negativeFeedbackItems = useMemo(() => {
    return feedbackItems.filter((item) => item.feedbackType === 'negative')
  }, [feedbackItems])

  const reportScores = useMemo(
    () =>
      tutorEvaluations.map((evaluation) => ({
        ...evaluation,
        authorName: getTeamMemberName(teams, evaluation.authorCourseParticipationID),
      })),
    [teams, tutorEvaluations],
  )

  const evaluatorCount = useMemo(
    () =>
      new Set(tutorEvaluations.map((evaluation) => evaluation.authorCourseParticipationID)).size,
    [tutorEvaluations],
  )

  return (
    <QueryGate queries={[evaluationsQuery, feedbackItemsQuery]}>
      {() => {
        if (!tutor) {
          return <ErrorPage message={`The requested ${tutorLabel.text} could not be found.`} />
        }

        const pageTitle = `${tutorLabel.title} Evaluation Results for ${getStudentName(tutor)}`

        return (
          <>
            <div className='space-y-4 print:hidden'>
              <StickyHeader
                floating
                expandedContent={<ManagementPageHeader>{pageTitle}</ManagementPageHeader>}
              >
                {(docked) => (
                  <ParticipantNavigation
                    participants={orderedTutors}
                    currentId={tutorId}
                    onNavigate={navigateToTutor}
                    compact={docked}
                  >
                    <span className='truncate font-semibold'>{pageTitle}</span>
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
                  <Button variant='outline' onClick={printPage} className='gap-2'>
                    <Printer className='h-4 w-4' />
                    PDF / Print
                  </Button>
                </div>
              )}
            </div>

            <PrintReport
              title={pageTitle}
              subtitle={tutor.teamName}
              meta={
                evaluatorCount > 1 ? (
                  <span>
                    <strong>Evaluators:</strong> {evaluatorCount}
                  </span>
                ) : undefined
              }
              categories={tutorEvaluationCategories}
              scores={reportScores}
              feedbackItems={feedbackItems}
            />
          </>
        )
      }}
    </QueryGate>
  )
}
