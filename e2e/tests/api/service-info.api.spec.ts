import { test, expect, request, APIRequestContext } from '@playwright/test'
import {
  ASSESSMENT_API,
  BASE_URL,
  CERTIFICATE_API,
  EXAMPLE_API,
  INFRASTRUCTURE_SETUP_API,
  INTERVIEW_API,
  PRESENTATION_API,
  SELF_TEAM_ALLOCATION_API,
  TEAM_ALLOCATION_API,
} from '../../src/env'

// Every phase server reports its version on the public /info endpoint from the
// SERVER_IMAGE_TAG it was started with; the admin System Status page renders it
// per service. The deployments pass the tag of the image they run, the e2e
// stack builds from source and passes this marker (docker-compose.e2e.yml).
const EXPECTED_VERSION = 'local'

const SERVICES: ReadonlyArray<{ name: string; api: string }> = [
  { name: 'self-team-allocation', api: SELF_TEAM_ALLOCATION_API },
  { name: 'assessment', api: ASSESSMENT_API },
  { name: 'example-service', api: EXAMPLE_API },
  { name: 'interview', api: INTERVIEW_API },
  { name: 'certificate', api: CERTIFICATE_API },
  { name: 'presentation', api: PRESENTATION_API },
  { name: 'team-allocation', api: TEAM_ALLOCATION_API },
  { name: 'infrastructure-setup', api: INFRASTRUCTURE_SETUP_API },
]

const CLIENTS: ReadonlyArray<{ name: string; path: string }> = [
  { name: 'self_team_allocation_component', path: '/self-team-allocation' },
  { name: 'assessment_component', path: '/assessment' },
  { name: 'example_component', path: '/example' },
  { name: 'matching_component', path: '/matching' },
  { name: 'interview_component', path: '/interview' },
  { name: 'certificate_component', path: '/certificate' },
  { name: 'presentation_component', path: '/presentation' },
  { name: 'team_allocation_component', path: '/team-allocation' },
  { name: 'infrastructure_setup_component', path: '/infrastructure-setup' },
]

test.describe('service info', () => {
  let client: APIRequestContext

  test.beforeAll(async () => {
    client = await request.newContext({ baseURL: BASE_URL })
  })

  test.afterAll(async () => {
    await client.dispose()
  })

  for (const { name, api } of SERVICES) {
    test(`${name} reports its version`, async () => {
      const res = await client.get(`${api}/info`)
      expect(res.status()).toBe(200)

      const info = (await res.json()) as {
        serviceName: string
        version: string
        healthy: boolean
      }
      expect(info.serviceName).toBe(name)
      expect(info.version).toBe(EXPECTED_VERSION)
      expect(info.healthy).toBe(true)
    })
  }

  for (const { name, path } of CLIENTS) {
    test(`${name} serves its remote entry and reports its version`, async () => {
      const entry = await client.get(`${path}/remoteEntry.js`)
      expect(entry.status()).toBe(200)

      const res = await client.get(`${path}/mf-manifest.json`)
      expect(res.status()).toBe(200)

      const manifest = (await res.json()) as {
        name: string
        metaData: { buildInfo: { buildVersion: string } }
      }
      expect(manifest.name).toBe(name)
      expect(manifest.metaData.buildInfo.buildVersion).toBe(EXPECTED_VERSION)
    })
  }
})
