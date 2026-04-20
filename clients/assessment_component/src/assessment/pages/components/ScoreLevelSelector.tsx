import type { ReactNode } from 'react'
import { User, Users } from 'lucide-react'
import {
  ScoreLevelSelector as BaseScoreLevelSelector,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../interfaces/assessmentType'
import { Competency } from '../../interfaces/competency'
import { ScoreLevel } from '@tumaet/prompt-shared-state'

import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'

interface ScoreLevelSelectorProps {
  className?: string
  competency: Competency
  selectedScore?: ScoreLevel
  onScoreChange: (value: ScoreLevel) => void
  completed: boolean
  assessmentType?: AssessmentType
  selfEvaluationCompetency?: Competency
  selfEvaluationScoreLevel?: ScoreLevel
  selfEvaluationStudentAnswers?: (() => ReactNode)[]
  peerEvaluationCompetency?: Competency
  peerEvaluationScoreLevel?: ScoreLevel
  peerEvaluationStudentAnswers?: (() => ReactNode)[]
}

const mapCompetencyDescriptionsByLevel = (competency: Competency): Record<ScoreLevel, string> => ({
  [ScoreLevel.VeryBad]: competency.descriptionVeryBad,
  [ScoreLevel.Bad]: competency.descriptionBad,
  [ScoreLevel.Ok]: competency.descriptionOk,
  [ScoreLevel.Good]: competency.descriptionGood,
  [ScoreLevel.VeryGood]: competency.descriptionVeryGood,
})

export const ScoreLevelSelector = ({
  className,
  competency,
  selectedScore,
  onScoreChange,
  completed,
  assessmentType = AssessmentType.ASSESSMENT,
  selfEvaluationCompetency,
  selfEvaluationScoreLevel,
  selfEvaluationStudentAnswers,
  peerEvaluationCompetency,
  peerEvaluationScoreLevel,
  peerEvaluationStudentAnswers,
}: ScoreLevelSelectorProps) => {
  const { coursePhaseConfig } = useCoursePhaseConfigStore()
  const showIndicators = coursePhaseConfig?.evaluationResultsVisible || completed
  const indicators: Partial<Record<ScoreLevel, ReactNode[]>> = {}

  if (selfEvaluationCompetency && selfEvaluationScoreLevel) {
    indicators[selfEvaluationScoreLevel] = [
      ...(indicators[selfEvaluationScoreLevel] ?? []),
      <TooltipProvider key={`self-evaluation-${selfEvaluationScoreLevel}-${competency.id}`}>
        <Tooltip>
          <TooltipTrigger asChild>
            <User size={20} className='text-blue-500 dark:text-blue-300' />
          </TooltipTrigger>
          <TooltipContent>
            <div className='font-semibold'>Self Evaluation Results</div>
            <div className='text-sm text-muted-foreground'>
              <span className='font-semibold'>Statement:</span> {selfEvaluationCompetency.name}
            </div>
            {selfEvaluationStudentAnswers && selfEvaluationStudentAnswers.length > 0 ? (
              <div className='mt-2 space-y-1'>
                {selfEvaluationStudentAnswers.map((studentAnswer, index) => (
                  <div key={`self-answer-${index}`}>{studentAnswer()}</div>
                ))}
              </div>
            ) : null}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>,
    ]
  }

  if (peerEvaluationCompetency && peerEvaluationScoreLevel) {
    indicators[peerEvaluationScoreLevel] = [
      ...(indicators[peerEvaluationScoreLevel] ?? []),
      <TooltipProvider key={`peer-evaluation-${peerEvaluationScoreLevel}-${competency.id}`}>
        <Tooltip>
          <TooltipTrigger asChild>
            <Users size={20} className='text-green-500 dark:text-green-300' />
          </TooltipTrigger>
          <TooltipContent>
            {assessmentType !== AssessmentType.TUTOR ? (
              <div>
                <div className='font-semibold'>Peer Evaluation Results</div>
                <div className='text-sm text-muted-foreground'>
                  <span className='font-semibold'>Statement:</span> {peerEvaluationCompetency.name}
                </div>
              </div>
            ) : null}
            {peerEvaluationStudentAnswers && peerEvaluationStudentAnswers.length > 0
              ? peerEvaluationStudentAnswers.map((studentAnswer, index) => (
                  <div key={`peer-answer-${index}`}>{studentAnswer()}</div>
                ))
              : undefined}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>,
    ]
  }

  return (
    <BaseScoreLevelSelector
      className={className}
      selectedScore={selectedScore}
      onScoreChange={onScoreChange}
      completed={completed}
      descriptionsByLevel={mapCompetencyDescriptionsByLevel(competency)}
      showIndicators={showIndicators}
      indicators={indicators}
      hideUnselectedOnDesktop={false}
    />
  )
}
