import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import { type ReactNode, useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import { getFeedbackItemsForStudent } from '../../network/queries/getFeedbackItemsForStudent'
import { getPeerEvaluationsForParticipantInPhase } from '../../network/queries/getPeerEvaluationsForParticipantInPhase'
import { getSelfEvaluationsForParticipantInPhase } from '../../network/queries/getSelfEvaluationsForParticipantInPhase'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import { usePeerEvaluationCategoryStore } from '../../zustand/usePeerEvaluationCategoryStore'
import { useSelfEvaluationCategoryStore } from '../../zustand/useSelfEvaluationCategoryStore'

import { FeedbackItemDisplayPanel } from '../components/FeedbackItemDisplayPanel/FeedbackItemDisplayPanel'
import { CategoryEvaluation } from '../TutorEvaluationResultsPage/components/CategoryEvaluation'

interface EvaluationParticipantResultsPageProps {
  assessmentType: AssessmentType.SELF | AssessmentType.PEER
}

export const EvaluationParticipantResultsPage = ({
  assessmentType,
}: EvaluationParticipantResultsPageProps): ReactNode => {
  const { phaseId, courseParticipationID } = useParams<{
    phaseId: string
    courseParticipationID: string
  }>()

  const { participations } = useParticipationStore()
  const { selfEvaluationCategories } = useSelfEvaluationCategoryStore()
  const { peerEvaluationCategories } = usePeerEvaluationCategoryStore()

  const {
    data: evaluations = [],
    isPending: isEvaluationsPending,
    isError: isEvaluationsError,
    refetch: refetchEvaluations,
  } = useQuery({
    queryKey: [assessmentType, 'evaluations', phaseId, courseParticipationID],
    queryFn: () =>
      assessmentType === AssessmentType.SELF
        ? getSelfEvaluationsForParticipantInPhase(phaseId ?? '', courseParticipationID ?? '')
        : getPeerEvaluationsForParticipantInPhase(phaseId ?? '', courseParticipationID ?? ''),
    enabled: !!phaseId && !!courseParticipationID,
  })

  const {
    data: feedbackItems = [],
    isPending: isFeedbackItemsPending,
    isError: isFeedbackItemsError,
    refetch: refetchFeedbackItems,
  } = useQuery({
    queryKey: ['student-feedback-items', phaseId, courseParticipationID],
    queryFn: () => getFeedbackItemsForStudent(phaseId ?? '', courseParticipationID ?? ''),
    enabled: !!phaseId && !!courseParticipationID,
  })

  const participant = useMemo(
    () =>
      participations.find(
        (participation) => participation.courseParticipationID === courseParticipationID,
      ),
    [courseParticipationID, participations],
  )

  const categories =
    assessmentType === AssessmentType.SELF ? selfEvaluationCategories : peerEvaluationCategories
  const pageTitle =
    assessmentType === AssessmentType.SELF ? 'Self Evaluation Results' : 'Peer Evaluation Results'

  const evaluationsByCategory = useMemo(
    () =>
      new Map(
        categories.map((category) => {
          const competencyIDs = new Set(category.competencies.map((competency) => competency.id))
          return [
            category.id,
            evaluations.filter((evaluation) => competencyIDs.has(evaluation.competencyID)),
          ]
        }),
      ),
    [categories, evaluations],
  )

  const typedFeedbackItems = useMemo(
    () => feedbackItems.filter((item) => item.type === assessmentType),
    [assessmentType, feedbackItems],
  )

  const positiveFeedbackItems = useMemo(
    () => typedFeedbackItems.filter((item) => item.feedbackType === 'positive'),
    [typedFeedbackItems],
  )

  const negativeFeedbackItems = useMemo(
    () => typedFeedbackItems.filter((item) => item.feedbackType === 'negative'),
    [typedFeedbackItems],
  )

  const isPending = isEvaluationsPending || isFeedbackItemsPending
  const isError = isEvaluationsError || isFeedbackItemsError
  const refetch = () => {
    refetchEvaluations()
    refetchFeedbackItems()
  }

  if (isError) {
    return <ErrorPage onRetry={refetch} />
  }

  if (isPending) {
    return (
      <div className='flex justify-center items-center h-64'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  if (!participant) {
    return <ErrorPage message='The requested participant could not be found.' />
  }

  const studentName = `${participant.student.firstName} ${participant.student.lastName}`

  return (
    <div className='space-y-4'>
      <ManagementPageHeader>
        {pageTitle} for {studentName}
      </ManagementPageHeader>

      {categories.length === 0 ? (
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
            {categories.map((category) => (
              <CategoryEvaluation
                key={category.id}
                category={category}
                assessmentType={assessmentType}
                evaluations={evaluationsByCategory.get(category.id) ?? []}
              />
            ))}
          </div>

          <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
            <FeedbackItemDisplayPanel
              feedbackItems={negativeFeedbackItems}
              feedbackType='negative'
              studentName={participant.student.firstName}
            />
            <FeedbackItemDisplayPanel
              feedbackItems={positiveFeedbackItems}
              feedbackType='positive'
              studentName={participant.student.firstName}
            />
          </div>
        </div>
      )}
    </div>
  )
}
