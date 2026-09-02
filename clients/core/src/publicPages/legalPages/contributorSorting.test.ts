import { describe, expect, it } from 'vitest'
import { byPinnedThenCommits } from './contributorSorting'
import type { ContributorWithInfo } from './interfaces/Contributor'

const contributor = (
  login: string,
  contributions: number,
  pinnedPosition?: number,
): ContributorWithInfo => ({
  login,
  avatar_url: `https://avatars.githubusercontent.com/${login}`,
  html_url: `https://github.com/${login}`,
  contributions,
  type: 'User',
  name: login,
  contribution: 'Contributor',
  pinnedPosition,
})

const order = (contributors: ContributorWithInfo[]) =>
  [...contributors].sort(byPinnedThenCommits).map((c) => c.login)

describe('byPinnedThenCommits', () => {
  it('sorts unpinned contributors by descending commit count', () => {
    expect(
      order([contributor('few', 3), contributor('many', 320), contributor('some', 71)]),
    ).toEqual(['many', 'some', 'few'])
  })

  it('puts pinned contributors first regardless of commit count', () => {
    expect(order([contributor('prolific', 747), contributor('pinned', 33, 1)])).toEqual([
      'pinned',
      'prolific',
    ])
  })

  it('orders multiple pinned contributors by their pinned position', () => {
    expect(
      order([contributor('second', 500, 2), contributor('first', 1, 1), contributor('rest', 900)]),
    ).toEqual(['first', 'second', 'rest'])
  })

  it('keeps the incoming order for contributors with the same commit count', () => {
    expect(order([contributor('a', 1), contributor('b', 1), contributor('c', 1)])).toEqual([
      'a',
      'b',
      'c',
    ])
  })

  it('sorts an unpinned contributor without commits in this repository last', () => {
    expect(order([contributor('nocommits', 0), contributor('one', 1)])).toEqual([
      'one',
      'nocommits',
    ])
  })

  it('keeps a pinned contributor without commits in this repository at its pinned position', () => {
    expect(
      order([
        contributor('prolific', 747),
        contributor('nocommits', 0, 2),
        contributor('lead', 500),
      ]),
    ).toEqual(['nocommits', 'prolific', 'lead'])
  })
})
