import { ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentSettingsCard } from './components/AssessmentSettingsCard/AssessmentSettingsCard'
import { EvaluationSettingsCard } from './components/EvaluationSettingsCard'
import { useSettingsPageController } from './hooks/useSettingsPageController'

export const SettingsPage = () => {
  const controller = useSettingsPageController()

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Assessment Settings</ManagementPageHeader>

      <AssessmentSettingsCard
        isSaving={controller.isSaving}
        assessmentCard={controller.assessmentCard}
        assessmentVisibility={controller.assessmentVisibility}
      />

      <EvaluationSettingsCard
        distinctionText='Reflection by the student on their own work in this phase.'
        isSaving={controller.isSaving}
        card={controller.evaluationCards[AssessmentType.SELF]}
      />
      <EvaluationSettingsCard
        distinctionText='Feedback between peers to assess collaboration and team contribution.'
        isSaving={controller.isSaving}
        card={controller.evaluationCards[AssessmentType.PEER]}
      />
      <EvaluationSettingsCard
        distinctionText='Feedback from students about their tutors in this phase.'
        isSaving={controller.isSaving}
        card={controller.evaluationCards[AssessmentType.TUTOR]}
      />
    </div>
  )
}
