import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { getDeveloperProfile } from '../../network/queries/getDeveloperProfile'
import { DeveloperProfile } from '../../interfaces/DeveloperProfile'

export const useGetDeveloperProfile = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<DeveloperProfile | undefined>({
    queryKey: ['devOpsDeveloperProfile', phaseId],
    queryFn: () => getDeveloperProfile(phaseId ?? ''),
  })
}
