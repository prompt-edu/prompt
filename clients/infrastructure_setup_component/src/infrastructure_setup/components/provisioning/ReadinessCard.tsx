import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  cn,
  Skeleton,
  useToast,
} from '@tumaet/prompt-ui-components'
import { CheckCircle2, Circle, Loader2, Play, XCircle } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { ProviderType } from '../../interfaces/providerConfig'
import type { ResourceConfig } from '../../interfaces/resourceConfig'
import { describeTriggerSummary } from '../../interfaces/triggerSummary'
import { triggerExecution } from '../../network/mutations/triggerExecution'
import { getProviderConfigs } from '../../network/queries/getProviderConfigs'
import { getProvisioningPreview } from '../../network/queries/getProvisioningPreview'
import { getResourceConfigs } from '../../network/queries/getResourceConfigs'
import { getSetupConfig } from '../../network/queries/getSetupConfig'
import { describeError, hasStatus } from '../../utils/describeError'
import { providerName, resourceLabel } from '../../utils/resourceLabel'
import { SectionError } from '../SectionError'

type CheckState = 'ok' | 'blocked' | 'info' | 'loading'

interface Check {
  label: string
  state: CheckState
  detail: ReactNode
  fix?: { to: string; label: string }
}

interface Props {
  courseId: string
  coursePhaseID: string
  // Instances a run is still working on, from the polled instance list.
  running: number
}

// {{semesterTag}} and the aliases the server accepts for it.
const SEMESTER_TAG_PLACEHOLDER = /\{\{\s*\.?semester(tag)?\s*\}\}/i

const usesSemesterTag = (configs: ResourceConfig[]) =>
  configs.some(
    (config) =>
      SEMESTER_TAG_PLACEHOLDER.test(config.nameTemplate) ||
      SEMESTER_TAG_PLACEHOLDER.test(JSON.stringify(config.resourceExtraConfig ?? {})),
  )

const listNames = (types: ProviderType[]) => types.map(providerName).join(', ')

// Server errors are written like Go errors, lowercase and without a full stop.
const asSentence = (message: string) => {
  const text = message.charAt(0).toUpperCase() + message.slice(1)
  return /[.!?]$/.test(text) ? text : `${text}.`
}

const CheckIcon = ({ state }: { state: CheckState }) => {
  switch (state) {
    case 'ok':
      return <CheckCircle2 className='h-5 w-5 shrink-0 text-green-600' />
    case 'blocked':
      return <XCircle className='h-5 w-5 shrink-0 text-red-600' />
    case 'loading':
      return <Loader2 className='h-5 w-5 shrink-0 animate-spin text-muted-foreground' />
    case 'info':
      return <Circle className='h-5 w-5 shrink-0 text-muted-foreground' />
  }
}

export const ReadinessCard = ({ courseId, coursePhaseID, running }: Props) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const configurationPath = `/management/course/${courseId}/${coursePhaseID}/configuration`

  const providersQuery = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => getProviderConfigs(coursePhaseID),
  })
  const resourcesQuery = useQuery({
    queryKey: ['resource-configs', coursePhaseID],
    queryFn: () => getResourceConfigs(coursePhaseID),
  })
  const setupQuery = useQuery({
    queryKey: ['setup-config', coursePhaseID],
    queryFn: () => getSetupConfig(coursePhaseID),
  })
  // The phase's own checks come from data already loaded, so they are known before the
  // dry run is asked: the server would only refuse it for the same reasons.
  const providers = providersQuery.data ?? []
  const resources = resourcesQuery.data ?? []
  const semesterTag = setupQuery.data?.semesterTag ?? ''

  const configured = providers.filter((p) => p.configured).map((p) => p.providerType)
  const needed = [...new Set(resources.map((r) => r.providerType))]
  const missingCredentials = needed.filter((type) => !configured.includes(type))

  const providerCheck: Check =
    providers.length === 0
      ? {
          label: 'Providers',
          state: 'blocked',
          detail: 'No provider is configured yet.',
          fix: { to: `${configurationPath}#providers`, label: 'Add a provider' },
        }
      : missingCredentials.length > 0
        ? {
            label: 'Providers',
            state: 'blocked',
            detail: `${listNames(missingCredentials)} ${missingCredentials.length === 1 ? 'has' : 'have'} no credentials.`,
            fix: { to: `${configurationPath}#providers`, label: 'Enter credentials' },
          }
        : {
            label: 'Providers',
            state: 'ok',
            detail: `${listNames(configured)} ${configured.length === 1 ? 'has' : 'have'} credentials.`,
          }

  const resourceCheck: Check =
    resources.length === 0
      ? {
          label: 'Resources',
          state: 'blocked',
          detail: 'Nothing is configured to be created yet.',
          fix: { to: `${configurationPath}#resources`, label: 'Add a resource' },
        }
      : {
          label: 'Resources',
          state: 'ok',
          detail: resources
            .map(
              (r) =>
                `${resourceLabel(r.providerType, r.resourceType)} per ${r.scope === 'per_team' ? 'team' : 'student'}`,
            )
            .join(', '),
        }

  const semesterTagCheck: Check =
    semesterTag !== ''
      ? { label: 'Semester tag', state: 'ok', detail: semesterTag }
      : usesSemesterTag(resources)
        ? {
            label: 'Semester tag',
            state: 'blocked',
            detail: (
              <>
                A template uses <code>{`{{semesterTag}}`}</code>, but no tag is saved.
              </>
            ),
            fix: { to: `${configurationPath}#semester-tag`, label: 'Set the semester tag' },
          }
        : { label: 'Semester tag', state: 'info', detail: 'Not set. No template uses it.' }

  const ownChecksPass = [providerCheck, resourceCheck, semesterTagCheck].every(
    (check) => check.state !== 'blocked',
  )

  // The server's answer to "what would a trigger do now". It resolves the teams and
  // students through core, so it also catches what the phase's own data cannot show,
  // such as a phase no teams reach. A 400 is an answer, not a failure worth retrying.
  const previewQuery = useQuery({
    queryKey: ['provisioning-preview', coursePhaseID],
    queryFn: () => getProvisioningPreview(coursePhaseID),
    enabled: !!providersQuery.data && !!resourcesQuery.data && !!setupQuery.data && ownChecksPass,
    retry: (count, err) => !hasStatus(err, 400) && count < 2,
  })

  const { mutate: provision, isPending: isProvisioning } = useMutation({
    mutationFn: () => triggerExecution(coursePhaseID),
    onSuccess: (triggered) => {
      const startedWork = triggered.queued + triggered.requeued > 0
      toast({
        title: startedWork ? 'Provisioning started' : 'Nothing left to provision',
        description: describeTriggerSummary(triggered),
      })
    },
    onError: (err: unknown) => {
      toast({
        title: hasStatus(err, 409)
          ? 'A run is already in progress'
          : 'Failed to start provisioning',
        description: describeError(err),
        variant: 'destructive',
      })
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      queryClient.invalidateQueries({ queryKey: ['provisioning-preview', coursePhaseID] })
    },
  })

  if (providersQuery.isLoading || resourcesQuery.isLoading || setupQuery.isLoading) {
    return <Skeleton className='h-64 w-full' />
  }
  if (providersQuery.isError || resourcesQuery.isError || setupQuery.isError) {
    return (
      <SectionError
        message='Failed to load the configuration of this phase.'
        onRetry={() => {
          providersQuery.refetch()
          resourcesQuery.refetch()
          setupQuery.refetch()
        }}
      />
    )
  }

  const preview = previewQuery.data

  // The server refuses for the same reasons the items above report, so until they are
  // fixed its message would only repeat one of them.
  const targetCheck: Check = !ownChecksPass
    ? {
        label: 'Teams and students',
        state: 'info',
        detail: 'Resolved once the items above are fixed.',
      }
    : previewQuery.isLoading
      ? { label: 'Teams and students', state: 'loading', detail: 'Resolving them through core…' }
      : previewQuery.isError
        ? {
            label: 'Teams and students',
            state: 'blocked',
            detail: asSentence(describeError(previewQuery.error)),
            // The phase's own settings are fine, so a refusal is about its place in the
            // course: no teams reach it through the phase graph.
            fix: hasStatus(previewQuery.error, 400)
              ? {
                  to: `/management/course/${courseId}/configurator`,
                  label: 'Open the course configurator',
                }
              : undefined,
          }
        : {
            label: 'Teams and students',
            state: 'ok',
            detail: [
              preview?.teams != null &&
                `${preview.teams} ${preview.teams === 1 ? 'team' : 'teams'}`,
              preview?.students != null &&
                `${preview.students} ${preview.students === 1 ? 'student' : 'students'}`,
            ]
              .filter(Boolean)
              .join(', '),
          }

  const checks = [providerCheck, resourceCheck, semesterTagCheck, targetCheck]
  const blocked = checks.some((check) => check.state === 'blocked')
  const work = preview ? preview.queued + preview.requeued : 0

  let summary: ReactNode
  if (running > 0) {
    summary = `A run is in progress: ${running} ${running === 1 ? 'resource is' : 'resources are'} still being set up. You can start the next run when it finishes.`
  } else if (blocked) {
    summary = 'Fix the items above to provision.'
  } else if (!preview) {
    summary = 'Working out what the next run does…'
  } else if (work === 0) {
    summary = 'Everything is provisioned. The next run has nothing to do.'
  } else {
    const parts = [
      preview.queued > 0 && `creates ${preview.queued}`,
      preview.requeued > 0 && `retries ${preview.requeued} that failed or left members out`,
      preview.upToDate > 0 &&
        `leaves ${preview.upToDate} as ${preview.upToDate === 1 ? 'it is' : 'they are'}`,
    ].filter(Boolean)
    summary = `The next run ${parts.join(', ')}.`
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{blocked ? 'Not ready to provision' : 'Ready to provision'}</CardTitle>
        <CardDescription>
          Provisioning creates every configured resource for every team and student of this phase,
          and adds the members. Nothing is ever deleted.
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        <ul className='space-y-3'>
          {checks.map((check) => (
            <li key={check.label} className='flex items-start gap-3'>
              <CheckIcon state={check.state} />
              <div className='min-w-0 flex-1 text-sm'>
                <span className='font-medium'>{check.label}</span>
                <span
                  className={cn(
                    'ml-2',
                    check.state === 'blocked' ? 'text-red-700' : 'text-muted-foreground',
                  )}
                >
                  {check.detail}
                </span>
                {check.fix && (
                  <Link to={check.fix.to} className='ml-2 text-blue-600 hover:underline'>
                    {check.fix.label}
                  </Link>
                )}
              </div>
            </li>
          ))}
        </ul>

        <div className='flex flex-wrap items-center justify-between gap-4 border-t pt-4'>
          <p className='text-sm'>{summary}</p>
          <Button
            onClick={() => provision()}
            disabled={isProvisioning || blocked || running > 0 || !preview || work === 0}
          >
            {isProvisioning ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Play className='mr-2 h-4 w-4' />
            )}
            {isProvisioning ? 'Starting…' : 'Provision resources'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
