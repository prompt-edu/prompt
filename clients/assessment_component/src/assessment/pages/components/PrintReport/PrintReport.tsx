import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'
import { cn } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'

import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { CompetencyScore } from '../../../interfaces/competencyScore'
import type { FeedbackItem } from '../../../interfaces/feedbackItem'
import { getAverageScoreLevel } from '../../utils/getAverageScoreLevel'
import { getScoreLevelDescription } from '../../utils/getScoreLevelDescription'
import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'
import { FeedbackSection } from './FeedbackSection'
import { ScoreChip } from './ScoreChip'

export interface PrintReportScore extends CompetencyScore {
  authorName?: string
}

interface PrintReportProps {
  title: string
  subtitle?: string
  meta?: ReactNode
  categories: CategoryWithCompetencies[]
  scores: PrintReportScore[]
  categoryComments?: Record<string, string>
  feedbackItems?: FeedbackItem[]
  className?: string
  children?: ReactNode
}

// Deliberately free of responsive utilities: a real A4 print is ~680px wide
// (below the md breakpoint) while Playwright's print emulation keeps the full
// viewport, so md:/lg: variants would render differently in the two contexts.
export const PrintReport = ({
  title,
  subtitle,
  meta,
  categories,
  scores,
  categoryComments,
  feedbackItems = [],
  className,
  children,
}: PrintReportProps) => {
  const sortedCategories = [...categories].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <section className={cn('print-report hidden text-black print:block', className)}>
      <header className='mb-6 break-inside-avoid border-b border-gray-300 pb-4'>
        <h1 className='text-2xl font-bold'>{title}</h1>
        {subtitle && <p className='mt-1 text-sm text-gray-700'>{subtitle}</p>}
        {meta && (
          <div className='mt-3 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm'>{meta}</div>
        )}
      </header>

      {sortedCategories.map((category) => {
        const categoryScores = scores.filter((score) =>
          category.competencies.some((competency) => competency.id === score.competencyID),
        )
        const categoryScore = getWeightedScoreLevel(categoryScores, [category])
        const comment = categoryComments?.[category.id]
        const sortedCompetencies = [...category.competencies].sort((a, b) =>
          a.name.localeCompare(b.name),
        )

        return (
          <section key={category.id} className='mb-6'>
            <div className='mb-2 flex items-center justify-between gap-2 border-b border-gray-200 pb-1'>
              <h2 className='text-lg font-semibold'>{category.name}</h2>
              {categoryScores.length > 0 && (
                <ScoreChip scoreLevel={mapNumberToScoreLevel(categoryScore)} />
              )}
            </div>

            {comment && (
              <p className='mb-3 whitespace-pre-wrap text-sm italic text-gray-700'>{comment}</p>
            )}

            <div className='space-y-3'>
              {sortedCompetencies.map((competency) => {
                const competencyScores = categoryScores.filter(
                  (score) => score.competencyID === competency.id,
                )
                const averageScoreLevel = getAverageScoreLevel(competencyScores)

                return (
                  <div
                    key={competency.id}
                    className='break-inside-avoid rounded-sm border border-gray-200 p-3'
                  >
                    <div className='flex items-center justify-between gap-2'>
                      <h3 className='text-sm font-medium'>{competency.name}</h3>
                      {averageScoreLevel ? (
                        <ScoreChip scoreLevel={averageScoreLevel} />
                      ) : (
                        <span className='text-xs text-gray-400'>Not assessed</span>
                      )}
                    </div>
                    {averageScoreLevel && (
                      <p className='mt-1 text-sm text-gray-700'>
                        {getScoreLevelDescription(averageScoreLevel, competency)}
                      </p>
                    )}
                    {competencyScores.length > 1 && (
                      <ul className='mt-2 space-y-1'>
                        {competencyScores.map((score) => (
                          <li key={score.id} className='flex items-center gap-2 text-xs'>
                            <ScoreChip scoreLevel={score.scoreLevel} />
                            <span className='text-gray-700'>{score.authorName}</span>
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

      <FeedbackSection feedbackItems={feedbackItems} />

      {children}
    </section>
  )
}
