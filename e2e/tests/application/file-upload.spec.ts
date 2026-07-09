import { test, expect } from '@playwright/test'
import { ApplicationFormPage } from '../../src/pages/ApplicationFormPage'
import { OPEN_APPLICATION_PHASE_ID } from '../../src/data/constants'

// Browser counterpart to tests/api/application-files.api.spec.ts: drives the
// real applicant FileUpload widget so the client's own presign -> PUT-to-S3 ->
// complete calls (from @tumaet/prompt-ui-components) get exercised, not
// hand-crafted HTTP. The public /apply page needs no login.
test.describe('application file upload (browser)', () => {
  test('an applicant uploads a file and it round-trips through S3', async ({ page }) => {
    const body = Buffer.from('e2e browser upload')
    const appForm = new ApplicationFormPage(page)

    await appForm.goto(OPEN_APPLICATION_PHASE_ID)
    await appForm.continueToForm()

    // Arm before selecting the file: the widget uploads on selection.
    const completePromise = page.waitForResponse(
      (r) => r.url().includes('/files/complete') && r.request().method() === 'POST',
    )
    await appForm.uploadFile('cv.txt', body)

    const complete = await completePromise
    expect(complete.status(), await complete.text()).toBe(201)
    await appForm.expectFileListed('cv.txt')

    // Presigned download TTL is ~30s, so fetch immediately.
    const { downloadUrl } = (await complete.json()) as { downloadUrl: string }
    expect(downloadUrl).toBeTruthy()
    const download = await page.request.get(downloadUrl)
    expect(download.ok(), `download failed: ${download.status()}`).toBeTruthy()
    expect(await download.body()).toEqual(body)
  })
})
