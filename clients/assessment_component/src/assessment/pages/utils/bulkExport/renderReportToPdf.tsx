import type { ReactElement } from 'react'
import { flushSync } from 'react-dom'
import { type Block, pageStarts } from './pageStarts'

const MARGIN_MM = 15
const CONTENT_WIDTH_MM = 180
const CONTENT_HEIGHT_MM = 267
const CONTENT_WIDTH_PX = 680
const MM_PER_PX = CONTENT_WIDTH_MM / CONTENT_WIDTH_PX
const PAGE_HEIGHT_PX = CONTENT_HEIGHT_MM / MM_PER_PX
const SCALE = 2
const KEEP_TOGETHER = '.break-inside-avoid, h1, h2, h3, li'

const textLineRects = (container: HTMLElement): DOMRect[] => {
  const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT)
  const range = document.createRange()
  const rects: DOMRect[] = []
  while (walker.nextNode()) {
    range.selectNodeContents(walker.currentNode)
    rects.push(...Array.from(range.getClientRects()))
  }
  return rects
}

const blocksOf = (container: HTMLElement): Block[] => {
  const containerTop = container.getBoundingClientRect().top
  const elementRects = Array.from(container.querySelectorAll(KEEP_TOGETHER), (element) =>
    element.getBoundingClientRect(),
  )
  return [...elementRects, ...textLineRects(container)].map(({ top, bottom }) => ({
    top: top - containerTop,
    bottom: bottom - containerTop,
  }))
}

const cropPage = (canvas: HTMLCanvasElement, startPx: number, heightPx: number) => {
  const page = document.createElement('canvas')
  page.width = canvas.width
  page.height = Math.ceil(heightPx * SCALE)
  page
    .getContext('2d')
    ?.drawImage(
      canvas,
      0,
      startPx * SCALE,
      canvas.width,
      page.height,
      0,
      0,
      page.width,
      page.height,
    )
  return page
}

export const renderReportToPdf = async (report: ReactElement): Promise<Blob> => {
  const [{ createRoot }, { default: html2canvas }, { jsPDF }] = await Promise.all([
    import('react-dom/client'),
    import('html2canvas-pro'),
    import('jspdf'),
  ])

  const container = document.createElement('div')
  container.style.cssText = `position:fixed;top:0;left:-10000px;width:${CONTENT_WIDTH_PX}px;background:#fff`
  document.body.appendChild(container)
  const root = createRoot(container)
  try {
    flushSync(() => root.render(report))
    const contentHeight = container.scrollHeight
    const starts = pageStarts(contentHeight, PAGE_HEIGHT_PX, blocksOf(container))
    const canvas = await html2canvas(container, { scale: SCALE, backgroundColor: '#ffffff' })

    const pdf = new jsPDF({ unit: 'mm', format: 'a4' })
    starts.forEach((start, index) => {
      const height = (starts[index + 1] ?? contentHeight) - start
      if (index > 0) pdf.addPage()
      pdf.addImage(
        cropPage(canvas, start, height).toDataURL('image/jpeg', 0.92),
        'JPEG',
        MARGIN_MM,
        MARGIN_MM,
        CONTENT_WIDTH_MM,
        height * MM_PER_PX,
      )
    })
    return pdf.output('blob')
  } finally {
    root.unmount()
    container.remove()
  }
}
