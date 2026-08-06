import { ScoreChip } from '../../components/AssessmentPrintReport/ScoreChip'
import type { EvaluationResultCategory } from './EvaluationResultsSection'

interface EvaluationPrintReportProps {
  selfSections: EvaluationResultCategory[]
  peerSections: EvaluationResultCategory[]
}

const PrintSection = ({
  title,
  categories,
}: {
  title: string
  categories: EvaluationResultCategory[]
}) => {
  if (categories.length === 0) return null

  return (
    <section className='mb-6'>
      <h2 className='mb-2 border-b border-gray-300 pb-1 text-lg font-semibold'>{title}</h2>
      {categories.map((category) => (
        <div key={category.id} className='mb-4 break-inside-avoid'>
          <h3 className='mb-2 text-sm font-semibold'>{category.name}</h3>
          <div className='space-y-2'>
            {category.rows.map((row) => (
              <div
                key={row.competencyID}
                className='break-inside-avoid rounded-sm border border-gray-200 p-3'
              >
                <div className='flex items-center justify-between gap-2'>
                  <h4 className='text-sm font-medium'>{row.name}</h4>
                  <ScoreChip scoreLevel={row.scoreLevel} />
                </div>
                <p className='mt-1 text-sm text-gray-700'>{row.description}</p>
              </div>
            ))}
          </div>
        </div>
      ))}
    </section>
  )
}

export const EvaluationPrintReport = ({
  selfSections,
  peerSections,
}: EvaluationPrintReportProps) => {
  if (selfSections.length === 0 && peerSections.length === 0) return null

  return (
    <div className='print-report hidden text-black print:block'>
      <header className='mb-6 break-inside-avoid border-b border-gray-300 pb-4'>
        <h1 className='text-2xl font-bold'>Evaluation Results</h1>
      </header>

      <PrintSection title='Your self-evaluation' categories={selfSections} />
      <PrintSection title='Peer feedback (averaged)' categories={peerSections} />
    </div>
  )
}
