import { useState } from 'react'
import { ChevronRight, ChevronDown } from 'lucide-react'

import { CategoryWithCompetencies } from '../../../interfaces/category'
import { Assessment } from '../../../interfaces/assessment'
import { AggregatedEvaluationResult } from '../../../interfaces/assessmentResults'
import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'

import { useStudentAssessmentStore } from '../../../zustand/useStudentAssessmentStore'

import { getWeightedScoreLevel } from '../../utils/getWeightedScoreLevel'

import { AssessmentStatusBadge, StudentScoreBadge } from '../../components/badges'

import { AssessmentForm } from './AssessmentForm/AssessmentForm'
import { CategoryComment } from './CategoryComment/CategoryComment'

interface CategoryAssessmentProps {
  category: CategoryWithCompetencies
  assessments: Assessment[]
  completed: boolean
  courseParticipationID: string
  peerEvaluationResults?: AggregatedEvaluationResult[]
  selfEvaluationResults?: AggregatedEvaluationResult[]
  hidePeerEvaluationDetails?: boolean
}

export const CategoryAssessment = ({
  category,
  assessments,
  completed,
  courseParticipationID,
  peerEvaluationResults,
  selfEvaluationResults,
  hidePeerEvaluationDetails = false,
}: CategoryAssessmentProps) => {
  const categoryAssessment = useStudentAssessmentStore((state) =>
    state.categoryAssessments.find((ca) => ca.categoryID === category.id),
  )

  const [isExpanded, setIsExpanded] = useState(true)

  const toggleExpand = () => {
    setIsExpanded(!isExpanded)
  }

  const categoryScore = getWeightedScoreLevel(assessments, [category])
  const sortedCompetencies = [...category.competencies].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <div className='mb-6'>
      <div className='flex items-center justify-center mb-4 gap-1'>
        <div className='flex items-center justify-center w-full'>
          <button
            onClick={toggleExpand}
            className='p-1 mr-2 hover:bg-accent rounded-xs focus:outline-hidden'
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
            />
          )}
        </div>
      </div>

      {isExpanded && (
        <div id={`content-${category.id}`} className='space-y-5'>
          <CategoryComment
            categoryID={category.id}
            courseParticipationID={courseParticipationID}
            categoryAssessment={categoryAssessment}
            completed={completed}
          />
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
                      courseParticipationID={courseParticipationID}
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
