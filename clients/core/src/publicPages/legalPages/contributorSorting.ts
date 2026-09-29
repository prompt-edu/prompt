import type { ContributorWithInfo } from './interfaces/Contributor'

/**
 * Orders the about page contributor cards. Pinned entries come first in their configured
 * order, everyone else follows by commit count. The GitHub API already returns contributors
 * sorted by commit count, so ties keep its order.
 */
export const byPinnedThenCommits = (a: ContributorWithInfo, b: ContributorWithInfo) => {
  if (a.pinnedPosition !== b.pinnedPosition) {
    return (
      (a.pinnedPosition ?? Number.MAX_SAFE_INTEGER) - (b.pinnedPosition ?? Number.MAX_SAFE_INTEGER)
    )
  }
  return b.contributions - a.contributions
}
