export interface Block {
  top: number
  bottom: number
}

// A break that would cut a block moves up to its top, unless the block is taller than a page.
export const pageStarts = (
  contentHeight: number,
  pageHeight: number,
  blocks: Block[],
): number[] => {
  const fitting = blocks.filter((block) => block.bottom - block.top <= pageHeight)
  const starts = [0]
  let start = 0
  while (contentHeight - start > pageHeight) {
    const end = start + pageHeight
    const cutTops = fitting
      .filter((block) => block.top > start && block.top < end && block.bottom > end)
      .map((block) => block.top)
    start = Math.min(end, ...cutTops)
    starts.push(start)
  }
  return starts
}
