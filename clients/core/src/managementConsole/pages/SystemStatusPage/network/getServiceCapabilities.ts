import axios from 'axios'
import { parseURL } from '@/utils/parseURL'
import { KnownService, ServiceInfo } from '../interfaces/serviceCapabilities'

export const getServiceInfo = async (service: KnownService): Promise<ServiceInfo> => {
  const baseUrl = parseURL(service.host)
  const response = await axios.get<ServiceInfo>(`${baseUrl}/${service.apiBasePath}/info`)
  return response.data
}
