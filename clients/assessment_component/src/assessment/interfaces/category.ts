import type { Competency } from './competency'

export type Category = {
  id: string
  name: string
  shortName: string
  description?: string
  weight: number
}

export type CategoryWithCompetencies = Category & {
  competencies: Competency[]
}

export interface CreateCategoryRequest {
  name: string
  shortName: string
  description?: string
  weight: number
  assessmentSchemaID: string
}

export interface UpdateCategoryRequest {
  id: string
  name?: string
  shortName?: string
  description?: string
  weight?: number
  assessmentSchemaID: string
}
