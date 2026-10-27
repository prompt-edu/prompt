import { describe, expect, it } from 'vitest'
import type { CategoryWithCompetencies } from '../../../interfaces/category'
import type { Competency } from '../../../interfaces/competency'
import {
  findCompetencyNameConflict,
  moveCategory,
  moveCompetency,
  toSchemaOrderRequest,
} from './schemaOrder'

const competency = (id: string, categoryID: string, name = id): Competency => ({
  id,
  categoryID,
  name,
  shortName: '',
  description: '',
  descriptionVeryBad: '',
  descriptionBad: '',
  descriptionOk: '',
  descriptionGood: '',
  descriptionVeryGood: '',
  weight: 1,
})

const category = (id: string, competencyNames: string[]): CategoryWithCompetencies => ({
  id,
  name: id,
  shortName: id,
  weight: 1,
  competencies: competencyNames.map((name) => competency(`${id}-${name}`, id, name)),
})

const layout = (categories: CategoryWithCompetencies[]) =>
  categories.map((item) => [item.id, item.competencies.map((entry) => entry.name)])

const schema = () => [
  category('a', ['one', 'two', 'three']),
  category('b', ['four']),
  category('c', []),
]

describe('moveCategory', () => {
  it('moves a category to its new position', () => {
    expect(moveCategory(schema(), 0, 2).map((item) => item.id)).toEqual(['b', 'c', 'a'])
    expect(moveCategory(schema(), 2, 0).map((item) => item.id)).toEqual(['c', 'a', 'b'])
  })

  it('leaves the input untouched', () => {
    const categories = schema()

    moveCategory(categories, 0, 1)

    expect(categories.map((item) => item.id)).toEqual(['a', 'b', 'c'])
  })
})

describe('moveCompetency', () => {
  it('reorders competencies within a category', () => {
    const moved = moveCompetency(
      schema(),
      { categoryID: 'a', index: 0 },
      { categoryID: 'a', index: 2 },
    )

    expect(layout(moved)).toEqual([
      ['a', ['two', 'three', 'one']],
      ['b', ['four']],
      ['c', []],
    ])
  })

  it('moves a competency into another category and updates its category', () => {
    const moved = moveCompetency(
      schema(),
      { categoryID: 'a', index: 1 },
      { categoryID: 'b', index: 0 },
    )

    expect(layout(moved)).toEqual([
      ['a', ['one', 'three']],
      ['b', ['two', 'four']],
      ['c', []],
    ])
    expect(moved[1].competencies[0].categoryID).toBe('b')
  })

  it('moves a competency into an empty category', () => {
    const moved = moveCompetency(
      schema(),
      { categoryID: 'b', index: 0 },
      { categoryID: 'c', index: 0 },
    )

    expect(layout(moved)).toEqual([
      ['a', ['one', 'two', 'three']],
      ['b', []],
      ['c', ['four']],
    ])
  })

  it('ignores a move from or to an unknown position', () => {
    const categories = schema()

    expect(
      moveCompetency(categories, { categoryID: 'a', index: 9 }, { categoryID: 'b', index: 0 }),
    ).toBe(categories)
    expect(
      moveCompetency(categories, { categoryID: 'a', index: 0 }, { categoryID: 'x', index: 0 }),
    ).toBe(categories)
  })
})

describe('findCompetencyNameConflict', () => {
  const categories = [category('a', ['shared', 'unique']), category('b', ['shared'])]

  it('reports a name the target category already has', () => {
    expect(findCompetencyNameConflict(categories, { categoryID: 'a', index: 0 }, 'b')).toBe(
      'shared',
    )
  })

  it('allows names the target category does not have', () => {
    expect(
      findCompetencyNameConflict(categories, { categoryID: 'a', index: 1 }, 'b'),
    ).toBeUndefined()
  })

  it('allows reordering within the same category', () => {
    expect(
      findCompetencyNameConflict(categories, { categoryID: 'a', index: 0 }, 'a'),
    ).toBeUndefined()
  })
})

describe('toSchemaOrderRequest', () => {
  it('lists every category with its competency ids in order', () => {
    expect(toSchemaOrderRequest(schema())).toEqual({
      categories: [
        { id: 'a', competencyIDs: ['a-one', 'a-two', 'a-three'] },
        { id: 'b', competencyIDs: ['b-four'] },
        { id: 'c', competencyIDs: [] },
      ],
    })
  })
})
