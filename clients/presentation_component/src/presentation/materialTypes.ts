import type { MaterialType } from './interfaces'

export interface MaterialTypeDefinition {
  type: MaterialType
  label: string
  description: string
  // Shown to presenters and used as the file dialog's accept filter. The server enforces
  // the matching media types, so this only has to be the friendly version of the same rule.
  extensions: string[]
  formats: string
  // Rendered as a muted note under the label, for limits worth mentioning up front.
  note?: string
}

// Keys and order mirror the server catalog in
// servers/presentation/presentation/materialTypes.go. A key may be added but never renamed,
// because stored requirements and uploads reference it.
export const MATERIAL_TYPE_CATALOG: MaterialTypeDefinition[] = [
  {
    type: 'slides',
    label: 'Presentation slides',
    description: 'The deck presented during the session.',
    extensions: ['.pdf', '.ppt', '.pptx', '.odp'],
    formats: 'PDF, PPT, PPTX or ODP',
  },
  {
    type: 'summary',
    label: 'Summary or report',
    description: 'A written summary of the presented work.',
    extensions: ['.pdf', '.doc', '.docx', '.odt'],
    formats: 'PDF, DOC, DOCX or ODT',
  },
  {
    type: 'handout',
    label: 'Handout',
    description: 'A one-pager for the audience.',
    extensions: ['.pdf'],
    formats: 'PDF',
  },
  {
    type: 'poster',
    label: 'Poster',
    description: 'A poster version of the presentation.',
    extensions: ['.pdf', '.png', '.jpg', '.jpeg'],
    formats: 'PDF, PNG or JPG',
  },
  {
    type: 'code',
    label: 'Source code archive',
    description: 'The code behind the presented work.',
    extensions: ['.zip'],
    formats: 'ZIP',
  },
  {
    type: 'recording',
    label: 'Video recording',
    description: 'A recording of the presentation.',
    extensions: ['.mp4'],
    formats: 'MP4',
    note: 'Keep it below the upload size limit.',
  },
]

const definitionsByType = new Map<string, MaterialTypeDefinition>(
  MATERIAL_TYPE_CATALOG.map((definition) => [definition.type, definition]),
)

// Unknown keys only appear when a phase was configured by a newer server, so they fall back
// to the raw key rather than hiding the upload.
export const getMaterialTypeDefinition = (type: string): MaterialTypeDefinition =>
  definitionsByType.get(type) ?? {
    type: type as MaterialType,
    label: type,
    description: '',
    extensions: [],
    formats: 'Any allowed file type',
  }

export const getMaterialTypeAccept = (definition: MaterialTypeDefinition): string | undefined =>
  definition.extensions.length > 0 ? definition.extensions.join(',') : undefined

// Sorts a stored requirement list into catalog order, so the slots always appear in the same
// sequence no matter how the lecturer selected them.
export const sortMaterialTypes = (types: MaterialType[]): MaterialType[] =>
  MATERIAL_TYPE_CATALOG.filter((definition) => types.includes(definition.type)).map(
    (definition) => definition.type,
  )
