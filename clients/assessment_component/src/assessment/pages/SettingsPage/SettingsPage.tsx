import { ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'

import { AssessmentType } from '../../interfaces/assessmentType'
import { AssessmentSettingsCard } from './components/AssessmentSettingsCard/AssessmentSettingsCard'
import { EvaluationSettingsCard } from './components/EvaluationSettingsCard'
import { SettingsPageProvider } from './components/SettingsPageContext'
import { useSettingsPageController } from './hooks/useSettingsPageController'

export const SettingsPage = () => {
  const controller = useSettingsPageController()

  if (controller.isSchemasError) {
    return <ErrorPage />
  }

  if (controller.isSchemasPending) {
    return (
      <div className='flex h-64 items-center justify-center'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

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
