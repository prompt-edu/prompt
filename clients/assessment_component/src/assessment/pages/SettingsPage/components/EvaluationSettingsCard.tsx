import { AssessmentType } from '../../../interfaces/assessmentType'
import { useGetAllAssessmentSchemas } from '../../hooks/useGetAllAssessmentSchemas'
import { useEvaluationSettingsCardState } from '../hooks/useEvaluationSettingsCardState'
import { SchemaConfigurationCard } from './SchemaConfigurationCard'

interface EvaluationSettingsCardProps {
  assessmentType: AssessmentType.SELF | AssessmentType.PEER | AssessmentType.TUTOR
  distinctionText: string
}

export const EvaluationSettingsCard = ({
  assessmentType,
  distinctionText,
}: EvaluationSettingsCardProps) => {
  const { isSaving, card } = useEvaluationSettingsCardState(assessmentType)
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()

  return (
    <SchemaConfigurationCard
      {...card}
      schemas={schemas ?? []}
      disabled={isSaving || isSchemasPending || isSchemasError}
      isSaving={isSaving}
    >
      <p className='text-xs text-muted-foreground'>{distinctionText}</p>
      {isSchemasError && (
        <p className='text-xs text-destructive'>
          Assessment schemas could not be loaded. Please refresh and try again.
        </p>
      )}
    </SchemaConfigurationCard>
  )
}
