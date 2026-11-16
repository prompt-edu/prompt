import type { Org } from '@core/interfaces/org'
import { describe, expect, it } from 'vitest'
import {
  childOrgs,
  descendantIds,
  orgLabel,
  parentCandidates,
  withParentLabels,
} from './orgHierarchy'

const org = (id: string, parentOrgID: string | null = null): Org => ({
  id,
  parentOrgID,
  name: `Org ${id}`,
  slug: id,
  school: null,
  university: null,
  website: null,
  contactEmail: null,
  createdAt: '2026-10-27T12:00:00Z',
})

const orgs = [
  org('root'),
  org('child', 'root'),
  org('grandchild', 'child'),
  org('sibling', 'root'),
  org('other'),
]

const ids = (list: Org[]) => list.map(({ id }) => id)

describe('orgLabel', () => {
  it('names the org with its slug, since names are not unique', () => {
    expect(orgLabel({ name: 'AET', slug: 'aet' })).toBe('AET (aet)')
  })
})

describe('childOrgs', () => {
  it('returns only the direct children', () => {
    expect(ids(childOrgs(orgs, 'root'))).toEqual(['child', 'sibling'])
    expect(childOrgs(orgs, 'grandchild')).toEqual([])
  })
})

describe('descendantIds', () => {
  it('collects every level below the org', () => {
    expect([...descendantIds(orgs, 'root')].sort()).toEqual(['child', 'grandchild', 'sibling'])
    expect([...descendantIds(orgs, 'child')]).toEqual(['grandchild'])
    expect(descendantIds(orgs, 'other').size).toBe(0)
  })

  it('terminates on cyclic data', () => {
    const cyclic = [org('a', 'c'), org('b', 'a'), org('c', 'b')]
    expect([...descendantIds(cyclic, 'a')].sort()).toEqual(['a', 'b', 'c'])
  })

  it('handles a deep chain', () => {
    const chain = Array.from({ length: 10_000 }, (_, i) => org(`${i}`, i === 0 ? null : `${i - 1}`))
    expect(descendantIds(chain, '0').size).toBe(9_999)
  })
})

describe('parentCandidates', () => {
  it('offers every org when creating one', () => {
    expect(ids(parentCandidates(orgs))).toEqual(ids(orgs))
  })

  it('never offers the org itself or one of its descendants', () => {
    expect(ids(parentCandidates(orgs, 'child'))).toEqual(['root', 'sibling', 'other'])
    expect(ids(parentCandidates(orgs, 'root'))).toEqual(['other'])
  })
})

describe('withParentLabels', () => {
  it('labels each org with its parent, and top-level orgs with none', () => {
    const rows = withParentLabels(orgs)
    expect(rows.map(({ id, parentLabel }) => [id, parentLabel])).toEqual([
      ['root', null],
      ['child', 'Org root (root)'],
      ['grandchild', 'Org child (child)'],
      ['sibling', 'Org root (root)'],
      ['other', null],
    ])
  })
})
