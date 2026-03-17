import { useQuery } from '@tanstack/react-query'
import { getStudentEnrollments } from '@core/network/queries/getStudentEnrollments'

export const useStudentEnrollments = (studentId?: string) => {
  return useQuery({
    queryKey: ['studentEnrollments', studentId],
    queryFn: () => getStudentEnrollments(studentId!),
    enabled: !!studentId,
  })
}
