import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

export const useStudent = (studentId?: string) => {
  return useQuery({
    queryKey: coreKeys.students.byId(studentId),
    queryFn: () => coreApi.students.byID(studentId!),
    enabled: !!studentId,
  })
}
