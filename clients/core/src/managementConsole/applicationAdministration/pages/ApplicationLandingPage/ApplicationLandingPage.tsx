import { useStudyPrograms } from '@core/network/hooks/useStudyPrograms'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  ManagementPageHeader,
  MissingConfig,
  ParticipationStatistics,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { ExternalLink } from 'lucide-react'
import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { useParseApplicationMetaData } from '../../hooks/useParseApplicationMetaData'
import type { ApplicationMetaData } from '../../interfaces/applicationMetaData'
import { getIsApplicationConfigured } from '../../utils/getApplicationIsConfigured'
import { useApplicationStore } from '../../zustand/useApplicationStore'
import { APPLICATION_PASS_STATUS_LABELS } from '../ApplicationParticipantsPage/components/table/getApplicationStatusBadge'
import { ApplicationStatusCard } from './diagrams/ApplicationStatusCard'
import { useHideMailingWarning } from './hooks/useHideMailingWarning'
import { useMissingConfigs } from './hooks/useMissingConfig'

export const ApplicationLandingPage = () => {
  const [applicationMetaData, setApplicationMetaData] = useState<ApplicationMetaData | null>(null)
  const { pathname } = useLocation()
  const { coursePhase, participations } = useApplicationStore()
  const { hideMailingWarning } = useHideMailingWarning()
  const studyPrograms = useStudyPrograms()
  const missingConfigs = useMissingConfigs(
    applicationMetaData,
    coursePhase,
    pathname,
    hideMailingWarning,
  )

  useParseApplicationMetaData(coursePhase, setApplicationMetaData)

  const isApplicationConfigured = getIsApplicationConfigured(applicationMetaData)
  const statusCard = (
    <ApplicationStatusCard
      applicationMetaData={applicationMetaData}
      applicationPhaseIsConfigured={isApplicationConfigured}
    />
  )

  return (
    <div>
      <div className='flex items-center justify-between gap-4'>
        <ManagementPageHeader>Application Administration</ManagementPageHeader>
        {coursePhase?.id && (
          <Button asChild variant='outline' className='mb-6'>
            <Link to={`/apply/${coursePhase.id}`} target='_blank' rel='noopener noreferrer'>
              <ExternalLink className='h-4 w-4' />
              Open Application Form
            </Link>
          </Button>
        )}
      </div>
      <MissingConfig elements={missingConfigs} />
      {studyPrograms.data ? (
        <ParticipationStatistics
          participations={participations}
          studyPrograms={studyPrograms.data}
          passStatusLabels={APPLICATION_PASS_STATUS_LABELS}
          scores={participations.flatMap(({ score }) => (score === null ? [] : [score]))}
          scoreTitle='Application Score'
        >
          {statusCard}
        </ParticipationStatistics>
      ) : (
        <div className='grid gap-6 md:grid-cols-2 xl:grid-cols-3'>
          {statusCard}
          {studyPrograms.isError ? (
            <Alert variant='destructive' className='xl:col-span-2'>
              <AlertTitle>Failed to load study programs.</AlertTitle>
              <AlertDescription>
                The participation statistics need the study programs. Reload the page to try again.
              </AlertDescription>
            </Alert>
          ) : (
            <Skeleton className='w-full min-h-[300px] xl:col-span-2' />
          )}
        </div>
      )}
    </div>
  )
}
