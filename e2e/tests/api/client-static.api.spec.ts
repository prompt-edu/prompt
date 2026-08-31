import { test, expect, request, APIRequestContext } from '@playwright/test'
import { BASE_URL } from '../../src/env'

// Static-serving contract of the client nginx (clients/nginx/spa.conf, included
// by e2e/nginx/client-core.conf). A missing asset must 404: answering it with
// the SPA document under HTTP 200 makes the browser parse HTML as JavaScript,
// which white-screens the app with no recoverable error.
test.describe('client static serving', () => {
  let client: APIRequestContext

  test.beforeAll(async () => {
    client = await request.newContext({ baseURL: BASE_URL })
  })

  test.afterAll(async () => {
    await client.dispose()
  })

  test('a missing bundle 404s instead of returning the app shell', async () => {
    const res = await client.get('/main.0000000000000000.js')
    expect(res.status()).toBe(404)
  })

  test('a missing precompressed bundle 404s', async () => {
    const res = await client.get('/main.0000000000000000.js.gz')
    expect(res.status()).toBe(404)
  })

  test('a missing stylesheet 404s', async () => {
    const res = await client.get('/dist/styles.css')
    expect(res.status()).toBe(404)
  })

  test('the API is not answered with the app shell', async () => {
    const res = await client.get('/api/courses')
    expect(res.status()).toBe(404)
    expect(await res.text()).not.toContain('id="root"')
  })

  test('an unknown route still falls back to the app shell, uncached', async () => {
    const res = await client.get('/courses/does-not-exist/deep')
    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('text/html')
    expect(res.headers()['cache-control']).toContain('no-store')
    expect(await res.text()).toContain('id="root"')
  })

  test('the entry document is served uncached', async () => {
    for (const path of ['/', '/index.html']) {
      const res = await client.get(path)
      expect(res.status()).toBe(200)
      expect(res.headers()['cache-control']).toContain('no-store')
    }
  })

  test('the generated runtime config is served uncached', async () => {
    const res = await client.get('/env.js')
    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('javascript')
    expect(res.headers()['cache-control']).toContain('no-store')
    expect(await res.text()).toContain('window.env')
  })

  test('a remote manifest is proxied and served uncached', async () => {
    const res = await client.get('/assessment/remoteEntry.js')
    expect(res.status()).toBe(200)
    expect(res.headers()['content-type']).toContain('javascript')
    expect(res.headers()['cache-control']).toContain('no-store')
  })
})
