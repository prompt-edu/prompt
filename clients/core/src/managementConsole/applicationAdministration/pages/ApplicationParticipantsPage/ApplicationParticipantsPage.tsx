import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { ApplicationParticipantsTable } from './components/table/ApplicationParticipantsTable'
import { ReactNode } from 'react'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import AssessmentScoreUpload from './components/ScoreUpload/ScoreUpload'
import { ApplicationManualAddingDialog } from './components/ApplicationManualAddingDialog/ApplicationManualAddingDialog'
import { useParams } from 'react-router-dom'

export const ApplicationParticipantsPage = (): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { participations, coursePhase } = useApplicationStore()
  const customScoresEnabled = Boolean(coursePhase?.restrictedData?.['useCustomScores'])

  return (
    <div className='relative flex flex-col min-w-0'>
      <div className='flex justify-between'>
        <ManagementPageHeader>Application Participants</ManagementPageHeader>
        <div className='flex gap-3'>
          {participations && customScoresEnabled && (
            <AssessmentScoreUpload applications={participations} />
          )}
          <ApplicationManualAddingDialog existingApplications={participations ?? []} />
        </div>
      </div>
      <ApplicationParticipantsTable phaseId={phaseId!} />
    </div>
  )
}
