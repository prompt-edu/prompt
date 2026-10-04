type Id = string | undefined

export const certificateKeys = {
  config: (phaseId: Id) => ['config', phaseId] as const,
  // Not the shared-state `participants` entry: this list carries download statuses
  participants: (phaseId: Id) => ['certificate-participants', phaseId] as const,
  myStatus: (phaseId: Id) => ['certificateStatus', phaseId] as const,
}
