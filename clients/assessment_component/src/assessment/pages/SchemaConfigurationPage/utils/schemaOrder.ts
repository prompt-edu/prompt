import type {
  CategoryWithCompetencies,
  UpdateSchemaOrderRequest,
} from '../../../interfaces/category'

export const CATEGORY_DROP_TYPE = 'category'
export const COMPETENCY_DROP_TYPE = 'competency'

export interface CompetencyPosition {
  categoryID: string
  index: number
}

const moveItem = <T>(items: T[], fromIndex: number, toIndex: number): T[] => {
  const reordered = [...items]
  const [moved] = reordered.splice(fromIndex, 1)
  reordered.splice(toIndex, 0, moved)
  return reordered
}

export const moveCategory = (
  categories: CategoryWithCompetencies[],
  fromIndex: number,
  toIndex: number,
): CategoryWithCompetencies[] => moveItem(categories, fromIndex, toIndex)

export const moveCompetency = (
  categories: CategoryWithCompetencies[],
  from: CompetencyPosition,
  to: CompetencyPosition,
): CategoryWithCompetencies[] => {
  const moved = categories.find((category) => category.id === from.categoryID)?.competencies[
    from.index
  ]
  if (!moved || !categories.some((category) => category.id === to.categoryID)) {
    return categories
  }

  if (from.categoryID === to.categoryID) {
    return categories.map((category) =>
      category.id === from.categoryID
        ? { ...category, competencies: moveItem(category.competencies, from.index, to.index) }
        : category,
    )
  }

  return categories.map((category) => {
    if (category.id === from.categoryID) {
      return {
        ...category,
        competencies: category.competencies.filter((_, index) => index !== from.index),
      }
    }
    if (category.id === to.categoryID) {
      const competencies = [...category.competencies]
      competencies.splice(to.index, 0, { ...moved, categoryID: to.categoryID })
      return { ...category, competencies }
    }
    return category
  })
}

/** Returns the name that would appear twice in the target category, if the move would cause that. */
export const findCompetencyNameConflict = (
  categories: CategoryWithCompetencies[],
  from: CompetencyPosition,
  toCategoryID: string,
): string | undefined => {
  if (from.categoryID === toCategoryID) {
    return undefined
  }

  const moved = categories.find((category) => category.id === from.categoryID)?.competencies[
    from.index
  ]
  const target = categories.find((category) => category.id === toCategoryID)
  if (!moved || !target?.competencies.some((competency) => competency.name === moved.name)) {
    return undefined
  }
  return moved.name
}

export const toSchemaOrderRequest = (
  categories: CategoryWithCompetencies[],
): UpdateSchemaOrderRequest => ({
  categories: categories.map((category) => ({
    id: category.id,
    competencyIDs: category.competencies.map((competency) => competency.id),
  })),
})
