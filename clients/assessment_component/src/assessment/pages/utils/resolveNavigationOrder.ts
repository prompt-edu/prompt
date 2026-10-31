export const resolveNavigationOrder = (
  snapshotIds: string[] | undefined,
  liveIds: string[],
  currentId: string | undefined,
): string[] => {
  const liveIdSet = new Set(liveIds)
  const validSnapshotIds = (snapshotIds ?? []).filter((id) => liveIdSet.has(id))
  return currentId && validSnapshotIds.includes(currentId) ? validSnapshotIds : liveIds
}
