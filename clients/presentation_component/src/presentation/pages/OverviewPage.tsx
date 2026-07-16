import { useQuery } from '@tanstack/react-query'
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  ErrorPage,
  LoadingPage,
  ManagementPageHeader,
} from '@tumaet/prompt-ui-components'
import { CalendarClock, MapPin, MessageSquareText, Users } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { MaterialsPanel } from '../components/MaterialsPanel'
import { useCoursePhaseId, usePresentationAccess } from '../hooks'
import type { PresentationSummary } from '../interfaces'
import { presentationApi } from '../network'
import { formatDateTime } from '../utils'

interface PresentationCardProps {
  coursePhaseId: string
  presentation: PresentationSummary
  isStaff: boolean
  onOpenFeedback: () => void
}

const PresentationCard = ({
  coursePhaseId,
  presentation,
  isStaff,
  onOpenFeedback,
}: PresentationCardProps) => (
  <div className='space-y-4'>
    <Card>
      <CardHeader className='flex-row items-start justify-between gap-4'>
        <div>
          <div className='mb-2 flex flex-wrap items-center gap-2'>
            <Badge variant='outline'>
              {presentation.targetType === 'team' ? 'Team' : 'Individual'}
            </Badge>
            {presentation.feedbackReleasedAt ? <Badge>Feedback released</Badge> : null}
          </div>
          <CardTitle>{presentation.targetName}</CardTitle>
          <CardDescription className='mt-2 space-y-1'>
            <span className='flex items-center gap-2'>
              <CalendarClock className='h-4 w-4' />
              {formatDateTime(presentation.startTime)} –{' '}
              {new Date(presentation.endTime).toLocaleTimeString([], {
                hour: '2-digit',
                minute: '2-digit',
              })}
            </span>
            {presentation.location ? (
              <span className='flex items-center gap-2'>
                <MapPin className='h-4 w-4' />
                {presentation.location}
              </span>
            ) : null}
          </CardDescription>
        </div>
        {isStaff || presentation.feedbackReleasedAt ? (
          <Button onClick={onOpenFeedback}>
            <MessageSquareText className='mr-2 h-4 w-4' />
            {isStaff ? 'Open feedback' : 'View feedback'}
          </Button>
        ) : null}
      </CardHeader>
      {isStaff ? (
        <CardContent className='flex flex-wrap gap-4 text-sm text-muted-foreground'>
          <span>{presentation.materialCount ?? 0} materials</span>
          <span>{presentation.submittedFeedbackCount ?? 0} submitted evaluations</span>
          {presentation.feedbackReleasedByName ? (
            <span>Released by {presentation.feedbackReleasedByName}</span>
          ) : null}
          {presentation.feedbackReleaseName ? (
            <span>Release: {presentation.feedbackReleaseName}</span>
          ) : null}
        </CardContent>
      ) : null}
    </Card>
    <MaterialsPanel coursePhaseId={coursePhaseId} presentation={presentation} isStaff={isStaff} />
  </div>
)

const OverviewPage = () => {
  const coursePhaseId = useCoursePhaseId()
  const { isStaff } = usePresentationAccess()
  const navigate = useNavigate()

  const presentationsQuery = useQuery({
    queryKey: ['presentations', coursePhaseId, isStaff ? 'all' : 'own'],
    queryFn: () =>
      isStaff
        ? presentationApi.getPresentations(coursePhaseId)
        : presentationApi
            .getOwnPresentation(coursePhaseId)
            .then((presentation) => (presentation ? [presentation] : [])),
    enabled: Boolean(coursePhaseId),
  })

  if (presentationsQuery.isLoading) return <LoadingPage />
  if (presentationsQuery.isError) {
    return (
      <ErrorPage
        message='Presentations could not be loaded.'
        onRetry={() => void presentationsQuery.refetch()}
      />
    )
  }

  const presentations = presentationsQuery.data ?? []

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>{isStaff ? 'Presentations' : 'My presentation'}</ManagementPageHeader>
        <p className='text-muted-foreground'>
          {isStaff
            ? 'Review scheduled presentations, materials, and instructor feedback.'
            : 'Find your presentation time and submit slides or supporting materials.'}
        </p>
      </div>

      {presentations.length === 0 ? (
        <Card>
          <CardContent className='flex flex-col items-center gap-3 py-12 text-center'>
            <Users className='h-10 w-10 text-muted-foreground' />
            <div>
              <p className='font-medium'>No presentation is assigned yet</p>
              <p className='text-sm text-muted-foreground'>
                {isStaff
                  ? 'Create slots and assign presenters from the schedule page.'
                  : 'Your instructors will publish your presentation slot here.'}
              </p>
            </div>
            {isStaff ? (
              <Button variant='outline' onClick={() => navigate('schedule')}>
                Manage schedule
              </Button>
            ) : null}
          </CardContent>
        </Card>
      ) : (
        <div className='space-y-8'>
          {presentations.map((presentation) => (
            <PresentationCard
              key={presentation.id}
              coursePhaseId={coursePhaseId}
              presentation={presentation}
              isStaff={isStaff}
              onOpenFeedback={() => navigate(`presentations/${presentation.id}`)}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default OverviewPage
