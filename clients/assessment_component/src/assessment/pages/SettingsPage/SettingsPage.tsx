import { Card, ManagementPageHeader } from '@tumaet/prompt-ui-components'

import { AssessmentType } from '../../interfaces/assessmentType'
import { useGetCoursePhaseConfig } from '../hooks/useGetCoursePhaseConfig'
import { AssessmentReminderCard } from './components/AssessmentReminderCard/AssessmentReminderCard'
import { AssessmentSettingsCard } from './components/AssessmentSettingsCard/AssessmentSettingsCard'
import { ReleaseResultsSection } from './components/AssessmentSettingsCard/components/ReleaseResultsSection'
import { EvaluationSettingsCard } from './components/EvaluationSettingsCard'
import { GradeExportCard } from './components/GradeExportCard/GradeExportCard'

export const SettingsPage = () => {
  const { data: coursePhaseConfig } = useGetCoursePhaseConfig()
  const assessmentEnabled = coursePhaseConfig?.assessmentEnabled ?? true

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

      {/* Outside the assessment card: evaluation-only phases release results too */}
      <Card className='border-border shadow-xs'>
        <div className='space-y-6 p-6'>
          <div className='space-y-2'>
            <h2 className='text-xl font-semibold text-foreground'>Results</h2>
            <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>
              Control when students in this phase can see what this phase produced.
            </p>
          </div>

          <ReleaseResultsSection isSaving={false} />
        </div>
      </Card>

      <AssessmentReminderCard />

      {assessmentEnabled && <GradeExportCard />}
    </div>
  )
}
