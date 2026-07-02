import {
  test as base,
  request,
  APIRequestContext,
  expect,
} from '@playwright/test'
import { CORE_API_URL, KEYCLOAK_CLIENT_ID, tokenEndpoint } from '../env'
import { Role, ROLES } from '../data/roles'

// Mint a bearer token via Keycloak's direct access grant (password grant).
// `prompt-client` is a public client with this grant enabled, so no secret.
export async function tokenFor(role: Role): Promise<string> {
  const account = ROLES[role]
  const ctx = await request.newContext()
  try {
    const res = await ctx.post(tokenEndpoint(), {
      form: {
        grant_type: 'password',
        client_id: KEYCLOAK_CLIENT_ID,
        username: account.username,
        password: account.password,
        scope: 'openid',
      },
    })
    if (!res.ok()) {
      throw new Error(
        `token request for "${role}" failed: ${res.status()} ${await res.text()}`,
      )
    }
    const body = (await res.json()) as { access_token: string }
    return body.access_token
  } finally {
    await ctx.dispose()
  }
}

// An APIRequestContext targeting the core API, authenticated as the given role.
export async function apiContextFor(role: Role): Promise<APIRequestContext> {
  const token = await tokenFor(role)
  return request.newContext({
    baseURL: CORE_API_URL,
    extraHTTPHeaders: { Authorization: `Bearer ${token}` },
  })
}

// API-layer test with an `apiAs(role)` helper. Created contexts are disposed
// automatically at the end of the test.
export const test = base.extend<{
  apiAs: (role: Role) => Promise<APIRequestContext>
}>({
  apiAs: async ({}, use) => {
    const created: APIRequestContext[] = []
    const factory = async (role: Role) => {
      const ctx = await apiContextFor(role)
      created.push(ctx)
      return ctx
    }
    await use(factory)
    await Promise.all(created.map((c) => c.dispose()))
  },
})

export { expect }
