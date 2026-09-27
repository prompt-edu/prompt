/**
 * Picks the order to navigate in: the snapshot taken on the overview, reduced to ids that still
 * exist, as long as it contains the current id. Otherwise the live order.
 */
export const resolveNavigationOrder = (
  snapshotIds: string[] | undefined,
  liveIds: string[],
  currentId: string | undefined,
): string[] => {
  const liveIdSet = new Set(liveIds)
  const validSnapshotIds = (snapshotIds ?? []).filter((id) => liveIdSet.has(id))
  return currentId && validSnapshotIds.includes(currentId) ? validSnapshotIds : liveIds
}
