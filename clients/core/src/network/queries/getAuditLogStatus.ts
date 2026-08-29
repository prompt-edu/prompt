import type { AuditLogStatus } from '@core/managementConsole/auditLog/interfaces/auditLog'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getAuditLogStatus = async (): Promise<AuditLogStatus> => {
  return (await axiosInstance.get('/api/audit-log/status')).data
}
