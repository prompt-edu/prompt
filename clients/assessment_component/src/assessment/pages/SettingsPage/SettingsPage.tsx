import { ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentSettingsCard } from './components/AssessmentSettingsCard/AssessmentSettingsCard'
import { EvaluationSettingsCard } from './components/EvaluationSettingsCard'
import { SettingsPageProvider } from './components/SettingsPageContext'
import { useSettingsPageController } from './hooks/useSettingsPageController'

export const SettingsPage = () => {
  const controller = useSettingsPageController()

  return (
    <SettingsPageProvider value={controller}>
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
          distinctionText='Independent tutor perspective to complement peer and self evaluations.'
        />
      </div>
    </SettingsPageProvider>
  )
}
