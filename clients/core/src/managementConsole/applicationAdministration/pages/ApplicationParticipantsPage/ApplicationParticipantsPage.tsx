import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { ApplicationManualAddingDialog } from './components/ApplicationManualAddingDialog/ApplicationManualAddingDialog'
import { ImportStudents } from './components/ImportStudents/ImportStudents'
import AssessmentScoreUpload from './components/ScoreUpload/ScoreUpload'
import { ApplicationParticipantsTable } from './components/table/ApplicationParticipantsTable'

export const ApplicationParticipantsPage = (): ReactNode => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const { participations, coursePhase } = useApplicationStore()
  const customScoresEnabled = Boolean(coursePhase?.restrictedData?.useCustomScores)
  const importModeEnabled = coursePhase?.restrictedData?.applicationMode === 'import'

  return (
    <div className='relative flex flex-col min-w-0'>
      <div className='flex justify-between'>
        <ManagementPageHeader>Application Participants</ManagementPageHeader>
        <div className='flex gap-3'>
          {participations && customScoresEnabled && (
            <AssessmentScoreUpload applications={participations} />
          )}
          {importModeEnabled && <ImportStudents existingApplications={participations ?? []} />}
          {/* The manual-add dialog stays available in import mode: it is the only way to add a
              student without a TUM login (e.g. an exchange student), which the CSV import rejects. */}
          <ApplicationManualAddingDialog existingApplications={participations ?? []} />
        </div>
      </div>
      <ApplicationParticipantsTable phaseId={phaseId!} />
    </div>
  )
}
