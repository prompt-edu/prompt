import { ScoreLevel } from '@tumaet/prompt-shared-state'
import { Fragment } from 'react'

import type { CategoryWithCompetencies } from '../../../interfaces/category'
import { PrintReport } from '../../components/PrintReport/PrintReport'
import { ScoreChip } from '../../components/PrintReport/ScoreChip'
import { getScoreLevelDescription } from '../../utils/getScoreLevelDescription'

interface SchemaPrintReportProps {
  categories: CategoryWithCompetencies[]
  schemaType: string
  schemaName?: string
  schemaDescription?: string
}

const byName = (a: { name: string }, b: { name: string }) => a.name.localeCompare(b.name)

export const SchemaPrintReport = ({
  categories,
  schemaType,
  schemaName,
  schemaDescription,
}: SchemaPrintReportProps) => (
  <PrintReport
    title={schemaName || `${schemaType} schema`}
    subtitle={schemaDescription}
    meta={
      <span>
        <strong>Type:</strong> {schemaType}
      </span>
    }
  >
    {[...categories].sort(byName).map((category) => (
      <section key={category.id} className='mb-6'>
        <div className='mb-2 flex items-center justify-between gap-2 border-b border-gray-200 pb-1'>
          <h2 className='text-lg font-semibold'>{category.name}</h2>
          <span className='text-xs text-gray-600'>Weight: {category.weight}</span>
        </div>

        {category.description && (
          <p className='mb-3 whitespace-pre-wrap text-sm text-gray-700'>{category.description}</p>
        )}

        <div className='space-y-3'>
          {[...category.competencies].sort(byName).map((competency) => (
            <div
              key={competency.id}
              className='break-inside-avoid rounded-sm border border-gray-200 p-3'
            >
              <div className='flex items-center justify-between gap-2'>
                <h3 className='text-sm font-medium'>{competency.name}</h3>
                <span className='text-xs text-gray-600'>Weight: {competency.weight}</span>
              </div>
              {competency.description && (
                <p className='mt-1 text-sm text-gray-700'>{competency.description}</p>
              )}
              <dl className='mt-2 grid grid-cols-[auto_1fr] items-baseline justify-items-start gap-x-3 gap-y-1 text-xs'>
                {Object.values(ScoreLevel).map((scoreLevel) => (
                  <Fragment key={scoreLevel}>
                    <dt>
                      <ScoreChip scoreLevel={scoreLevel} />
                    </dt>
                    <dd className='text-gray-700'>
                      {getScoreLevelDescription(scoreLevel, competency)}
                    </dd>
                  </Fragment>
                ))}
              </dl>
            </div>
          ))}
        </div>
      </section>
    ))}
  </PrintReport>
)
