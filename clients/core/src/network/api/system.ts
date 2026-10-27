import type {
  ClientInfo,
  ClientRemote,
} from '@core/managementConsole/pages/SystemStatusPage/interfaces/clientInfo'
import type { CoursePhaseType } from '@core/managementConsole/pages/SystemStatusPage/interfaces/coursePhaseType'
import type { ServiceInfo } from '@core/managementConsole/pages/SystemStatusPage/interfaces/serviceCapabilities'
import axios from 'axios'
import { API_PREFIX, coreRequest } from '../client'

const CLIENT_PROBE_TIMEOUT_MS = 5_000

const fetchUncached = (url: string, method: 'GET' | 'HEAD' = 'GET') =>
  fetch(url, { method, cache: 'no-store', signal: AbortSignal.timeout(CLIENT_PROBE_TIMEOUT_MS) })

const readBuildVersion = async (remoteUrl: string): Promise<string | undefined> => {
  try {
    const response = await fetchUncached(`${remoteUrl}/mf-manifest.json`)
    if (!response.ok) {
      return undefined
    }
    const manifest = await response.json()
    return manifest?.metaData?.buildInfo?.buildVersion
  } catch {
    return undefined
  }
}

export const system = {
  /** Reached only to see whether core answers at all, so the body is not read. */
  coreInfo: (): Promise<unknown> => coreRequest.get(`${API_PREFIX}/hello`),

  /**
   * Each phase type names the base URL of the microservice behind it, a host that is not core's
   * own, and the probe is unauthenticated, so it goes through plain axios rather than either shared
   * instance.
   */
  serviceInfo: async (service: CoursePhaseType): Promise<ServiceInfo> =>
    (await axios.get<ServiceInfo>(`${service.baseUrl}/info`)).data,

  clientInfo: async (remote: ClientRemote): Promise<ClientInfo> => {
    const [entry, buildVersion] = await Promise.all([
      fetchUncached(`${remote.url}/remoteEntry.js`, 'HEAD'),
      readBuildVersion(remote.url),
    ])
    if (!entry.ok) {
      throw new Error(`${remote.name}: remoteEntry.js answered ${entry.status}`)
    }
    return { buildVersion }
  },
}
