import { useParams } from 'react-router-dom'
import { useState } from 'react'
import { ChevronRight, ChevronDown } from 'lucide-react'

import { CategoryWithCompetencies } from '../../../interfaces/category'
import { Assessment } from '../../../interfaces/assessment'
import { AggregatedEvaluationResult } from '../../../interfaces/assessmentResults'
import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'

import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'

import { AssessmentStatusBadge, StudentScoreBadge } from '../../components/badges'

import { AssessmentForm } from './AssessmentForm/AssessmentForm'

interface CategoryAssessmentProps {
  category: CategoryWithCompetencies
  assessments: Assessment[]
  completed: boolean
  peerEvaluationResults?: AggregatedEvaluationResult[]
  selfEvaluationResults?: AggregatedEvaluationResult[]
  hidePeerEvaluationDetails?: boolean
}

export const CategoryAssessment = ({
  category,
  assessments,
  completed,
  peerEvaluationResults,
  selfEvaluationResults,
  hidePeerEvaluationDetails = false,
}: CategoryAssessmentProps) => {
  const { courseParticipationID } = useParams<{
    courseParticipationID: string
  }>()

  const [isExpanded, setIsExpanded] = useState(true)

  const toggleExpand = () => {
    setIsExpanded(!isExpanded)
  }

  const categoryScore = getWeightedScoreLevel(assessments, [category])
  const sortedCompetencies = [...category.competencies].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <div key={category.id} className='mb-6'>
      <div className='flex grow items-center justify-center mb-4 gap-1'>
        <div className='flex items-center justify-center w-full'>
          <button
            onClick={toggleExpand}
            className='p-1 mr-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xs focus:outline-hidden'
            aria-expanded={isExpanded}
            aria-controls={`content-${category.id}`}
          >
            {isExpanded ? <ChevronDown size={18} /> : <ChevronRight size={18} />}
          </button>

          <h2 className='text-xl font-semibold tracking-tight grow'>{category.name}</h2>
        </div>

        <div className='flex items-center justify-center gap-1'>
          {assessments.length > 0 && (
            <StudentScoreBadge
              scoreLevel={mapNumberToScoreLevel(categoryScore)}
              scoreNumeric={categoryScore}
            />
          )}
          {!completed && (
            <AssessmentStatusBadge
              remainingAssessments={category.competencies.length - assessments.length}
              isFinalized={completed}
            />
          )}
        </div>
      </div>

      {isExpanded && (
        <div id={`content-${category.id}`} className='pt-4 pb-2 space-y-5 border-t mt-2'>
          {category.competencies.length === 0 ? (
            <p className='text-sm text-muted-foreground italic'>
              No competencies available in this category.
            </p>
          ) : (
            <div className='grid gap-4'>
              {sortedCompetencies.map((competency) => {
                const assessment = assessments.find((ass) => ass.competencyID === competency.id)
                const peerAverage = peerEvaluationResults?.find(
                  (result) => result.competencyID === competency.id,
                )
                const selfAverage = selfEvaluationResults?.find(
                  (result) => result.competencyID === competency.id,
                )

                return (
                  <div key={competency.id}>
                    <AssessmentForm
                      courseParticipationID={courseParticipationID ?? ''}
                      competency={competency}
                      assessment={assessment}
                      completed={completed}
                      peerEvaluationAverageScore={peerAverage?.averageScoreNumeric}
                      selfEvaluationAverageScore={selfAverage?.averageScoreNumeric}
                      hidePeerEvaluationDetails={hidePeerEvaluationDetails}
                    />
                  </div>
                )
              })}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
