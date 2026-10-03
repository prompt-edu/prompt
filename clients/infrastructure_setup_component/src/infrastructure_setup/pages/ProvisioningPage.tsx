import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { InstanceList } from '../components/provisioning/InstanceList'
import { ReadinessCard } from '../components/provisioning/ReadinessCard'
import { getInstances } from '../network/queries/getInstances'

const isRunning = (status: string) => status === 'pending' || status === 'in_progress'

export const ProvisioningPage = () => {
  const { courseId, phaseId } = useParams<{ courseId: string; phaseId: string }>()
  const queryClient = useQueryClient()

  const {
    data: instances,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['instances', phaseId],
    queryFn: () => getInstances(phaseId ?? ''),
    enabled: !!phaseId,
    refetchInterval: (query) =>
      (query.state.data ?? []).some((i) => isRunning(i.status)) ? 3000 : false,
  })

  const running = (instances ?? []).filter((i) => isRunning(i.status)).length

  // What the next run would do changes when a run ends, since its failures become
  // retries. The preview is fetched through core, so it is refreshed then rather than
  // polled alongside the instances.
  const wasRunning = useRef(running > 0)
  useEffect(() => {
    if (wasRunning.current && running === 0) {
      queryClient.invalidateQueries({ queryKey: ['provisioning-preview', phaseId] })
    }
    wasRunning.current = running > 0
  }, [running, phaseId, queryClient])

  if (!courseId || !phaseId) return null

  return (
    <div className='max-w-5xl space-y-8'>
      <div>
        <ManagementPageHeader>Provisioning</ManagementPageHeader>
        <p className='text-muted-foreground'>
          Create the configured resources for every team and student of this phase, and follow their
          progress.
        </p>
      </div>
      <ReadinessCard courseId={courseId} coursePhaseID={phaseId} running={running} />
      <InstanceList
        coursePhaseID={phaseId}
        instances={instances}
        isLoading={isLoading}
        isError={isError}
        onRefresh={() => refetch()}
      />
    </div>
  )
}

export default ProvisioningPage
