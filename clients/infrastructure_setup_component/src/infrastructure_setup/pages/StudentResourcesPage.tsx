import { useQuery } from '@tanstack/react-query'
import {
  getPermissionString,
  Role,
  useAuthStore,
  useCourseStore,
} from '@tumaet/prompt-shared-state'
import {
  Alert,
  AlertDescription,
  AlertTitle,
  Button,
  Card,
  CardContent,
  ErrorPage,
  LoadingPage,
} from '@tumaet/prompt-ui-components'
import { ExternalLink, Server, TriangleAlert } from 'lucide-react'
import { useParams } from 'react-router-dom'
import { ResourceStateBadge } from '../components/ResourceStateBadge'
import type { MyResource } from '../interfaces/myResource'
import type { ResourceConfig } from '../interfaces/resourceConfig'
import { getMyResources } from '../network/queries/getMyResources'
import { getResourceConfigs } from '../network/queries/getResourceConfigs'
import { providerName, resourceLabel } from '../utils/resourceLabel'
import {
  type ResourceState,
  resourceState,
  STUDENT_RESOURCE_STATE_LABELS,
} from '../utils/resourceState'

const isRunning = (status: string | null) => status === 'pending' || status === 'in_progress'

// What the student can do about a state, or why there is nothing to do.
const hintFor = (state: ResourceState, resource: MyResource): string | undefined => {
  switch (state) {
    case 'not_provisioned':
      return 'Your instructors have not set this up yet.'
    case 'setting_up':
      return 'This is being created right now.'
    case 'not_added':
      return `You could not be added yet. Often you have to sign in to ${providerName(resource.providerType)} once first. Then let your instructors know, so they can add you.`
    case 'partly_ready':
      return 'If you cannot open it, let your instructors know.'
    case 'failed':
      return 'Setting this up ran into a problem. Your instructors can see it and will retry.'
    case 'ready':
      return resource.url
        ? undefined
        : `There is nothing to open here: this grants you access where you sign in with ${providerName(resource.providerType)}.`
  }
}

// A lecturer sees the entries a student gets, before anything is provisioned for them.
const previewOf = (config: ResourceConfig): MyResource => ({
  resourceConfigId: config.id,
  providerType: config.providerType,
  resourceType: config.resourceType,
  scope: config.scope,
  status: null,
  granted: null,
  name: '',
  teamName: '',
  url: null,
})

const ResourceEntry = ({ resource }: { resource: MyResource }) => {
  const state = resourceState(resource.status, resource.granted)
  const owner = resource.scope === 'per_team' ? resource.teamName || 'your team' : 'just for you'
  const hint = hintFor(state, resource)

  return (
    <Card>
      <CardContent className='flex flex-wrap items-start justify-between gap-4 p-4'>
        <div className='min-w-0 space-y-1'>
          <p className='font-medium'>
            {resourceLabel(resource.providerType, resource.resourceType)}
            <span className='ml-2 font-normal text-muted-foreground'>{owner}</span>
          </p>
          {resource.name && <p className='font-mono text-sm'>{resource.name}</p>}
          {hint && <p className='text-sm text-muted-foreground'>{hint}</p>}
        </div>
        <div className='flex items-center gap-2'>
          <ResourceStateBadge state={state} label={STUDENT_RESOURCE_STATE_LABELS[state]} />
          {resource.url && (state === 'ready' || state === 'partly_ready') && (
            <Button asChild variant='outline' size='sm'>
              <a href={resource.url} target='_blank' rel='noopener noreferrer'>
                <ExternalLink className='mr-1 h-3 w-3' />
                Open
              </a>
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export const StudentResourcesPage = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const { courses, isStudentOfCourse } = useCourseStore()
  const { permissions } = useAuthStore()

  const course = courses.find((c) => c.id === courseId)
  const isStudent = isStudentOfCourse(courseId ?? '')
  // The resource configs are lecturer data: an editor gets the notice without a preview.
  const canPreview = [Role.PROMPT_ADMIN, Role.COURSE_LECTURER].some((role) =>
    permissions.includes(getPermissionString(role, course?.name, course?.semesterTag)),
  )

  const myResources = useQuery({
    queryKey: ['my-resources', phaseId],
    queryFn: () => getMyResources(phaseId ?? ''),
    enabled: !!phaseId && isStudent,
    // While something is being set up, the page follows it to Ready on its own.
    refetchInterval: (query) =>
      (query.state.data ?? []).some((r) => isRunning(r.status)) ? 5000 : false,
  })

  const preview = useQuery({
    queryKey: ['resource-configs', phaseId],
    queryFn: () => getResourceConfigs(phaseId ?? ''),
    enabled: !!phaseId && !isStudent && canPreview,
  })

  const shown = isStudent ? myResources : preview
  if (shown.isLoading) {
    return <LoadingPage />
  }
  if (shown.isError) {
    return (
      <ErrorPage description='Failed to load your resources.' onRetry={() => shown.refetch()} />
    )
  }

  const resources = isStudent ? (myResources.data ?? []) : (preview.data ?? []).map(previewOf)

  return (
    <div className='mx-auto max-w-3xl space-y-6'>
      {!isStudent && (
        <Alert>
          <TriangleAlert className='h-4 w-4' />
          <AlertTitle>You are not a student of this course.</AlertTitle>
          <AlertDescription>
            Students see the resources provisioned for them and their team here.
            {canPreview
              ? ' Below is what they get from the current configuration, before anything is provisioned.'
              : ''}
          </AlertDescription>
        </Alert>
      )}

      <div className='space-y-2'>
        <div className='flex items-center gap-2'>
          <Server className='h-6 w-6 text-blue-500' />
          <h1 className='text-2xl font-semibold'>Infrastructure Setup</h1>
        </div>
        <p className='text-muted-foreground'>
          The tools your course set up for you and your team. Open them from here once they are
          ready.
        </p>
      </div>

      {(isStudent || canPreview) &&
        (resources.length === 0 ? (
          <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
            There is nothing to set up for you in this phase yet.
          </div>
        ) : (
          <div className='space-y-2'>
            {resources.map((resource) => (
              <ResourceEntry
                key={`${resource.resourceConfigId}-${resource.teamName}-${resource.name}`}
                resource={resource}
              />
            ))}
          </div>
        ))}
    </div>
  )
}

export default StudentResourcesPage
