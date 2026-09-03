import type { CoursePhaseType } from '@core/managementConsole/pages/SystemStatusPage/interfaces/coursePhaseType'
import type { ServiceInfo } from '@core/managementConsole/pages/SystemStatusPage/interfaces/serviceCapabilities'
import axios from 'axios'
import { API_PREFIX, coreRequest } from '../client'

export const system = {
  /** Reached only to see whether core answers at all, so the body is not read. */
  coreInfo: (): Promise<unknown> => coreRequest.get(`${API_PREFIX}/hello`),

  /**
   * The only request core makes to a host that is not its own: each phase type names the base URL
   * of the microservice behind it, and the probe is unauthenticated, so it goes through plain axios
   * rather than either shared instance.
   */
  serviceInfo: async (service: CoursePhaseType): Promise<ServiceInfo> =>
    (await axios.get<ServiceInfo>(`${service.baseUrl}/info`)).data,
}
