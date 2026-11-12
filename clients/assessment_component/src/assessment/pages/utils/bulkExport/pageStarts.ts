export interface Block {
  top: number
  bottom: number
}

// A break that would cut a block moves up to its top, unless the block is taller than a page.
export const pageStarts = (
  contentHeight: number,
  pageHeight: number,
  blocks: Block[],
  forcedBreaks: number[] = [],
): number[] => {
  const fitting = blocks.filter((block) => block.bottom - block.top <= pageHeight)
  const starts = [0]
  let start = 0
  while (contentHeight - start > pageHeight || forcedBreaks.some((top) => top > start)) {
    const end = start + pageHeight
    const forced = forcedBreaks.filter((top) => top > start && top <= end)
    const cutTops = fitting
      .filter((block) => block.top > start && block.top < end && block.bottom > end)
      .map((block) => block.top)
    start = forced.length > 0 ? Math.min(...forced) : Math.min(end, ...cutTops)
    starts.push(start)
  }
  return starts
}
