import { test, expect, request, APIRequestContext } from '@playwright/test'
import { CORE_API_URL } from '../../src/env'
import { OPEN_APPLICATION_PHASE_ID } from '../../src/data/constants'

// The external /apply endpoints are public (no token). This drives the full
// presign → PUT-to-S3 → complete → download round-trip, so it fails if the
// SeaweedFS wiring breaks — something the mock-backed Go tests can't catch.
test.describe('core API: application file upload', () => {
  let ctx: APIRequestContext

  test.beforeAll(async () => {
    ctx = await request.newContext({ baseURL: CORE_API_URL })
  })
  test.afterAll(async () => {
    await ctx.dispose()
  })

  test('a file round-trips through S3 via the apply endpoints', async () => {
    const body = Buffer.from('e2e application file contents')
    const base = `/api/apply/${OPEN_APPLICATION_PHASE_ID}/files`

    const presign = await ctx.post(`${base}/presign`, {
      data: { filename: 'cv.pdf', contentType: 'application/pdf' },
    })
    expect(presign.ok(), await presign.text()).toBeTruthy()
    const { uploadUrl, storageKey } = (await presign.json()) as {
      uploadUrl: string
      storageKey: string
    }
    expect(uploadUrl).toBeTruthy()
    expect(storageKey).toContain(`course-phase/${OPEN_APPLICATION_PHASE_ID}/`)

    // Absolute presigned URL; overrides baseURL. This is the real write to S3.
    const put = await ctx.put(uploadUrl, {
      headers: { 'content-type': 'application/pdf' },
      data: body,
    })
    expect(put.ok(), `upload failed: ${put.status()}`).toBeTruthy()

    const complete = await ctx.post(`${base}/complete`, {
      data: {
        storageKey,
        originalFilename: 'cv.pdf',
        contentType: 'application/pdf',
      },
    })
    expect(complete.status(), await complete.text()).toBe(201)
    const { downloadUrl } = (await complete.json()) as { downloadUrl: string }
    expect(downloadUrl).toBeTruthy()

    const download = await ctx.get(downloadUrl)
    expect(download.ok()).toBeTruthy()
    expect(await download.body()).toEqual(body)
  })
})
