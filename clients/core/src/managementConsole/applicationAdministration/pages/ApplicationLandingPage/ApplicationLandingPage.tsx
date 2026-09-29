import { PassStatus } from '@tumaet/prompt-shared-state'
import {
  Button,
  DemographicDistributionCard,
  groupByGender,
  groupBySemester,
  groupByStudyProgram,
  ManagementPageHeader,
  MissingConfig,
} from '@tumaet/prompt-ui-components'
import { ExternalLink } from 'lucide-react'
import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import type { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { ApplicationStatusCard } from './diagrams/ApplicationStatusCard'
import { AssessmentDiagram } from './diagrams/AssessmentDiagram'
import { useHideMailingWarning } from './hooks/useHideMailingWarning'
import { useMissingConfigs } from './hooks/useMissingConfig'

const APPLICATION_STATUS_LABELS = {
  [PassStatus.PASSED]: 'Accepted',
  [PassStatus.FAILED]: 'Rejected',
}

export const ApplicationLandingPage = () => {
  const [applicationMetaData, setApplicationMetaData] = useState<ApplicationMetaData | null>(null)
  const { pathname } = useLocation()
  const { coursePhase, participations } = useApplicationStore()
  const { hideMailingWarning } = useHideMailingWarning()
  const missingConfigs = useMissingConfigs(
    applicationMetaData,
    coursePhase,
    pathname,
    hideMailingWarning,
  )

  useParseApplicationMetaData(coursePhase, setApplicationMetaData)

  const isApplicationConfigured = getIsApplicationConfigured(applicationMetaData)

  return (
    <div>
      <div className='flex items-center justify-between gap-4'>
        <ManagementPageHeader>Application Administration</ManagementPageHeader>
        {coursePhase?.id && (
          <Button asChild variant='outline'>
            <Link to={`/apply/${coursePhase.id}`} target='_blank' rel='noopener noreferrer'>
              <ExternalLink className='h-4 w-4' />
              Open Application Form
            </Link>
          </Button>
        )}
      </div>
      <MissingConfig elements={missingConfigs} />
      <div className='grid gap-6 md:grid-cols-2 lg:grid-cols-3 mb-6'>
        <ApplicationStatusCard
          applicationMetaData={applicationMetaData}
          applicationPhaseIsConfigured={isApplicationConfigured}
        />
        <AssessmentDiagram applications={participations} />
        <DemographicDistributionCard
          title='Gender Distribution'
          description='Breakdown of student genders'
          groups={groupByGender(participations, (p) => p.student)}
          stackBy='passStatus'
          passStatusLabels={APPLICATION_STATUS_LABELS}
        />
      </div>
      <div className='grid gap-6 md:grid-cols-1 lg:grid-cols-2 mb-6'>
        <DemographicDistributionCard
          title='Study Program Distribution'
          description='Breakdown of student study programs'
          groups={groupByStudyProgram(participations, (p) => p.student)}
          stackBy='passStatus'
          passStatusLabels={APPLICATION_STATUS_LABELS}
        />
        <DemographicDistributionCard
          title='Semester Distribution'
          description='Breakdown of students by semester and degree'
          groups={groupBySemester(participations, (p) => p.student)}
          stackBy='studyDegree'
        />
      </div>
    </div>
  )
}
