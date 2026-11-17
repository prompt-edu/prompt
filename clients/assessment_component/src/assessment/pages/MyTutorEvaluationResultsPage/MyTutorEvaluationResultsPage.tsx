import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  ManagementPageHeader,
  QueryGate,
} from '@tumaet/prompt-ui-components'
import { ArrowLeft, MessageCircle } from 'lucide-react'
import { type ReactNode, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import type { AnonymousFeedbackItem } from '../../interfaces/evaluationResults'
import type { FeedbackType } from '../../interfaces/feedbackItem'
import { buildResultCategories, ResultCategoryList } from '../components/ResultCategoryList'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../hooks/useGetEvaluationCategoriesWithCompetencies'
import { useGetMyTutorEvaluationResults } from '../hooks/useGetMyTutorEvaluationResults'
import { getTutorLabel } from '../hooks/useTutorLabel'

const NoticeCard = ({ children }: { children: ReactNode }) => (
  <Card>
    <CardContent className='p-6'>
      <p className='text-sm text-muted-foreground'>{children}</p>
    </CardContent>
  </Card>
)

const AnonymousFeedbackPanel = ({
  feedbackItems,
  feedbackType,
}: {
  feedbackItems: AnonymousFeedbackItem[]
  feedbackType: FeedbackType
}) => {
  const items = feedbackItems.filter((item) => item.feedbackType === feedbackType)
  if (items.length === 0) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <MessageCircle
            className={`h-5 w-5 ${feedbackType === 'positive' ? 'text-green-600' : 'text-red-600'}`}
          />
          {feedbackType === 'positive'
            ? 'What did you do particularly well?'
            : 'Where could you improve?'}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ul className='space-y-2'>
          {items.map((item, index) => (
            <li
              // The texts carry no id, so they stay anonymous; their order is fixed by the server.
              key={`${index}-${item.feedbackText}`}
              className='rounded-md border border-border bg-muted/40 p-3 text-sm whitespace-pre-wrap'
            >
              {item.feedbackText}
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  )
}

export const MyTutorEvaluationResultsPage = () => {
  const navigate = useNavigate()
  const coursePhaseConfigQuery = useGetCoursePhaseConfig()
  const coursePhaseConfig = coursePhaseConfigQuery.data
  const tutorLabel = getTutorLabel(coursePhaseConfig)

  const resultsQuery = useGetMyTutorEvaluationResults()
  const { data: results, enabled } = resultsQuery

  const { data: tutorCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.TUTOR,
    enabled,
  )

  const categories = useMemo(
    () =>
      buildResultCategories(
        tutorCategories,
        new Map(
          (results?.results ?? []).map((result) => [
            result.competencyID,
            mapNumberToScoreLevel(result.averageScoreNumeric),
          ]),
        ),
      ),
    [tutorCategories, results],
  )

  return (
    <div className='w-full px-4 py-6 text-left'>
      <div className='mb-4'>
        <Button variant='outline' onClick={() => navigate('..')} className='gap-2'>
          <ArrowLeft className='h-4 w-4' />
          Back to overview
        </Button>
      </div>
      <QueryGate queries={[coursePhaseConfigQuery]}>
        {() => (
          <>
            <ManagementPageHeader>Your {tutorLabel.title} Evaluation Results</ManagementPageHeader>

            {!enabled ? (
              <NoticeCard>
                {coursePhaseConfig?.tutorEvaluationEnabled
                  ? `${tutorLabel.title} evaluation results have not been released yet.`
                  : `This phase has no ${tutorLabel.text} evaluation.`}
              </NoticeCard>
            ) : (
              <QueryGate queries={[resultsQuery]}>
                {() => {
                  if (!results) {
                    return (
                      <NoticeCard>
                        There are no {tutorLabel.text} evaluation results for you in this phase,
                        because you are not listed as the {tutorLabel.text} of a team.
                      </NoticeCard>
                    )
                  }

                  return (
                    <div className='space-y-6'>
                      <p className='text-sm text-muted-foreground'>
                        {results.evaluatorCount === 1
                          ? '1 student evaluated you.'
                          : `${results.evaluatorCount} students evaluated you.`}{' '}
                        Scores are averaged across all of them, and competencies rated by only one
                        student are not shown. Positive and negative feedback are each shown without
                        their authors once at least two students wrote some.
                      </p>

                      {categories.length > 0 ? (
                        <ResultCategoryList categories={categories} />
                      ) : (
                        <NoticeCard>
                          Not enough students evaluated you yet to show averaged scores.
                        </NoticeCard>
                      )}

                      {results.feedbackItems.length > 0 && (
                        <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                          <AnonymousFeedbackPanel
                            feedbackItems={results.feedbackItems}
                            feedbackType='negative'
                          />
                          <AnonymousFeedbackPanel
                            feedbackItems={results.feedbackItems}
                            feedbackType='positive'
                          />
                        </div>
                      )}
                    </div>
                  )
                }}
              </QueryGate>
            )}
          </>
        )}
      </QueryGate>
    </div>
  )
}
