import { AssessmentType } from '../../../interfaces/assessmentType'
import { useGetAllAssessmentSchemas } from '../../hooks/useGetAllAssessmentSchemas'
import { SchemaConfigurationCard } from './SchemaConfigurationCard'
import { useSettingsPageContext } from './SettingsPageContext'

type EvaluationAssessmentType = AssessmentType.SELF | AssessmentType.PEER | AssessmentType.TUTOR

interface EvaluationSettingsCardProps {
  assessmentType: EvaluationAssessmentType
  distinctionText: string
}

export const EvaluationSettingsCard = ({
  assessmentType,
  distinctionText,
}: EvaluationSettingsCardProps) => {
  const { isSaving, evaluationCards } = useSettingsPageContext()
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()
  const card = evaluationCards[assessmentType]

  return (
    <SchemaConfigurationCard
      {...card}
      schemas={schemas ?? []}
      disabled={isSaving || isSchemasPending || isSchemasError}
      isSaving={isSaving}
    >
      <p className='text-xs text-slate-500'>{distinctionText}</p>
      {isSchemasError && (
        <p className='text-xs text-rose-600'>
          Assessment schemas could not be loaded. Please refresh and try again.
        </p>
      )}
    </SchemaConfigurationCard>
  )
}
