import type { ClientRemote } from '../interfaces/clientInfo'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'
import type { ServiceInfo } from '../interfaces/serviceCapabilities'
import { isRealMicroservice } from './isRealMicroservice'

export type ServiceStatus = 'Offline' | 'OnlineUnhealthy' | 'Online'

export interface PhaseStatusEntry {
  phaseType: CoursePhaseType
  hasServerInfo: boolean
  remote?: ClientRemote
}

export interface ProbeResult<T> {
  data: T | undefined
  isPending: boolean
  isError: boolean
}

export function buildPhaseStatusEntries(
  phaseTypes: CoursePhaseType[],
  remotes: ClientRemote[],
): PhaseStatusEntry[] {
  return phaseTypes
    .map((phaseType) => ({
      phaseType,
      hasServerInfo: isRealMicroservice(phaseType.baseUrl),
      remote: remotes.find((remote) => remote.phaseTypeName === phaseType.name),
    }))
    .filter((entry) => entry.hasServerInfo || entry.remote !== undefined)
    .sort((a, b) => a.phaseType.name.localeCompare(b.phaseType.name))
}

export function serverStatus({ data, isError }: ProbeResult<ServiceInfo>): ServiceStatus {
  if (isError || data == null) {
    return 'Offline'
  }
  return data.healthy ? 'Online' : 'OnlineUnhealthy'
}

export function clientStatus({ data, isError }: ProbeResult<unknown>): ServiceStatus {
  return isError || data == null ? 'Offline' : 'Online'
}

export function phaseCardStatus(
  server: ProbeResult<ServiceInfo> | undefined,
  client: ProbeResult<unknown> | undefined,
): ServiceStatus | undefined {
  if (server?.isPending || client?.isPending) {
    return undefined
  }
  const statuses = [server && serverStatus(server), client && clientStatus(client)].filter(
    (status) => status !== undefined,
  )
  if (statuses.every((status) => status === 'Offline')) {
    return 'Offline'
  }
  return statuses.every((status) => status === 'Online') ? 'Online' : 'OnlineUnhealthy'
}
