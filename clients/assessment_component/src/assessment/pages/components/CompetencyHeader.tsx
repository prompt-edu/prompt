import { ClipboardCheck, RotateCcw } from 'lucide-react'
import {
  Button,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@tumaet/prompt-ui-components'

import { Competency } from '../../interfaces/competency'
import { CompetencyScore } from '../../interfaces/competencyScore'
import { AssessmentType } from '../../interfaces/assessmentType'

import { useStudentEvaluationStore } from '../../zustand/useStudentEvaluationStore'

interface CompetencyHeaderProps {
  className?: string
  competency: Competency
  competencyScore?: CompetencyScore
  completed: boolean
  onResetClick: () => void
  assessmentType?: AssessmentType
}

export const CompetencyHeader = ({
  className,
  competency,
  competencyScore,
  completed,
  onResetClick,
  assessmentType = AssessmentType.ASSESSMENT,
}: CompetencyHeaderProps) => {
  const { studentName } = useStudentEvaluationStore()
  const competencyName =
    assessmentType === AssessmentType.PEER
      ? competency.name.replace(/This person|this person/g, studentName ?? 'This Person')
      : competency.name

  return (
    <div className={className}>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <ClipboardCheck className='h-4 w-4 text-muted-foreground shrink-0' />
          <h3 className='text-base font-medium'>{competencyName}</h3>
        </div>
        {competencyScore && !completed && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant='ghost' onClick={onResetClick}>
                  <RotateCcw />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Reset this selection</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
      {assessmentType === AssessmentType.ASSESSMENT && (
        <p className='text-xs text-muted-foreground line-clamp-2'>{competency.description}</p>
      )}
    </div>
  )
}
