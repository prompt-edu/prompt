import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

export const useStudentEnrollments = (studentId?: string) => {
  return useQuery({
    queryKey: coreKeys.students.enrollments(studentId),
    queryFn: () => coreApi.students.enrollments(studentId!),
    enabled: !!studentId,
  })
}
