import { coreKeys } from '@core/network/cache'
import { getStudent } from '@core/network/queries/getStudent'
import { useQuery } from '@tanstack/react-query'

export const useStudent = (studentId?: string) => {
  return useQuery({
    queryKey: coreKeys.students.byId(studentId),
    queryFn: () => getStudent(studentId!),
    enabled: !!studentId,
  })
}
