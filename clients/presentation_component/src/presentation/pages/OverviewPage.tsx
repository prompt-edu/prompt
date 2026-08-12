import { useQuery } from '@tanstack/react-query'
import {
  Alert,
  AlertDescription,
  AlertTitle,
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
import { CalendarClock, CalendarCog, MapPin, MessageSquareText, TriangleAlert } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { MaterialsPanel } from '../components/MaterialsPanel'
import { useCoursePhaseId, usePresentationAccess } from '../hooks'
import type { PresentationMaterial, PresentationSummary } from '../interfaces'
import { presentationApi } from '../network'
import { buildSampleMaterials, buildSamplePresentation } from '../samplePresentation'
import { formatDateTime, getApiError } from '../utils'

interface PresentationCardProps {
  presentation: PresentationSummary
  onOpenFeedback?: () => void
}

const PresentationCard = ({ presentation, onOpenFeedback }: PresentationCardProps) => (
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
      {onOpenFeedback ? (
        <Button onClick={onOpenFeedback}>
          <MessageSquareText className='mr-2 h-4 w-4' />
          View feedback
        </Button>
      ) : null}
    </CardHeader>
  </Card>
)

interface StudentViewProps {
  coursePhaseId: string
  presentation: PresentationSummary
  previewMaterials?: PresentationMaterial[]
  onOpenFeedback?: () => void
}

const StudentView = ({
  coursePhaseId,
  presentation,
  previewMaterials,
  onOpenFeedback,
}: StudentViewProps) => (
  <div className='space-y-4'>
    <PresentationCard presentation={presentation} onOpenFeedback={onOpenFeedback} />
    <MaterialsPanel
      coursePhaseId={coursePhaseId}
      presentation={presentation}
      isStaff={false}
      previewMaterials={previewMaterials}
    />
  </div>
)

// Everything on this page belongs to a presenter, so anybody else gets the same layout filled
// with sample data instead of an error. It mirrors the assessment evaluation pages, where
// instructors also see a disabled preview of the student experience.
const NonStudentPreview = ({
  coursePhaseId,
  canManagePhase,
}: {
  coursePhaseId: string
  canManagePhase: boolean
}) => {
  const navigate = useNavigate()
  const configQuery = useQuery({
    queryKey: ['presentation-config', coursePhaseId],
    queryFn: () => presentationApi.getConfig(coursePhaseId),
    enabled: Boolean(coursePhaseId),
  })

  if (configQuery.isLoading) return <LoadingPage />
  // The preview mirrors the phase configuration, so a genuinely unreachable phase API is still
  // reported rather than illustrated with defaults.
  if (configQuery.isError) {
    return (
      <ErrorPage
        message='The presentation phase could not be loaded.'
        onRetry={() => void configQuery.refetch()}
      />
    )
  }

  const targetMode = configQuery.data?.targetMode ?? 'individual'
  const presentation = buildSamplePresentation(coursePhaseId, targetMode)
  const materials = buildSampleMaterials(configQuery.data?.requiredMaterialTypes ?? [])

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>Presentations</ManagementPageHeader>
        <p className='text-muted-foreground'>
          The page presenters see for their own slot, their materials, and released feedback.
        </p>
      </div>

      <Alert>
        <TriangleAlert className='h-4 w-4' />
        <AlertTitle>You are not a student of this course.</AlertTitle>
        <AlertDescription className='space-y-3'>
          <p>
            The view below is filled with sample data and is disabled, to demonstrate what
            presenters get to see. Slots, presenter assignments, uploaded materials, and instructor
            feedback are managed from the schedule page.
          </p>
          {canManagePhase ? (
            <Button variant='outline' size='sm' onClick={() => navigate('schedule')}>
              <CalendarCog className='mr-2 h-4 w-4' />
              Manage schedule
            </Button>
          ) : null}
        </AlertDescription>
      </Alert>

      <div className='pointer-events-none select-none opacity-70'>
        <StudentView
          coursePhaseId={coursePhaseId}
          presentation={presentation}
          previewMaterials={materials}
        />
      </div>
    </div>
  )
}

const OverviewPage = () => {
  const coursePhaseId = useCoursePhaseId()
  const { canManagePhase, isStudent } = usePresentationAccess()
  const navigate = useNavigate()

  const presentationQuery = useQuery({
    queryKey: ['presentations', coursePhaseId, 'own'],
    queryFn: () => presentationApi.getOwnPresentation(coursePhaseId),
    enabled: isStudent && Boolean(coursePhaseId),
  })

  if (!isStudent)
    return <NonStudentPreview coursePhaseId={coursePhaseId} canManagePhase={canManagePhase} />

  if (presentationQuery.isLoading) return <LoadingPage />
  if (presentationQuery.isError) {
    // An unconnected team allocation is a phase misconfiguration the student cannot fix
    // and retrying will not resolve, so it gets an explanation rather than a retry page.
    if (getApiError(presentationQuery.error).code === 'team_not_resolved') {
      return (
        <ManagementPageHeader>
          This phase is set to team presentations, but no team allocation is connected to it. Please
          ask your lecturer to finish configuring the phase.
        </ManagementPageHeader>
      )
    }
    return (
      <ErrorPage
        message='Your presentation could not be loaded.'
        onRetry={() => void presentationQuery.refetch()}
      />
    )
  }

  const presentation = presentationQuery.data ?? undefined

  return (
    <div className='space-y-6'>
      <div>
        <ManagementPageHeader>My presentation</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Find your presentation time and hand in the materials your instructors ask for.
        </p>
      </div>

      {presentation ? (
        <StudentView
          coursePhaseId={coursePhaseId}
          presentation={presentation}
          onOpenFeedback={
            presentation.feedbackReleasedAt
              ? () => navigate(`presentations/${presentation.id}`)
              : undefined
          }
        />
      ) : (
        <Card>
          <CardContent className='flex flex-col items-center gap-3 py-12 text-center'>
            <CalendarClock className='h-10 w-10 text-muted-foreground' />
            <div>
              <p className='font-medium'>No presentation is assigned yet</p>
              <p className='text-sm text-muted-foreground'>
                Your instructors will publish your presentation slot here.
              </p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

export default OverviewPage
