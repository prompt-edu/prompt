import { getStudentEnrollments } from '@core/network/queries/getStudentEnrollments'
import { useQuery } from '@tanstack/react-query'

export const useStudentEnrollments = (studentId?: string) => {
  return useQuery({
    queryKey: ['studentEnrollments', studentId],
    queryFn: () => getStudentEnrollments(studentId!),
    enabled: !!studentId,
  })
}
