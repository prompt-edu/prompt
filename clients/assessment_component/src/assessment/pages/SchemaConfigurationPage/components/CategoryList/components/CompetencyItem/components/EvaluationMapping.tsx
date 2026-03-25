import { User, Users } from 'lucide-react'
import { useState } from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider,
} from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../../../../../../interfaces/assessmentType'
import { Competency } from '../../../../../../../interfaces/competency'

import { useUpdateCompetencyMapping } from '../hooks/useUpdateCompetencyMapping'

interface EvaluationMappingProps {
  assessmentType: AssessmentType
  competencies: Competency[]
  currentMapping: string | undefined
  competency: Competency
}

export const EvaluationMapping = ({
  assessmentType,
  competencies,
  currentMapping,
  competency,
}: EvaluationMappingProps) => {
  const isSelfEvaluation = assessmentType === AssessmentType.SELF
  const [error, setError] = useState<string | undefined>(undefined)
  const updateMappingMutation = useUpdateCompetencyMapping(setError)

  const handleMappingChange = (competencyId: string) => {
    const evaluationType = isSelfEvaluation ? 'self' : 'peer'

    if (competencyId === 'none') {
      // Remove mapping if "No mapping" is selected
      if (currentMapping) {
        updateMappingMutation.mutate({
          competency,
          newMappedCompetencyId: currentMapping,
          action: 'remove',
          evaluationType,
        })
      }
    } else {
      updateMappingMutation.mutate({
        competency,
        newMappedCompetencyId: competencyId,
        action: 'update',
        evaluationType,
        currentMapping,
      })
    }
  }

  return (
    <div className='flex gap-2'>
      <div className='flex items-center gap-2 mb-1'>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              {isSelfEvaluation ? (
                <User size={16} className='text-blue-500 dark:text-blue-300' />
              ) : (
                <Users size={16} className='text-green-500 dark:text-green-300' />
              )}
            </TooltipTrigger>
            <TooltipContent>
              {isSelfEvaluation ? 'Self Evaluation Mapping' : 'Peer Evaluation Mapping'}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>

      <Select value={currentMapping || 'none'} onValueChange={handleMappingChange}>
        <SelectTrigger className='flex-1 text-xs'>
          <SelectValue
            placeholder={
              isSelfEvaluation
                ? 'Select self evaluation competency'
                : 'Select peer evaluation competency'
            }
          />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value='none'>
            <span className='text-xs text-muted-foreground'>No mapping</span>
          </SelectItem>
          {competencies.map((comp) => (
            <SelectItem key={comp.id} value={comp.id}>
              <span className='text-xs'>{comp.name}</span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {error && <div className='text-xs text-destructive mt-1'>{error}</div>}
    </div>
  )
}
