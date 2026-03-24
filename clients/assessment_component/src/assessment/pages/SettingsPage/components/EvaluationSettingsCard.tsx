import { useGetAllAssessmentSchemas } from '../../hooks/useGetAllAssessmentSchemas'
import type { SettingsCardModel } from '../hooks/useSettingsPageController'
import { SchemaConfigurationCard } from './SchemaConfigurationCard'

interface EvaluationSettingsCardProps {
  distinctionText: string
  isSaving: boolean
  card: SettingsCardModel
}

export const EvaluationSettingsCard = ({
  distinctionText,
  isSaving,
  card,
}: EvaluationSettingsCardProps) => {
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
      <p className='text-xs text-slate-500'>{distinctionText}</p>
      {isSchemasError && (
        <p className='text-xs text-rose-600'>
          Assessment schemas could not be loaded. Please refresh and try again.
        </p>
      )}
    </SchemaConfigurationCard>
  )
}
