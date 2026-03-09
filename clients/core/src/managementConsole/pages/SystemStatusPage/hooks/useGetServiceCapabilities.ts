import { useQuery } from '@tanstack/react-query'
import { ServiceInfo } from '../interfaces/serviceCapabilities'
import { getServiceInfo } from '../network/getServiceCapabilities'
import { CoursePhaseType } from '../interfaces/coursePhaseType'

export const useGetServiceInfo = (coursePhaseType: CoursePhaseType) => {
  return useQuery<ServiceInfo, Error>({
    queryKey: ['serviceInfo', coursePhaseType.baseUrl],
    queryFn: () => getServiceInfo(coursePhaseType),
    retry: false,
    staleTime: 30_000,
  })
}
