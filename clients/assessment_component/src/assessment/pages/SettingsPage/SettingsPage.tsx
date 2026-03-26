import { ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentSettingsCard } from './components/AssessmentSettingsCard/AssessmentSettingsCard'
import { EvaluationSettingsCard } from './components/EvaluationSettingsCard'

export const SettingsPage = () => {
  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Assessment Settings</ManagementPageHeader>

      <AssessmentSettingsCard />

      <EvaluationSettingsCard
        assessmentType={AssessmentType.SELF}
        distinctionText='Reflection by the student on their own work in this phase.'
      />
      <EvaluationSettingsCard
        assessmentType={AssessmentType.PEER}
        distinctionText='Feedback between peers to assess collaboration and team contribution.'
      />
      <EvaluationSettingsCard
        assessmentType={AssessmentType.TUTOR}
        distinctionText='Feedback from students about their tutors in this phase.'
      />
    </div>
  )
}
