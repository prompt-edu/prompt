import { ScoreLevel } from '@tumaet/prompt-shared-state'
import {
  ScoreLevelSelector as BaseScoreLevelSelector,
  getLevelConfig,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'
import { User, UserCheck, Users } from 'lucide-react'
import type { ReactNode } from 'react'
import type { Assessment } from '../../interfaces/assessment'
import { AssessmentType } from '../../interfaces/assessmentType'
import type { Competency } from '../../interfaces/competency'

import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'

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
  independentAssessments?: Assessment[]
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
  independentAssessments = [],
}: ScoreLevelSelectorProps) => {
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const descriptionsByLevel = mapCompetencyDescriptionsByLevel(competency)
  const showIndicators = coursePhaseConfig?.evaluationResultsVisible || completed
  const indicators: Partial<Record<ScoreLevel, ReactNode[]>> = {}

  if (selfEvaluationCompetency && selfEvaluationScoreLevel) {
    indicators[selfEvaluationScoreLevel] = [
      ...(indicators[selfEvaluationScoreLevel] ?? []),
      <TooltipProvider key={`self-evaluation-${selfEvaluationScoreLevel}-${competency.id}`}>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type='button'
              aria-label={`Self evaluation result: ${getLevelConfig(selfEvaluationScoreLevel).title}`}
            >
              <User size={20} className='text-blue-500 dark:text-blue-300' aria-hidden />
            </button>
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
            <button
              type='button'
              aria-label={`Peer evaluation result: ${getLevelConfig(peerEvaluationScoreLevel).title}`}
            >
              <Users size={20} className='text-green-500 dark:text-green-300' aria-hidden />
            </button>
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

  for (const scoreLevel of new Set(independentAssessments.map((a) => a.scoreLevel))) {
    const assessors = independentAssessments.filter((a) => a.scoreLevel === scoreLevel)
    indicators[scoreLevel] = [
      ...(indicators[scoreLevel] ?? []),
      <TooltipProvider key={`independent-assessment-${scoreLevel}-${competency.id}`}>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type='button'
              aria-label={`Independent assessments: ${getLevelConfig(scoreLevel).title}`}
            >
              <UserCheck size={20} className='text-purple-500 dark:text-purple-300' aria-hidden />
            </button>
          </TooltipTrigger>
          <TooltipContent>
            <div className='font-semibold'>Independent Assessments</div>
            {assessors.map((assessor) => (
              <div key={assessor.id} className='text-sm'>
                {assessor.author}
              </div>
            ))}
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
      descriptionsByLevel={descriptionsByLevel}
      showIndicators={showIndicators}
      indicators={indicators}
      hideUnselectedOnDesktop={false}
    />
  )
}
