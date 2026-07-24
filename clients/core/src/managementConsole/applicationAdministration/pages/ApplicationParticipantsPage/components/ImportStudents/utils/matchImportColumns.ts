export type ColumnTarget =
  | 'firstName'
  | 'lastName'
  | 'universityLogin'
  | 'email'
  | 'matriculationNumber'
  | 'gender'
  | 'nationality'
  | 'studyProgram'
  | 'studyDegree'
  | 'currentSemester'
  | 'question'
  | 'ignore'

export type AttributeTarget = Exclude<ColumnTarget, 'question' | 'ignore'>

export const REQUIRED_TARGETS: AttributeTarget[] = [
  'firstName',
  'lastName',
  'universityLogin',
  'email',
]

export const TARGET_LABELS: Record<ColumnTarget, string> = {
  firstName: 'First Name',
  lastName: 'Last Name',
  universityLogin: 'University ID',
  email: 'Email',
  matriculationNumber: 'Matriculation Number',
  gender: 'Gender',
  nationality: 'Nationality',
  studyProgram: 'Study Program',
  studyDegree: 'Study Degree',
  currentSemester: 'Current Semester',
  question: 'Import as Question',
  ignore: 'Ignore',
}

const normalize = (header: string): string =>
  header
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]/g, '')

const ALIASES: Record<AttributeTarget, string[]> = {
  firstName: ['firstname', 'first', 'vorname', 'givenname'],
  lastName: ['lastname', 'last', 'surname', 'nachname', 'familyname'],
  universityLogin: ['universityid', 'universitylogin', 'tumid', 'login', 'kennung'],
  email: ['email', 'emailaddress', 'mail', 'mailaddress', 'emailadresse'],
  matriculationNumber: [
    'matriculationnumber',
    'matriculation',
    'matrikelnummer',
    'matrikel',
    'matriculationno',
  ],
  gender: ['gender', 'geschlecht', 'sex'],
  nationality: ['nationality', 'nationalitaet', 'country', 'citizenship'],
  studyProgram: ['studyprogram', 'studiengang', 'program', 'programme', 'major'],
  studyDegree: ['studydegree', 'degree', 'abschluss'],
  currentSemester: ['currentsemester', 'semester', 'fachsemester'],
}

const autoDetectTarget = (header: string): ColumnTarget => {
  const normalized = normalize(header)
  const entries = Object.entries(ALIASES) as [AttributeTarget, string[]][]
  for (const [target, aliases] of entries) {
    if (aliases.includes(normalized)) {
      return target
    }
  }
  return 'question'
}

/**
 * Builds the initial header -> target mapping by auto-detecting attribute columns. An attribute
 * target can be assigned to at most one column; extra matches fall back to being imported as a
 * question so the lecturer can decide.
 */
export const buildInitialMapping = (headers: string[]): Record<string, ColumnTarget> => {
  const mapping: Record<string, ColumnTarget> = {}
  const usedAttributeTargets = new Set<ColumnTarget>()
  for (const header of headers) {
    let target = autoDetectTarget(header)
    if (target !== 'question' && usedAttributeTargets.has(target)) {
      target = 'question'
    }
    if (target !== 'question') {
      usedAttributeTargets.add(target)
    }
    mapping[header] = target
  }
  return mapping
}

export interface MappingValidationError {
  message: string
}

/**
 * Validates that every required attribute is mapped to exactly one column and that no attribute is
 * mapped more than once. Returns null when the mapping is valid.
 */
export const validateMapping = (
  mapping: Record<string, ColumnTarget>,
): MappingValidationError | null => {
  const counts = new Map<ColumnTarget, number>()
  for (const target of Object.values(mapping)) {
    counts.set(target, (counts.get(target) ?? 0) + 1)
  }

  for (const required of REQUIRED_TARGETS) {
    const count = counts.get(required) ?? 0
    if (count === 0) {
      return { message: `Column for "${TARGET_LABELS[required]}" is required.` }
    }
  }

  const attributeTargets = Object.keys(ALIASES) as AttributeTarget[]
  for (const target of attributeTargets) {
    if ((counts.get(target) ?? 0) > 1) {
      return { message: `"${TARGET_LABELS[target]}" is mapped to more than one column.` }
    }
  }

  return null
}
