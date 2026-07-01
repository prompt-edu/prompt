import { cn } from '@tumaet/prompt-ui-components'

import type { StudentAssessment } from '../../../interfaces/studentAssessment'

import {
  AssessmentStatusBadge,
  GradeSuggestionBadgeWithTooltip,
  StudentScoreBadge,
} from '../../components/badges'

interface StudentAssessmentBadgesProps {
  studentAssessment: StudentAssessment
  remainingAssessments: number
  className?: string
}

export const StudentAssessmentBadges = ({
  studentAssessment,
  remainingAssessments,
  className,
}: StudentAssessmentBadgesProps) => (
  <div className={cn('flex flex-wrap items-center justify-center gap-1', className)}>
    <AssessmentStatusBadge
      remainingAssessments={remainingAssessments}
      isFinalized={studentAssessment.assessmentCompletion.completed}
      className='h-6'
    />
    {studentAssessment.assessmentCompletion && (
      <GradeSuggestionBadgeWithTooltip
        gradeSuggestion={studentAssessment.assessmentCompletion.gradeSuggestion}
        text={true}
        className='h-6'
      />
    )}
    {studentAssessment.assessments.length > 0 && (
      <StudentScoreBadge
        scoreLevel={studentAssessment.studentScore.scoreLevel}
        scoreNumeric={studentAssessment.studentScore.scoreNumeric}
        showTooltip={true}
        className='h-6'
      />
    )}
  </div>
)
