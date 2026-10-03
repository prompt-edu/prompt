import { useQuery } from '@tanstack/react-query'
import { useGetCoursePhaseParticipants } from '@tumaet/prompt-shared-state'
import {
  CoursePhaseParticipationsTable,
  ErrorPage,
  type ExtraParticipantColumn,
  LoadingPage,
  ManagementPageHeader,
  type ParticipantRow,
  type TableFilter,
} from '@tumaet/prompt-ui-components'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { ResourceStateBadge } from '../components/ResourceStateBadge'
import type { ResourceConfig } from '../interfaces/resourceConfig'
import { getInstances } from '../network/queries/getInstances'
import { getResourceConfigs } from '../network/queries/getResourceConfigs'
import { resourceLabel } from '../utils/resourceLabel'
import {
  participantResourceState,
  RESOURCE_STATE_LABELS,
  type ResourceState,
} from '../utils/resourceState'

const isRunning = (status: string) => status === 'pending' || status === 'in_progress'

const columnId = (config: ResourceConfig) => `resource-${config.id}`

// A column per config, named the way people say it. Two configs that read the same
// ("GitLab group" per team and per student) are told apart by scope, then by template.
const columnHeaders = (configs: ResourceConfig[]): Map<string, string> => {
  const count = (labels: string[], label: string) => labels.filter((l) => l === label).length
  const base = configs.map((c) => resourceLabel(c.providerType, c.resourceType))
  const scoped = configs.map((c, i) =>
    count(base, base[i]) > 1
      ? `${base[i]} (per ${c.scope === 'per_team' ? 'team' : 'student'})`
      : base[i],
  )
  return new Map(
    configs.map((c, i) => [
      c.id,
      count(scoped, scoped[i]) > 1 ? `${scoped[i]}: ${c.nameTemplate}` : scoped[i],
    ]),
  )
}

// The select filter of the table only matches accessor-key columns on its own; these
// columns derive their value, so they bring the same any-of match along.
const matchesAnyLabel = (
  row: { getValue: (id: string) => unknown },
  id: string,
  filterValue: unknown,
) =>
  !Array.isArray(filterValue) || filterValue.length === 0 || filterValue.includes(row.getValue(id))

export const ParticipantsPage = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const {
    data: coursePhaseParticipations,
    isPending: participationsPending,
    isError: participationsError,
    refetch: refetchParticipations,
  } = useGetCoursePhaseParticipants()

  const {
    data: configs,
    isPending: configsPending,
    isError: configsError,
    refetch: refetchConfigs,
  } = useQuery({
    queryKey: ['resource-configs', phaseId],
    queryFn: () => getResourceConfigs(phaseId ?? ''),
    enabled: !!phaseId,
  })

  const {
    data: instances,
    isPending: instancesPending,
    isError: instancesError,
    refetch: refetchInstances,
  } = useQuery({
    queryKey: ['instances', phaseId],
    queryFn: () => getInstances(phaseId ?? ''),
    enabled: !!phaseId,
    refetchInterval: (query) =>
      (query.state.data ?? []).some((i) => isRunning(i.status)) ? 3000 : false,
  })

  const participations = coursePhaseParticipations?.participations

  // state[participation][config], computed once per data change rather than per cell.
  const states = useMemo(() => {
    const byParticipation = new Map<string, Map<string, ResourceState>>()
    for (const participation of participations ?? []) {
      const perConfig = new Map<string, ResourceState>()
      for (const config of configs ?? []) {
        perConfig.set(
          config.id,
          participantResourceState(instances ?? [], config.id, participation.courseParticipationID),
        )
      }
      byParticipation.set(participation.courseParticipationID, perConfig)
    }
    return byParticipation
  }, [participations, configs, instances])

  const headers = useMemo(() => columnHeaders(configs ?? []), [configs])

  const extraColumns: ExtraParticipantColumn<string>[] = useMemo(
    () =>
      (configs ?? []).map((config) => {
        const id = columnId(config)
        return {
          id,
          header: headers.get(config.id) ?? config.id,
          accessorFn: (row: ParticipantRow) => row[id] as string,
          cell: ({ row }: { row: { original: ParticipantRow } }) => {
            const state =
              states.get(row.original.courseParticipationID)?.get(config.id) ?? 'not_provisioned'
            return <ResourceStateBadge state={state} label={RESOURCE_STATE_LABELS[state]} />
          },
          extraData: (participations ?? []).map((participation) => {
            const state =
              states.get(participation.courseParticipationID)?.get(config.id) ?? 'not_provisioned'
            return {
              courseParticipationID: participation.courseParticipationID,
              value: RESOURCE_STATE_LABELS[state],
              stringValue: RESOURCE_STATE_LABELS[state],
            }
          }),
          enableSorting: true,
          enableColumnFilter: true,
          filterFn: matchesAnyLabel,
        }
      }),
    [configs, headers, participations, states],
  )

  const extraFilters: TableFilter<ParticipantRow>[] = useMemo(
    () =>
      (configs ?? []).map((config) => {
        const header = headers.get(config.id) ?? config.id
        return {
          type: 'select' as const,
          id: columnId(config),
          label: header,
          options: Object.values(RESOURCE_STATE_LABELS),
          badge: { label: header, displayValue: (value: unknown) => String(value) },
        }
      }),
    [configs, headers],
  )

  const refetch = () => {
    refetchParticipations()
    refetchConfigs()
    refetchInstances()
  }

  if (participationsError || configsError || instancesError) {
    return (
      <ErrorPage
        onRetry={refetch}
        description='Could not load the participants or their resources.'
      />
    )
  }
  if (participationsPending || configsPending || instancesPending) {
    return <LoadingPage />
  }

  const total = participations?.length ?? 0
  const fullyReady = [...states.values()].filter(
    (perConfig) =>
      perConfig.size > 0 && [...perConfig.values()].every((state) => state === 'ready'),
  ).length

  return (
    <div className='space-y-4'>
      <div>
        <ManagementPageHeader>Participants</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Every participant of this phase, with what each resource means for them. To move students
          on, filter a resource by Ready, select them, and set them passed.
        </p>
      </div>

      {(configs ?? []).length > 0 && (
        <div className='rounded-lg bg-muted p-4 text-sm'>
          <span className='font-medium'>{fullyReady}</span> of{' '}
          <span className='font-medium'>{total}</span> participants can reach every resource.
        </div>
      )}

      <CoursePhaseParticipationsTable
        phaseId={phaseId!}
        participants={participations ?? []}
        extraColumns={extraColumns}
        extraFilters={extraFilters}
      />
    </div>
  )
}

export default ParticipantsPage
