export type Competency = {
  id: string
  categoryID: string
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

export type CreateCompetencyRequest = {
  categoryID: string
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

export type UpdateCompetencyRequest = {
  id: string
  categoryID?: string
  name?: string
  shortName?: string
  description?: string
  descriptionVeryBad?: string
  descriptionBad?: string
  descriptionOk?: string
  descriptionGood?: string
  descriptionVeryGood?: string
  weight?: number
}
