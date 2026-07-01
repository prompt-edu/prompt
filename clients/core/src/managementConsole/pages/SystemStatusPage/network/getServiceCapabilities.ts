import axios from 'axios'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'
import type { ServiceInfo } from '../interfaces/serviceCapabilities'

export const getServiceInfo = async (service: CoursePhaseType): Promise<ServiceInfo> => {
  const response = await axios.get<ServiceInfo>(`${service.baseUrl}/info`)
  return response.data
}
