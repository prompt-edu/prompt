import { Page, Locator, expect } from '@playwright/test'

// Seeded course_phase_type ids (e2e/seed/e2e_seed.sql). Application is the only
// initial-phase type, so it must be added first; Matching is a core-based
// non-initial type (no phase-module service needed).
export const PHASE_TYPE_IDS: Record<string, string> = {
  Application: 'a1111111-1111-1111-1111-111111111111',
  Matching: 'b2222222-2222-2222-2222-222222222222',
}

// /management/course/:id/configurator — the @xyflow/react DAG editor.
export class CourseConfiguratorPage {
  readonly heading: Locator

  constructor(
    private readonly page: Page,
    private readonly courseId: string,
  ) {
    this.heading = page.getByRole('heading', { name: 'Course Configurator' })
  }

  async goto() {
    await this.page.goto(`/management/course/${this.courseId}/configurator`)
    await expect(this.heading).toBeVisible({ timeout: 20_000 })
    const flow = this.page.locator('.react-flow')
    await flow.waitFor({ timeout: 20_000 })
    // Zoom out so dropped nodes render small: at the default zoom the wide phase
    // cards overlap and their participants handles sit under a node body. The
    // zoom Controls sit under the phase-type panel, so dispatch the click
    // directly on the button (bypasses the overlap hit-test).
    const zoomOut = this.page.locator('.react-flow__controls-zoomout')
    for (let i = 0; i < 5; i++) await zoomOut.dispatchEvent('click')
  }

  private async collapsePanelHover() {
    const box = await this.page.locator('.react-flow').boundingBox()
    if (box) await this.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  }

  // A node located by its visible phase-type name (stable across the temp->
  // persisted id transition that happens on save).
  node(phaseTypeName: string): Locator {
    return this.page.locator('.react-flow__node').filter({ hasText: phaseTypeName })
  }

  // Drags a phase type from the palette onto the canvas. ReactFlow uses native
  // HTML5 drag-drop, which Playwright's mouse-based drag cannot drive, so we
  // dispatch the events with a shared DataTransfer. onDrop is wired on the
  // .react-flow root (Canvas.tsx). Returns the new node's data-id (a temp
  // no-valid-id-* until the next save).
  async addPhase(phaseTypeName: string): Promise<string> {
    const typeId = PHASE_TYPE_IDS[phaseTypeName]
    await this.page.getByTestId('phase-type-panel').hover()
    const item = this.page.getByTestId(`phase-type-item-${typeId}`)
    await expect(item).toBeVisible()

    const flow = this.page.locator('.react-flow')
    const box = await flow.boundingBox()
    if (!box) throw new Error('no .react-flow bounding box')
    const existing = await this.page.locator('.react-flow__node').count()
    const clientX = box.x + 180 + existing * 360
    const clientY = box.y + box.height * 0.45

    const dt = await this.page.evaluateHandle(() => new DataTransfer())
    await item.dispatchEvent('dragstart', { dataTransfer: dt })
    await flow.dispatchEvent('dragover', {
      dataTransfer: dt,
      bubbles: true,
      cancelable: true,
      clientX,
      clientY,
    })
    await flow.dispatchEvent('drop', {
      dataTransfer: dt,
      bubbles: true,
      cancelable: true,
      clientX,
      clientY,
    })

    const node = this.node(phaseTypeName)
    await expect(node).toBeVisible()
    const id = await node.getAttribute('data-id')
    if (!id) throw new Error(`node "${phaseTypeName}" has no data-id`)
    return id
  }

  // Connects two phases via their participants handles (pre-save temp ids).
  // ReactFlow's connectOnClick is enabled, so clicking the source handle then
  // the target handle completes the connection through the same onConnect path
  // as a drag (Playwright's synthetic pointer drag does not drive xyflow
  // connections reliably).
  async connect(sourceNodeId: string, targetNodeId: string) {
    await this.collapsePanelHover()
    const source = this.page.locator(
      `.react-flow__handle[data-handleid="participants-out-${sourceNodeId}"]`,
    )
    const target = this.page.locator(
      `.react-flow__handle[data-handleid="participants-in-${targetNodeId}"]`,
    )
    await source.click()
    await target.click()
    await expect(this.page.locator('.react-flow__edge')).toHaveCount(1)
  }

  // Selects a node and deletes it (Backspace is ReactFlow's default delete key),
  // then confirms in the shared alert dialog.
  async removePhase(phaseTypeName: string) {
    const node = this.node(phaseTypeName)
    await node.click()
    await this.page.keyboard.press('Backspace')
    const dialog = this.page.getByRole('alertdialog')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: 'Delete' }).click()
    await expect(node).toHaveCount(0)
  }

  async save() {
    const save = this.page.getByTestId('configurator-save')
    await expect(save).toBeVisible()
    await expect(save).toBeEnabled()
    await save.click()
    await expect(this.page.getByTestId('configurator-save')).toBeHidden()
  }
}
