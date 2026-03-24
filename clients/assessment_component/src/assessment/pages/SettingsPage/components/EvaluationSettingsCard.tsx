import { AssessmentType } from '../../../interfaces/assessmentType'
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
  const { schemas, isSaving, evaluationCards } = useSettingsPageContext()
  const card = evaluationCards[assessmentType]

  return (
    <SchemaConfigurationCard
      {...card}
      schemas={schemas}
      disabled={isSaving}
      isSaving={isSaving}
    >
      <p className='text-xs text-slate-500'>{distinctionText}</p>
    </SchemaConfigurationCard>
  )
}
