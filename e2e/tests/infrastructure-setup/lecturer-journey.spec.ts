import { expect } from '@playwright/test'
import { test } from '../../src/fixtures/auth'
import { INFRASTRUCTURE_SETUP_PHASE_ID } from '../../src/data/constants'
import {
  UNRESPONSIVE_PROVIDER_URL,
  createResourceConfigRaw,
  createStudentResourceConfig,
  getPhaseConfig,
  listInstances,
  listProviderConfigs,
  resetInfrastructureSetupPhase,
  setSemesterTag,
  triggerExecution,
  upsertGitlabProvider,
} from './helpers'

const PHASE_ID = INFRASTRUCTURE_SETUP_PHASE_ID

// The `lecturer` user holds the course-scoped ios2425-iPraktikumFull-Lecturer role.
test.use({ role: 'lecturer' })

// The phase is driven through its own API rather than the UI: this journey is about
// server behaviour (readiness, validation, the conflict guard). Rendering the remote
// inside the core shell is covered by the module federation smoke test.
test.describe('infrastructure setup: lecturer journey', () => {
  test.beforeAll(async () => {
    await resetInfrastructureSetupPhase(PHASE_ID)
  })

  test.afterAll(async () => {
    await resetInfrastructureSetupPhase(PHASE_ID)
  })

  test('a lecturer configures the phase, and a second run is refused', async () => {
    // 1. An unconfigured phase reports nothing as ready.
    const initial = await getPhaseConfig(PHASE_ID)
    expect(initial.status).toBe(200)
    expect(JSON.parse(initial.body)).toMatchObject({
      semesterTag: false,
      providerConfig: false,
      resourceConfig: false,
    })

    // 2. A resource config cannot be created before its provider has credentials.
    const premature = await createResourceConfigRaw(PHASE_ID, {
      providerType: 'gitlab',
      resourceType: 'group',
      scope: 'per_student',
      nameTemplate: '{{semesterTag}}-{{studentLogin}}',
      permissionMapping: {},
      resourceExtraConfig: {},
    })
    expect(premature.status).toBe(400)

    // 3. Configure the phase.
    await setSemesterTag(PHASE_ID, 'e2e25')
    await upsertGitlabProvider(PHASE_ID, UNRESPONSIVE_PROVIDER_URL)

    // Credentials are never returned; the provider only reports that it holds some.
    const providers = await listProviderConfigs(PHASE_ID)
    expect(providers.status).toBe(200)
    expect(providers.body).not.toContain('e2e-token')
    expect(JSON.parse(providers.body)).toMatchObject([
      { providerType: 'gitlab', configured: true },
    ])

    // 4. An unknown placeholder is rejected when the template is saved.
    const badTemplate = await createResourceConfigRaw(PHASE_ID, {
      providerType: 'gitlab',
      resourceType: 'group',
      scope: 'per_student',
      nameTemplate: '{{unknownPlaceholder}}',
      permissionMapping: {},
      resourceExtraConfig: {},
    })
    expect(badTemplate.status).toBe(400)

    // A resource type the provider cannot create is rejected too.
    const badResourceType = await createResourceConfigRaw(PHASE_ID, {
      providerType: 'gitlab',
      resourceType: 'channel',
      scope: 'per_student',
      nameTemplate: '{{semesterTag}}-{{studentLogin}}',
      permissionMapping: {},
      resourceExtraConfig: {},
    })
    expect(badResourceType.status).toBe(400)

    // A templated extra-config value is validated like a name template, so an
    // unresolvable placeholder cannot reach a real GitLab namespace during a run.
    const badParentGroupTemplate = await createResourceConfigRaw(PHASE_ID, {
      providerType: 'gitlab',
      resourceType: 'project',
      scope: 'per_student',
      nameTemplate: '{{semesterTag}}-{{studentLogin}}-app',
      permissionMapping: { student: 'developer' },
      resourceExtraConfig: { parent_group_template: '{{unknownPlaceholder}}' },
    })
    expect(badParentGroupTemplate.status).toBe(400)

    // GitLab projects are a supported kind now, so a well-formed project config is
    // accepted alongside the group one.
    const projectConfig = await createResourceConfigRaw(PHASE_ID, {
      providerType: 'gitlab',
      resourceType: 'project',
      scope: 'per_student',
      nameTemplate: '{{semesterTag}}-{{studentLogin}}-app',
      permissionMapping: { student: 'developer' },
      resourceExtraConfig: { parent_group_template: '{{semesterTag}}-{{studentLogin}}' },
    })
    expect(projectConfig.status).toBe(201)

    await createStudentResourceConfig(PHASE_ID, '{{semesterTag}}-{{studentLogin}}')

    // 5. The phase now reports itself ready.
    const ready = await getPhaseConfig(PHASE_ID)
    expect(JSON.parse(ready.body)).toMatchObject({
      semesterTag: true,
      providerConfig: true,
      resourceConfig: true,
    })

    // 6. Trigger provisioning. The provider never answers, so the instances stay
    // non-terminal and a second trigger must be refused rather than duplicating them.
    const first = await triggerExecution(PHASE_ID)
    expect(first.status).toBe(202)

    await expect
      .poll(async () => JSON.parse((await listInstances(PHASE_ID)).body).length, {
        timeout: 15_000,
      })
      .toBeGreaterThan(0)
    const countAfterFirst = JSON.parse((await listInstances(PHASE_ID)).body).length

    const second = await triggerExecution(PHASE_ID)
    expect(second.status).toBe(409)

    // The refused trigger must not have created a second set of instances.
    const instances = JSON.parse((await listInstances(PHASE_ID)).body)
    expect(instances.length).toBe(countAfterFirst)
    for (const instance of instances) {
      expect(['pending', 'in_progress']).toContain(instance.status)
    }
  })
})
