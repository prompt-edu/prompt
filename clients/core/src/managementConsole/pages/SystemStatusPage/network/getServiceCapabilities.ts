import axios from 'axios'
import { ServiceInfo } from '../interfaces/serviceCapabilities'
import { CoursePhaseType } from '../interfaces/coursePhaseType'

export const getServiceInfo = async (service: CoursePhaseType): Promise<ServiceInfo> => {
  const response = await axios.get<ServiceInfo>(`${service.baseUrl}/info`)
  return response.data
}
