/**
 * Downloads UTF-8 text as a file. Use `triggerBlobDataDownload` when the payload
 * is not text.
 */
export const triggerTextDownload = (content: string, filename: string, type: string): void => {
  const blob = new Blob([content], { type })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}
