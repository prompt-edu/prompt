interface CollapsedBreadcrumbs<T> {
  first: T
  hidden: T[]
  last: T
}

/**
 * Splits the trail for a header that is too narrow: the first and the last crumb stay visible,
 * everything between them goes into `hidden`. Trails of up to two crumbs have nothing to hide.
 */
export const collapseBreadcrumbs = <T>(crumbs: T[]): CollapsedBreadcrumbs<T> => ({
  first: crumbs[0],
  hidden: crumbs.slice(1, -1),
  last: crumbs[crumbs.length - 1],
})
