import type { ReactElement } from 'react'
import { flushSync } from 'react-dom'
import { triggerBlobDataDownload } from '../triggerBlobDataDownload'
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

const forcedBreaksOf = (container: HTMLElement): number[] => {
  const containerTop = container.getBoundingClientRect().top
  return Array.from(
    container.querySelectorAll('.break-before-page'),
    (element) => element.getBoundingClientRect().top - containerTop,
  )
}

// Captures page by page through a clipped window so no canvas exceeds browser size limits.
const capturePdf = async (fill: (content: HTMLElement) => () => void): Promise<Blob> => {
  const [{ default: html2canvas }, { jsPDF }] = await Promise.all([
    import('html2canvas-pro'),
    import('jspdf'),
  ])

  const container = document.createElement('div')
  container.style.cssText = `position:fixed;top:0;left:-10000px;width:${CONTENT_WIDTH_PX}px;overflow:hidden;background:#fff`
  const content = document.createElement('div')
  container.appendChild(content)
  document.body.appendChild(container)
  let clear = () => {}
  try {
    clear = fill(content)
    const contentHeight = content.scrollHeight
    const starts = pageStarts(
      contentHeight,
      PAGE_HEIGHT_PX,
      blocksOf(content),
      forcedBreaksOf(content),
    )

    const pdf = new jsPDF({ unit: 'mm', format: 'a4' })
    for (const [index, start] of starts.entries()) {
      const height = (starts[index + 1] ?? contentHeight) - start
      container.style.height = `${height}px`
      content.style.transform = `translateY(-${start}px)`
      const page = await html2canvas(container, { scale: SCALE, backgroundColor: '#ffffff' })
      if (index > 0) pdf.addPage()
      pdf.addImage(
        page.toDataURL('image/jpeg', 0.92),
        'JPEG',
        MARGIN_MM,
        MARGIN_MM,
        CONTENT_WIDTH_MM,
        height * MM_PER_PX,
      )
    }
    return pdf.output('blob')
  } finally {
    clear()
    container.remove()
  }
}

export const renderReportToPdf = async (report: ReactElement): Promise<Blob> => {
  const { createRoot } = await import('react-dom/client')
  return capturePdf((content) => {
    const root = createRoot(content)
    flushSync(() => root.render(report))
    return () => root.unmount()
  })
}

// Exports the print reports mounted on the current page, the same content Ctrl+P prints.
export const downloadPrintReportsPdf = async (fileName: string) => {
  const pdf = await capturePdf((content) => {
    for (const report of document.querySelectorAll('.print-report')) {
      const copy = report.cloneNode(true) as HTMLElement
      copy.classList.remove('hidden')
      content.appendChild(copy)
    }
    if (!content.hasChildNodes()) throw new Error('This page has no report to export')
    return () => {}
  })
  triggerBlobDataDownload(pdf, fileName, 'application/pdf')
}
