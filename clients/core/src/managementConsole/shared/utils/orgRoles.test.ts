import { describe, expect, it } from 'vitest'
import { hasAnyOrgRole, orgRolesOf, parseOrgRole } from './orgRoles'

describe('parseOrgRole', () => {
  it.each([
    ['org-ase-Admin', 'ase', 'Admin'],
    ['org-ase-Member', 'ase', 'Member'],
    ['org-ase-lab-Admin', 'ase-lab', 'Admin'],
    ['org-team-admin-Member', 'team-admin', 'Member'],
    ['org-cgi-lab-Member', 'cgi-lab', 'Member'],
    [`org-${'a'.repeat(50)}-Admin`, 'a'.repeat(50), 'Admin'],
  ])('reads %s as the %s role %s', (permission, slug, role) => {
    expect(parseOrgRole(permission)).toEqual({ slug, role })
  })

  it.each([
    ['the custom-group role of a course in semester org', 'org-ios-cg-Admin'],
    ['a custom-group role named Member', 'org-ios-cg-Member'],
    ['a custom-group role with a hyphenated group name', 'org-ios-cg-team-Admin'],
    ['a course lecturer role in semester org', 'org-ios-Lecturer'],
    ['a course editor role in semester org', 'org-ios-Editor'],
    ['a course student role in semester org', 'org-ios-Student'],
    ['a course role of another semester', 'ws25-ase-Lecturer'],
    ['the global admin role', 'PROMPT_Admin'],
    ['the global lecturer role', 'PROMPT_Lecturer'],
    ['a role without a slug', 'org--Admin'],
    ['the prefix and suffix only', 'org-Admin'],
    ['a one-character slug', 'org-a-Admin'],
    ['a slug that is too long', `org-${'a'.repeat(51)}-Admin`],
    ['an uppercase slug', 'org-ASE-Admin'],
    ['a course name with spaces', 'org-iOS Praktikum-Admin'],
    ['a lowercase role suffix', 'org-ase-admin'],
    ['an unknown role suffix', 'org-ase-Owner'],
    ['a role without the prefix', 'ase-Admin'],
    ['an uppercase prefix', 'Org-ase-Admin'],
    ['an empty string', ''],
  ])('rejects %s', (_case, permission) => {
    expect(parseOrgRole(permission)).toBeUndefined()
  })
})

describe('hasAnyOrgRole', () => {
  it('finds an org role among other roles', () => {
    expect(hasAnyOrgRole(['PROMPT_Lecturer', 'ws25-ase-Lecturer', 'org-ase-Member'])).toBe(true)
  })

  it('ignores roles that only look like org roles', () => {
    expect(hasAnyOrgRole(['PROMPT_Admin', 'org-ios-cg-Admin', 'org-ios-Lecturer'])).toBe(false)
    expect(hasAnyOrgRole([])).toBe(false)
  })
})

describe('orgRolesOf', () => {
  it('keeps the org roles, sorted, whatever order the token lists them in', () => {
    expect(
      orgRolesOf(['org-lab-Member', 'PROMPT_Lecturer', 'org-ase-Admin', 'org-ios-cg-Admin']),
    ).toEqual(['org-ase-Admin', 'org-lab-Member'])
    expect(orgRolesOf(['org-ase-Admin', 'org-lab-Member'])).toEqual(
      orgRolesOf(['org-lab-Member', 'org-ase-Admin']),
    )
  })

  it('is empty without org roles', () => {
    expect(orgRolesOf(['PROMPT_Admin', 'ws25-ase-Lecturer'])).toEqual([])
  })
})
