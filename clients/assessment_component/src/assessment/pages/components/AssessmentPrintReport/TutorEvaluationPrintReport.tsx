import { mapNumberToScoreLevel, mapScoreLevelToNumber } from '@tumaet/prompt-shared-state'

import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { Evaluation } from '../../../interfaces/evaluation'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { useTeamStore } from '../../../zustand/useTeamStore'
import { getScoreLevelDescription } from '../../utils/getScoreLevelDescription'
import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'
import { ScoreChip } from './AssessmentPrintReport'

interface TutorEvaluationPrintReportProps {
  tutorName: string
  teamName: string
  categories: CategoryWithCompetencies[]
  evaluations: Evaluation[]
  feedbackItems?: FeedbackItem[]
}

export const TutorEvaluationPrintReport = ({
  tutorName,
  teamName,
  categories,
  evaluations,
  feedbackItems = [],
}: TutorEvaluationPrintReportProps) => {
  const { teams } = useTeamStore()

  const resolveAuthorName = (authorCourseParticipationID: string) => {
    for (const team of teams) {
      const member = team.members.find((m) => m.id === authorCourseParticipationID)
      if (member) return `${member.firstName} ${member.lastName}`
    }
    return 'Unknown member'
  }

  const sortedCategories = [...categories].sort((a, b) => a.name.localeCompare(b.name))
  const positiveFeedback = feedbackItems.filter((item) => item.feedbackType === 'positive')
  const negativeFeedback = feedbackItems.filter((item) => item.feedbackType === 'negative')

  return (
    <div className='print-report hidden text-black print:block'>
      <header className='mb-6 break-inside-avoid border-b border-gray-300 pb-4'>
        <h1 className='text-2xl font-bold'>{tutorName}</h1>
        <p className='mt-1 text-sm text-gray-700'>Tutor Evaluation Results · Team {teamName}</p>
      </header>

      {sortedCategories.map((category) => {
        const categoryEvaluations = evaluations.filter((evaluation) =>
          category.competencies.some((competency) => competency.id === evaluation.competencyID),
        )
        const categoryScore = getWeightedScoreLevel(categoryEvaluations, [category])
        const sortedCompetencies = [...category.competencies].sort((a, b) =>
          a.name.localeCompare(b.name),
        )

        return (
          <section key={category.id} className='mb-6'>
            <div className='mb-2 flex items-center justify-between gap-2 border-b border-gray-200 pb-1'>
              <h2 className='text-lg font-semibold'>{category.name}</h2>
              {categoryEvaluations.length > 0 && (
                <ScoreChip scoreLevel={mapNumberToScoreLevel(categoryScore)} />
              )}
            </div>

            <div className='space-y-3'>
              {sortedCompetencies.map((competency) => {
                const competencyEvaluations = evaluations.filter(
                  (evaluation) => evaluation.competencyID === competency.id,
                )
                const averageLevel =
                  competencyEvaluations.length > 0
                    ? mapNumberToScoreLevel(
                        competencyEvaluations.reduce(
                          (sum, evaluation) => sum + mapScoreLevelToNumber(evaluation.scoreLevel),
                          0,
                        ) / competencyEvaluations.length,
                      )
                    : undefined

                return (
                  <div
                    key={competency.id}
                    className='break-inside-avoid rounded-sm border border-gray-200 p-3'
                  >
                    <div className='flex items-center justify-between gap-2'>
                      <h3 className='text-sm font-medium'>{competency.name}</h3>
                      {averageLevel ? (
                        <ScoreChip scoreLevel={averageLevel} />
                      ) : (
                        <span className='text-xs text-gray-400'>No evaluations</span>
                      )}
                    </div>
                    {averageLevel && (
                      <p className='mt-1 text-sm text-gray-700'>
                        {getScoreLevelDescription(averageLevel, competency)}
                      </p>
                    )}
                    {competencyEvaluations.length > 0 && (
                      <ul className='mt-2 space-y-1'>
                        {competencyEvaluations.map((evaluation) => (
                          <li key={evaluation.id} className='flex items-center gap-2 text-sm'>
                            <ScoreChip scoreLevel={evaluation.scoreLevel} />
                            <span className='text-gray-700'>
                              {resolveAuthorName(evaluation.authorCourseParticipationID)}
                            </span>
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                )
              })}
            </div>
          </section>
        )
      })}

      {(positiveFeedback.length > 0 || negativeFeedback.length > 0) && (
        <section className='break-inside-avoid'>
          <h2 className='mb-2 border-b border-gray-200 pb-1 text-lg font-semibold'>Feedback</h2>
          {positiveFeedback.length > 0 && (
            <div className='mb-3'>
              <h3 className='text-sm font-medium'>What went well</h3>
              <ul className='ml-5 list-disc text-sm text-gray-700'>
                {positiveFeedback.map((item) => (
                  <li key={item.id}>{item.feedbackText}</li>
                ))}
              </ul>
            </div>
          )}
          {negativeFeedback.length > 0 && (
            <div>
              <h3 className='text-sm font-medium'>Where to improve</h3>
              <ul className='ml-5 list-disc text-sm text-gray-700'>
                {negativeFeedback.map((item) => (
                  <li key={item.id}>{item.feedbackText}</li>
                ))}
              </ul>
            </div>
          )}
        </section>
      )}
    </div>
  )
}
