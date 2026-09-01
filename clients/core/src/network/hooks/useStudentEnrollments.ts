import { coreKeys } from '@core/network/cache'
import { getStudentEnrollments } from '@core/network/queries/getStudentEnrollments'
import { useQuery } from '@tanstack/react-query'

export const useStudentEnrollments = (studentId?: string) => {
  return useQuery({
    queryKey: coreKeys.students.enrollments(studentId),
    queryFn: () => getStudentEnrollments(studentId!),
    enabled: !!studentId,
  })
}
