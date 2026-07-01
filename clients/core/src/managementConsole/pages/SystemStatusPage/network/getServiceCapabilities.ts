import axios from 'axios'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { ServiceInfo } from '../interfaces/serviceCapabilities'

export const getServiceInfo = async (service: CoursePhaseType): Promise<ServiceInfo> => {
  const response = await axios.get<ServiceInfo>(`${service.baseUrl}/info`)
  return response.data
}
