import type { ScoreLevel } from '@tumaet/prompt-shared-state'
import { Card, CardContent } from '@tumaet/prompt-ui-components'

import type { CategoryWithCompetencies } from '../../interfaces/category'
import { getScoreLevelDescription } from '../utils/getScoreLevelDescription'
import { ScoreChip } from './PrintReport/ScoreChip'

interface ResultRow {
  competencyID: string
  name: string
  description: string
  scoreLevel: ScoreLevel
}

export interface ResultCategory {
  id: string
  name: string
  rows: ResultRow[]
}

/** Keeps the competencies that have a score level, and the categories that keep any. */
export const buildResultCategories = (
  categories: CategoryWithCompetencies[],
  levelByCompetency: Map<string, ScoreLevel>,
): ResultCategory[] =>
  categories
    .map((category) => ({
      id: category.id,
      name: category.name,
      rows: category.competencies.flatMap((competency) => {
        const scoreLevel = levelByCompetency.get(competency.id)
        if (!scoreLevel) return []
        return [
          {
            competencyID: competency.id,
            name: competency.name,
            description: getScoreLevelDescription(scoreLevel, competency),
            scoreLevel,
          },
        ]
      }),
    }))
    .filter((category) => category.rows.length > 0)

export const ResultCategoryList = ({ categories }: { categories: ResultCategory[] }) => (
  <div className='space-y-4'>
    {categories.map((category) => (
      <Card key={category.id}>
        <CardContent className='space-y-3 p-4'>
          <h3 className='text-sm font-semibold'>{category.name}</h3>
          {category.rows.map((row) => (
            <div
              key={row.competencyID}
              className='flex flex-wrap items-start justify-between gap-2 border-b border-border pb-2 last:border-0 last:pb-0'
            >
              <div className='min-w-0 flex-1'>
                <p className='text-sm font-medium'>{row.name}</p>
                <p className='text-sm text-muted-foreground'>{row.description}</p>
              </div>
              <ScoreChip scoreLevel={row.scoreLevel} />
            </div>
          ))}
        </CardContent>
      </Card>
    ))}
  </div>
)
