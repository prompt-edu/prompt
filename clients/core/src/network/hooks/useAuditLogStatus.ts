import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'
import { getAuditLogStatus } from '../queries/getAuditLogStatus'

// The toggle only changes on a redeploy, so it is fetched once per session.
export const useAuditLogStatus = () => {
  return useQuery({
    queryKey: coreKeys.auditLog.status(),
    queryFn: getAuditLogStatus,
    staleTime: Infinity,
  })
}

export const useAuditLogEnabled = (): boolean => useAuditLogStatus().data?.enabled ?? false
