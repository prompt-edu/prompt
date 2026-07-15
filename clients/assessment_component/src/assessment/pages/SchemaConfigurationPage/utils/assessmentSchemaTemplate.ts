import type { CategoryWithCompetencies } from '../../../interfaces/category'

export const ASSESSMENT_SCHEMA_TEMPLATE_VERSION = 1

export interface ExportedCompetency {
  name: string
  shortName: string
  description: string
  descriptionVeryBad: string
  descriptionBad: string
  descriptionOk: string
  descriptionGood: string
  descriptionVeryGood: string
  weight: number
}

export interface ExportedCategory {
  name: string
  shortName: string
  description?: string
  weight: number
  competencies: ExportedCompetency[]
}

export interface AssessmentSchemaTemplate {
  schemaVersion: number
  name?: string
  description?: string
  categories: ExportedCategory[]
}

export const buildAssessmentSchemaTemplate = (
  categories: CategoryWithCompetencies[],
  meta?: { name?: string; description?: string },
): AssessmentSchemaTemplate => ({
  schemaVersion: ASSESSMENT_SCHEMA_TEMPLATE_VERSION,
  ...(meta?.name ? { name: meta.name } : {}),
  ...(meta?.description ? { description: meta.description } : {}),
  categories: categories.map((category) => ({
    name: category.name,
    shortName: category.shortName,
    ...(category.description ? { description: category.description } : {}),
    weight: category.weight,
    competencies: category.competencies.map((competency) => ({
      name: competency.name,
      shortName: competency.shortName,
      description: competency.description,
      descriptionVeryBad: competency.descriptionVeryBad,
      descriptionBad: competency.descriptionBad,
      descriptionOk: competency.descriptionOk,
      descriptionGood: competency.descriptionGood,
      descriptionVeryGood: competency.descriptionVeryGood,
      weight: competency.weight,
    })),
  })),
})

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null && !Array.isArray(value)

const requireString = (value: unknown, field: string): string => {
  if (typeof value !== 'string') {
    throw new Error(`Field "${field}" must be a string.`)
  }
  return value
}

const optionalString = (value: unknown, field: string): string => {
  if (value === undefined || value === null) {
    return ''
  }
  return requireString(value, field)
}

const requireNumber = (value: unknown, field: string): number => {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    throw new Error(`Field "${field}" must be a number.`)
  }
  return value
}

const parseCompetency = (value: unknown, path: string): ExportedCompetency => {
  if (!isRecord(value)) {
    throw new Error(`${path} must be an object.`)
  }
  return {
    name: requireString(value.name, `${path}.name`),
    shortName: requireString(value.shortName, `${path}.shortName`),
    description: optionalString(value.description, `${path}.description`),
    descriptionVeryBad: optionalString(value.descriptionVeryBad, `${path}.descriptionVeryBad`),
    descriptionBad: optionalString(value.descriptionBad, `${path}.descriptionBad`),
    descriptionOk: optionalString(value.descriptionOk, `${path}.descriptionOk`),
    descriptionGood: optionalString(value.descriptionGood, `${path}.descriptionGood`),
    descriptionVeryGood: optionalString(value.descriptionVeryGood, `${path}.descriptionVeryGood`),
    weight: requireNumber(value.weight, `${path}.weight`),
  }
}

const parseCategory = (value: unknown, path: string): ExportedCategory => {
  if (!isRecord(value)) {
    throw new Error(`${path} must be an object.`)
  }
  if (!Array.isArray(value.competencies)) {
    throw new Error(`${path}.competencies must be an array.`)
  }
  const description = optionalString(value.description, `${path}.description`)
  return {
    name: requireString(value.name, `${path}.name`),
    shortName: requireString(value.shortName, `${path}.shortName`),
    ...(description ? { description } : {}),
    weight: requireNumber(value.weight, `${path}.weight`),
    competencies: value.competencies.map((competency, index) =>
      parseCompetency(competency, `${path}.competencies[${index}]`),
    ),
  }
}

export const parseAssessmentSchemaTemplate = (raw: string): AssessmentSchemaTemplate => {
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    throw new Error('The selected file is not valid JSON.')
  }

  if (!isRecord(parsed)) {
    throw new Error('The file does not contain an assessment schema template.')
  }

  if (parsed.schemaVersion !== ASSESSMENT_SCHEMA_TEMPLATE_VERSION) {
    throw new Error(
      `Unsupported schema version. Expected ${ASSESSMENT_SCHEMA_TEMPLATE_VERSION}, but received ${String(
        parsed.schemaVersion,
      )}.`,
    )
  }

  if (!Array.isArray(parsed.categories)) {
    throw new Error('The template must contain a "categories" array.')
  }

  return {
    schemaVersion: ASSESSMENT_SCHEMA_TEMPLATE_VERSION,
    ...(typeof parsed.name === 'string' ? { name: parsed.name } : {}),
    ...(typeof parsed.description === 'string' ? { description: parsed.description } : {}),
    categories: parsed.categories.map((category, index) =>
      parseCategory(category, `categories[${index}]`),
    ),
  }
}
