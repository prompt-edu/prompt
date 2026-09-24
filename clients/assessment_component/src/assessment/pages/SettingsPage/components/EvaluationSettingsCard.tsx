import { Input, Label } from '@tumaet/prompt-ui-components'
import { AssessmentType } from '../../../interfaces/assessmentType'
import { useGetAllAssessmentSchemas } from '../../hooks/useGetAllAssessmentSchemas'
import { DEFAULT_TUTOR_LABEL } from '../../hooks/useTutorLabel'
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
  const { isSaving, card, tutorDisplayName, onTutorDisplayNameChange } =
    useEvaluationSettingsCardState(assessmentType)
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
      {assessmentType === AssessmentType.TUTOR && (
        <div className='max-w-sm space-y-2'>
          <Label htmlFor='tutor-display-name'>Display name</Label>
          <Input
            id='tutor-display-name'
            value={tutorDisplayName}
            onChange={(event) => onTutorDisplayNameChange(event.target.value)}
            placeholder={DEFAULT_TUTOR_LABEL}
            maxLength={50}
            disabled={isSaving}
          />
          <p className='text-xs text-muted-foreground'>
            {`How tutors are called in this phase, for example Coach or PL. Leave empty to use "${DEFAULT_TUTOR_LABEL}".`}
          </p>
        </div>
      )}
      <p className='text-xs text-muted-foreground'>{distinctionText}</p>
      {isSchemasError && (
        <p className='text-xs text-destructive'>
          Assessment schemas could not be loaded. Please refresh and try again.
        </p>
      )}
    </SchemaConfigurationCard>
  )
}
