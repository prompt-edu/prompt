import { expect } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { test } from '../../src/fixtures/auth'
import { OrgDetailPage, OrgsPage } from '../../src/pages/OrgsPage'
import { uniqueSuffix } from '../courses/helpers'

interface Org {
  id: string
  parentOrgID: string | null
  slug: string
}

const deleteOrgsWithSlugs = async (slugs: string[]) => {
  const admin = await apiContextFor('admin')
  try {
    const response = await admin.get('/api/orgs/')
    const orgs = ((await response.json()) as Org[]).filter((org) => slugs.includes(org.slug))
    const children = orgs.filter((org) => org.parentOrgID !== null)
    const roots = orgs.filter((org) => org.parentOrgID === null)
    for (const org of [...children, ...roots]) {
      await admin.delete(`/api/orgs/${org.id}`)
    }
  } finally {
    await admin.dispose()
  }
}

test.describe('org administration', () => {
  test.use({ role: 'admin' })

  let parent: { name: string; slug: string }
  let child: { name: string; slug: string }

  test.beforeEach(({}, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    parent = { name: `E2E Chair ${suffix}`, slug: `e2e-chair-${suffix}` }
    child = { name: `E2E Lab ${suffix}`, slug: `e2e-lab-${suffix}` }
  })

  test.afterEach(async () => {
    await deleteOrgsWithSlugs([child.slug, parent.slug])
  })

  test('an admin creates a nested org, edits it, and deletes it once it is empty', async ({
    page,
  }) => {
    const orgs = new OrgsPage(page)
    await orgs.goto()

    await orgs.create({ name: parent.name })
    await expect(orgs.row(parent.slug)).toBeVisible()

    await orgs.create({
      name: child.name,
      slug: child.slug,
      parent: `${parent.name} (${parent.slug})`,
    })
    await expect(orgs.row(child.slug)).toContainText(`${parent.name} (${parent.slug})`)

    await orgs.open(child.slug, child.name)
    const detail = new OrgDetailPage(page)
    await detail.saveDetails({
      Name: `${child.name} renamed`,
      'University (optional)': 'Technical University of Munich',
      'Website (optional)': 'https://aet.cit.tum.de',
    })
    await detail.expectLoaded(`${child.name} renamed`)

    await detail.saveDetails({ 'Website (optional)': '' })
    await page.reload()
    await detail.expectLoaded(`${child.name} renamed`)
    await expect(detail.field('University (optional)')).toHaveValue('Technical University of Munich')
    await expect(detail.field('Website (optional)')).toHaveValue('')

    await orgs.goto()
    await orgs.open(parent.slug, parent.name)
    await expect(detail.childOrgLink(`${child.name} renamed (${child.slug})`)).toBeVisible()

    await detail.delete()
    await expect(page.getByText('the org still has child orgs', { exact: true })).toBeVisible()
    await detail.expectLoaded(parent.name)

    await detail.childOrgLink(`${child.name} renamed (${child.slug})`).click()
    await detail.expectLoaded(`${child.name} renamed`)
    await detail.delete()
    await expect(page.getByRole('heading', { level: 1, name: 'Orgs' })).toBeVisible()
    await expect(orgs.row(child.slug)).toBeHidden()

    await orgs.open(parent.slug, parent.name)
    await detail.delete()
    await expect(page.getByRole('heading', { level: 1, name: 'Orgs' })).toBeVisible()
    await expect(orgs.row(parent.slug)).toBeHidden()
  })

  test('the parent select offers neither the org itself nor its descendants', async ({
    page,
  }) => {
    const orgs = new OrgsPage(page)
    await orgs.goto()
    await orgs.create({ name: parent.name, slug: parent.slug })
    await orgs.create({
      name: child.name,
      slug: child.slug,
      parent: `${parent.name} (${parent.slug})`,
    })

    await orgs.open(parent.slug, parent.name)
    await page.getByRole('combobox', { name: 'Parent Org' }).click()
    await expect(page.getByRole('option', { name: 'None (top level)' })).toBeVisible()
    await expect(page.getByRole('option', { name: `${parent.name} (${parent.slug})` })).toHaveCount(0)
    await expect(page.getByRole('option', { name: `${child.name} (${child.slug})` })).toHaveCount(0)
    await page.keyboard.press('Escape')

    await orgs.goto()
    await orgs.open(child.slug, child.name)
    await new OrgDetailPage(page).moveTo('None (top level)')
    await orgs.goto()
    await expect(orgs.row(child.slug)).not.toContainText(parent.slug)
  })

  test('the create dialog rejects a slug that would collide with custom group roles', async ({
    page,
  }) => {
    const orgs = new OrgsPage(page)
    await orgs.goto()
    await page.getByRole('button', { name: 'Add Org' }).click()
    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('Name', { exact: true }).fill('Chair for CG')

    await expect(dialog.getByLabel('Slug')).toHaveValue('chair-for-cg')
    await expect(dialog.getByText('must not contain a "cg" segment')).toBeVisible()
    await expect(dialog.getByRole('button', { name: 'Create' })).toBeDisabled()
  })
})
