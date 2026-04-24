export interface CompanyRecord {
  preferredFieldOfStudy: string[]
  companyName: string
  fieldOfBusiness: string
  companySize: string
  ndaRequired: string
  teamSizeMin?: number
  teamSizeMax?: number
}

export interface ImportedTeamSizeConstraint {
  companyName: string
  lowerBound: number
  upperBound: number
}

export interface CompanyConstraintSuggestion {
  id: string
  title: string
  description: string
  companyNames: string[]
  property: 'de' | 'en' | 'cf-team-size'
  propertyName: string
  operator: '>=' | '=='
  value: string
  lowerBound: number
  upperBound: number
  canApplyAutomatically: boolean
}

export interface CompanyImportAnalysis {
  companyCount: number
  fieldsOfBusiness: Record<string, number>
  companySizes: Record<string, number>
  ndaRequiredCount: number
  preferredFieldsOfStudy: Record<string, number>
  suggestions: CompanyConstraintSuggestion[]
}

export const TEAM_ALLOCATION_PROFILE_STANDARD = 'standard'
export const TEAM_ALLOCATION_PROFILE_1000_PLUS = 'project_week_1000_plus'

export type TeamAllocationProfile =
  | typeof TEAM_ALLOCATION_PROFILE_STANDARD
  | typeof TEAM_ALLOCATION_PROFILE_1000_PLUS

export interface CompanyAllocationConfig {
  companies: CompanyRecord[]
  fields: string[]
  companyFieldMapping: Record<string, string>
  constraints: CompanyConstraintSuggestion[]
  importedTeamSizeConstraints: ImportedTeamSizeConstraint[]
  updatedAt: string
}
