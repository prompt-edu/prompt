import type {
  CompanyAllocationConfig,
  CompanyConstraintSuggestion,
  CompanyImportAnalysis,
  CompanyRecord,
  ImportedTeamSizeConstraint,
} from '../interfaces/companyImport'

type CsvRow = Record<string, string>

const splitList = (value: string): string[] =>
  value
    .split(/[;,|]/)
    .map((item) => item.trim())
    .filter(Boolean)

const normalizeHeader = (header: string): string =>
  header
    .trim()
    .toLowerCase()
    .replace(/[\s_-]+/g, '')
    .replace(/[^a-z0-9]/g, '')

const valueFromRow = (row: CsvRow, aliases: string[], fallback = ''): string => {
  for (const alias of aliases) {
    const value = row[normalizeHeader(alias)]
    if (value !== undefined && value.trim() !== '') {
      return value.trim()
    }
  }
  return fallback
}

const parseCsvLine = (line: string, delimiter: string): string[] => {
  const cells: string[] = []
  let cell = ''
  let insideQuotes = false

  for (let index = 0; index < line.length; index += 1) {
    const char = line[index]
    const nextChar = line[index + 1]

    if (char === '"' && nextChar === '"') {
      cell += '"'
      index += 1
      continue
    }

    if (char === '"') {
      insideQuotes = !insideQuotes
      continue
    }

    if (char === delimiter && !insideQuotes) {
      cells.push(cell.trim())
      cell = ''
      continue
    }

    cell += char
  }

  cells.push(cell.trim())
  return cells
}

const detectDelimiter = (headerLine: string): string => {
  const delimiters = [',', ';', '\t']
  return delimiters.reduce((best, delimiter) => {
    const currentCount = parseCsvLine(headerLine, delimiter).length
    const bestCount = parseCsvLine(headerLine, best).length
    return currentCount > bestCount ? delimiter : best
  }, ',')
}

const countBy = (values: string[]): Record<string, number> =>
  values.reduce<Record<string, number>>((counts, value) => {
    if (!value) return counts
    counts[value] = (counts[value] ?? 0) + 1
    return counts
  }, {})

const yesLike = (value: string): boolean =>
  ['yes', 'true', '1', 'ja', 'y'].includes(value.toLowerCase())

const parseOptionalInteger = (value: string): number | undefined => {
  if (!value.trim()) return undefined
  const parsed = Number.parseInt(value, 10)
  if (Number.isNaN(parsed) || parsed < 0) {
    throw new Error(`Invalid integer value "${value}" in CSV.`)
  }
  return parsed
}

export const parseCompanyCsv = (csvContent: string): CompanyRecord[] => {
  const lines = csvContent
    .replace(/^\uFEFF/, '')
    .split(/\r?\n/)
    .filter((line) => line.trim() !== '')

  if (lines.length < 2) {
    throw new Error('CSV must include a header row and at least one company row.')
  }

  const delimiter = detectDelimiter(lines[0])
  const headers = parseCsvLine(lines[0], delimiter).map(normalizeHeader)

  const rows = lines.slice(1).map((line) =>
    parseCsvLine(line, delimiter).reduce<CsvRow>((row, value, index) => {
      row[headers[index] ?? `column${index}`] = value
      return row
    }, {}),
  )

  return rows.map((row, index) => {
    const companyName = valueFromRow(row, ['companyName', 'company', 'name'])
    if (!companyName) {
      throw new Error(`Company name is missing in row ${index + 2}.`)
    }

    const teamSizeMin = parseOptionalInteger(valueFromRow(row, ['teamSizeMin', 'minTeamSize']))
    const teamSizeMax = parseOptionalInteger(valueFromRow(row, ['teamSizeMax', 'maxTeamSize']))
    if (teamSizeMin !== undefined && teamSizeMax !== undefined && teamSizeMin > teamSizeMax) {
      throw new Error(`Team size minimum exceeds maximum in row ${index + 2}.`)
    }

    return {
      preferredFieldOfStudy: splitList(
        valueFromRow(row, ['preferredFieldOfStudy', 'preferredStudyPrograms', 'studyPrograms']),
      ),
      companyName,
      fieldOfBusiness: valueFromRow(
        row,
        ['fieldOfBusiness', 'businessField', 'industry'],
        'Unspecified',
      ),
      companySize: valueFromRow(row, ['companySize', 'size'], 'Unspecified'),
      ndaRequired: yesLike(valueFromRow(row, ['ndaRequired', 'nda'])) ? 'yes' : 'no',
      teamSizeMin,
      teamSizeMax,
    }
  })
}

export const analyseCompanyRecords = (companies: CompanyRecord[]): CompanyImportAnalysis => {
  const fieldsOfBusiness = countBy(companies.map((company) => company.fieldOfBusiness))
  const companySizes = countBy(companies.map((company) => company.companySize))
  const preferredFieldsOfStudy = countBy(
    companies.flatMap((company) => company.preferredFieldOfStudy),
  )
  const ndaCompanies = companies.filter((company) => yesLike(company.ndaRequired))

  const suggestions: CompanyConstraintSuggestion[] = []

  companies.forEach((company) => {
    if (company.teamSizeMin === undefined && company.teamSizeMax === undefined) {
      return
    }

    suggestions.push({
      id: `team-size-${company.companyName.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`,
      title: `Team size for ${company.companyName}`,
      description: 'Apply the CSV-defined team size bounds to this imported company project.',
      companyNames: [company.companyName],
      property: 'cf-team-size',
      propertyName: 'Team Size',
      operator: '==',
      value: 'true',
      lowerBound: company.teamSizeMin ?? 0,
      upperBound: company.teamSizeMax ?? company.teamSizeMin ?? 999,
      canApplyAutomatically: true,
    })
  })

  if (ndaCompanies.length > 0) {
    suggestions.push({
      id: 'nda-german-b1',
      title: 'German communication coverage for NDA projects',
      description: 'Require at least one German B1/B2 speaker in companies marked as NDA-required.',
      companyNames: ndaCompanies.map((company) => company.companyName),
      property: 'de',
      propertyName: 'German',
      operator: '>=',
      value: 'B1/B2',
      lowerBound: 1,
      upperBound: 6,
      canApplyAutomatically: true,
    })
  }

  suggestions.push({
    id: 'english-b1-default',
    title: 'English communication baseline',
    description: 'Require at least two English B1/B2 speakers per imported company project.',
    companyNames: companies.map((company) => company.companyName),
    property: 'en',
    propertyName: 'English',
    operator: '>=',
    value: 'B1/B2',
    lowerBound: 2,
    upperBound: 6,
    canApplyAutomatically: true,
  })

  return {
    companyCount: companies.length,
    fieldsOfBusiness,
    companySizes,
    ndaRequiredCount: ndaCompanies.length,
    preferredFieldsOfStudy,
    suggestions,
  }
}

export const buildCompanyAllocationConfig = (
  companies: CompanyRecord[],
  analysis: CompanyImportAnalysis,
): CompanyAllocationConfig => {
  const fields = Object.keys(analysis.fieldsOfBusiness).sort((a, b) => a.localeCompare(b))
  const importedTeamSizeConstraints: ImportedTeamSizeConstraint[] = companies
    .filter((company) => company.teamSizeMin !== undefined || company.teamSizeMax !== undefined)
    .map((company) => ({
      companyName: company.companyName,
      lowerBound: company.teamSizeMin ?? 0,
      upperBound: company.teamSizeMax ?? company.teamSizeMin ?? 999,
    }))

  return {
    companies,
    fields,
    companyFieldMapping: companies.reduce<Record<string, string>>((mapping, company) => {
      mapping[company.companyName] = company.fieldOfBusiness
      return mapping
    }, {}),
    constraints: analysis.suggestions,
    importedTeamSizeConstraints,
    updatedAt: new Date().toISOString(),
  }
}
