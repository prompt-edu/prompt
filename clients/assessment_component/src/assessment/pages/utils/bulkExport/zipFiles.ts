import { triggerBlobDataDownload } from '../triggerBlobDataDownload'

export interface ExportFile {
  baseName: string
  extension: string
  content: Blob | string
}

export const personFileBaseName = (person: { firstName: string; lastName: string }) =>
  `${person.lastName}_${person.firstName}`.replace(/[^\p{L}\p{N}-]+/gu, '_')

export const uniqueFileNames = (files: Pick<ExportFile, 'baseName' | 'extension'>[]) => {
  const used = new Set<string>()
  return files.map(({ baseName, extension }) => {
    let name = `${baseName}.${extension}`
    for (let n = 2; used.has(name); n++) name = `${baseName}_${n}.${extension}`
    used.add(name)
    return name
  })
}

export const downloadZip = async (files: ExportFile[], zipName: string) => {
  const { default: JSZip } = await import('jszip')
  const zip = new JSZip()
  const names = uniqueFileNames(files)
  files.forEach((file, index) => {
    zip.file(names[index], file.content)
  })
  triggerBlobDataDownload(await zip.generateAsync({ type: 'blob' }), zipName, 'application/zip')
}

export const exportDate = () => new Date().toISOString().slice(0, 10)
