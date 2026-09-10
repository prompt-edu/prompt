export interface TriggerSummary {
  queued: number
  requeued: number
  upToDate: number
}

export const describeTriggerSummary = ({ queued, requeued, upToDate }: TriggerSummary): string => {
  const parts: string[] = []
  if (queued > 0) parts.push(`${queued} queued`)
  if (requeued > 0) parts.push(`${requeued} retried`)
  if (upToDate > 0) parts.push(`${upToDate} already provisioned`)
  return parts.join(', ')
}
