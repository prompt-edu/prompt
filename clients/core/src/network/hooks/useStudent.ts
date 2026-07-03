import { getStudent } from '@core/network/queries/getStudent'
import { useQuery } from '@tanstack/react-query'

export const useStudent = (studentId?: string) => {
  return useQuery({
    queryKey: ['student', studentId],
    queryFn: () => getStudent(studentId!),
    enabled: !!studentId,
  })
}
