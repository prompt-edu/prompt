/**
 * Downloads raw bytes as a file. The byte-oriented counterpart to
 * `triggerTextDownload`, needed whenever the payload is not UTF-8 text (for
 * example a Windows-1252 encoded CSV).
 */
export const triggerBlobDataDownload = (data: BlobPart, filename: string, type: string): void => {
  const blob = new Blob([data], { type })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}
