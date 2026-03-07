import { useQuery } from '@tanstack/react-query'
import { KnownService, ServiceInfo } from '../interfaces/serviceCapabilities'
import { getServiceInfo } from '../network/getServiceCapabilities'

export const useGetServiceInfo = (service: KnownService) => {
  return useQuery<ServiceInfo, Error>({
    queryKey: ['serviceInfo', service.apiBasePath],
    queryFn: () => getServiceInfo(service),
    retry: false,
    staleTime: 30_000,
  })
}
