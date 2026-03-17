type OpenFileDownloadParams = {
  downloadUrl: string
  fileName?: string
}

export const openFileDownload = async ({
  downloadUrl,
  fileName,
}: OpenFileDownloadParams): Promise<void> => {
  if (!downloadUrl) {
    console.error('No download URL available for this file.')
    return
  }

  const link = document.createElement('a')
  link.href = downloadUrl
  link.target = '_blank'
  link.rel = 'noopener noreferrer'
  if (fileName) {
    link.download = fileName
  }
  document.body.appendChild(link)
  link.click()
  link.remove()
}
