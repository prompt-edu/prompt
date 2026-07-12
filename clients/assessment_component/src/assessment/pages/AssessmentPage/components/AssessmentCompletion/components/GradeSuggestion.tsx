import { mapNumberToScoreLevel } from '@tumaet/prompt-shared-state'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { AssessmentType } from '../../../../../interfaces/assessmentType'
import { useStudentAssessmentStore } from '../../../../../zustand/useStudentAssessmentStore'
import { StudentScoreBadge } from '../../../../components/badges'
import { useGetCoursePhaseConfig } from '../../../../hooks/useGetCoursePhaseConfig'
import { useGetEvaluationCategoriesWithCompetencies } from '../../../../hooks/useGetEvaluationCategoriesWithCompetencies'
import { getWeightedScoreLevel } from '../../../../utils/getWeightedScoreLevel'
import { GRADE_SELECT_OPTIONS } from '../../../../utils/gradeConfig'

interface GradeSuggestionProps {
  onGradeSuggestionChange: (value: string) => void
  readOnly?: boolean
}

export const GradeSuggestion = ({
  onGradeSuggestionChange,
  readOnly = false,
}: GradeSuggestionProps) => {
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const { data: selfEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.SELF,
    coursePhaseConfig?.selfEvaluationEnabled ?? false,
  )
  const { data: peerEvaluationCategories } = useGetEvaluationCategoriesWithCompetencies(
    AssessmentType.PEER,
    coursePhaseConfig?.peerEvaluationEnabled ?? false,
  )
  const { studentScore, assessmentCompletion, selfEvaluations, peerEvaluations } =
    useStudentAssessmentStore()
  const showAverages = !readOnly && (coursePhaseConfig?.evaluationResultsVisible ?? false)
  const isCompleted = readOnly || (assessmentCompletion?.completed ?? false)
  const gradeSuggestionValue =
    assessmentCompletion?.gradeSuggestion && assessmentCompletion.gradeSuggestion > 0
      ? assessmentCompletion.gradeSuggestion.toFixed(1)
      : ''

  return (
    <Card>
      <CardHeader>
        <CardTitle className='mb-3'>Grade</CardTitle>
        {selfEvaluations &&
          selfEvaluations.length > 0 &&
          (() => {
            const weightedScoreLevel = getWeightedScoreLevel(
              selfEvaluations,
              selfEvaluationCategories,
            )
            return (
              <div className='flex flex-row items-center gap-2'>
                <p className='text-sm text-muted-foreground'>Self Evaluation Average:</p>
                <StudentScoreBadge
                  scoreLevel={mapNumberToScoreLevel(weightedScoreLevel)}
                  scoreNumeric={weightedScoreLevel}
                />
              </div>
            )
          })()}
        {peerEvaluations &&
          peerEvaluations.length > 0 &&
          (() => {
            const weightedScoreLevel = getWeightedScoreLevel(
              peerEvaluations,
              peerEvaluationCategories,
            )
            return (
              <div className='flex flex-row items-center gap-2'>
                <p className='text-sm text-muted-foreground'>Peer Evaluation Average:</p>
                <StudentScoreBadge
                  scoreLevel={mapNumberToScoreLevel(weightedScoreLevel)}
                  scoreNumeric={weightedScoreLevel}
                />
              </div>
            )
          })()}
        {showAverages && studentScore && studentScore.scoreNumeric > 0 && (
          <div className='flex flex-row items-center gap-2'>
            <p className='text-sm text-muted-foreground'>Your Assessment Average:</p>
            <StudentScoreBadge
              scoreLevel={studentScore.scoreLevel}
              scoreNumeric={studentScore.scoreNumeric}
              showTooltip={true}
            />
          </div>
        )}
      </CardHeader>
      <CardContent>
        <p className='text-sm font-medium text-gray-700 dark:text-gray-300 mb-2'>
          Your Grade Suggestion
        </p>
        {!readOnly && coursePhaseConfig?.gradeSuggestionVisible && (
          <p className='text-xs text-gray-500 dark:text-gray-400 my-1'>
            Your suggestion will be visible to the student once results are released.
          </p>
        )}
        <Select
          value={gradeSuggestionValue}
          onValueChange={onGradeSuggestionChange}
          disabled={isCompleted}
        >
          <SelectTrigger>
            <SelectValue placeholder='Select a Grade Suggestion for this Student ...' />
          </SelectTrigger>
          <SelectContent>
            {GRADE_SELECT_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </CardContent>
    </Card>
  )
}
