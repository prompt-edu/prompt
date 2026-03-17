import { useQuery } from '@tanstack/react-query'
import { getStudent } from '@core/network/queries/getStudent'

export const useStudent = (studentId?: string) => {
  return useQuery({
    queryKey: ['student', studentId],
    queryFn: () => getStudent(studentId!),
    enabled: !!studentId,
  })
}
