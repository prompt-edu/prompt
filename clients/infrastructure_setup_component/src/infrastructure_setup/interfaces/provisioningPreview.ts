// What a trigger would do right now, answered by the server without doing it.
export interface ProvisioningPreview {
  queued: number
  requeued: number
  upToDate: number
  // Instances a run is still working on. While above zero a trigger is refused.
  running: number
  // Null when no resource config uses that scope.
  teams: number | null
  students: number | null
}
