import assert from 'node:assert/strict'
import { existsSync, readdirSync } from 'node:fs'
import { describe, it } from 'node:test'
import { fileURLToPath } from 'node:url'
import { loadShardPatterns, MODULES, select, selectAll } from './ci-changes.mjs'

process.chdir(fileURLToPath(new URL('..', import.meta.url)))

const everything = selectAll()

describe('select', () => {
  it('skips CI for docs-only changes', () => {
    assert.deepEqual(select(['README.md', 'docs/intro.mdx', 'clients/core/readme.md', 'LICENSE']), {
      code: false,
      servers: [],
      clients: [],
      shards: [],
    })
  })

  it('runs everything for files outside any module', () => {
    for (const file of [
      'e2e/src/fixtures.ts',
      'e2e/tests/assessment/helpers.ts',
      'seed/core.sql',
      'docker-compose.e2e.yml',
      '.github/workflows/dev.yml',
      'api-stress/uv.lock',
      'keycloakConfig.json',
    ]) {
      assert.deepEqual(select(['servers/interview/main.go', file]), everything, file)
    }
  })

  it('runs every client and shard but no server for shared client files', () => {
    for (const file of [
      'clients/package.json',
      'clients/yarn.lock',
      'clients/Dockerfile',
      'clients/shared/rspack/createRspackConfig.mjs',
      'clients/.gitignore',
    ]) {
      assert.deepEqual(select([file]), { ...everything, servers: [] }, file)
    }
  })

  it('runs every server and shard but no client for shared server files', () => {
    for (const file of ['servers/Dockerfile', 'servers/new_phase/main.go']) {
      assert.deepEqual(select([file]), { ...everything, clients: [] }, file)
    }
  })

  it('runs every shard but only the changed server for a core server change', () => {
    assert.deepEqual(select(['servers/core/main.go']), {
      ...everything,
      servers: ['core'],
      clients: [],
    })
  })

  it('runs every shard but only the changed client for a core client change', () => {
    assert.deepEqual(select(['clients/core/src/App.tsx']), {
      ...everything,
      servers: [],
      clients: ['core'],
    })
  })

  it('runs a leaf module with the core shard only', () => {
    assert.deepEqual(
      select([
        'servers/infrastructure_setup/main.go',
        'clients/infrastructure_setup_component/src/x.tsx',
      ]),
      {
        code: true,
        servers: ['infrastructure_setup'],
        clients: ['infrastructure_setup_component'],
        shards: ['core', 'infrastructure-setup'],
      },
    )
  })

  it('follows the dependency graph transitively', () => {
    assert.deepEqual(select(['servers/team_allocation/main.go']).shards, [
      'core',
      'team-allocation',
      'self-team-allocation',
      'assessment',
      'certificate',
      'presentation',
      'ai',
    ])
    assert.deepEqual(select(['clients/matching_component/src/x.tsx']).shards, [
      'core',
      'matching',
      'assessment',
      'ai',
    ])
    assert.deepEqual(select(['servers/example_server/main.go']).shards, ['core', 'ai', 'example'])
  })

  it('reruns the phases that read team data for a team phase change', () => {
    for (const file of ['servers/team_allocation/main.go', 'servers/self_team_allocation/main.go']) {
      assert.ok(select([file]).shards.includes('presentation'), file)
    }
  })

  it('reruns matching and its dependents for an interview change', () => {
    assert.deepEqual(select(['servers/interview/main.go']).shards, [
      'core',
      'interview',
      'matching',
      'assessment',
      'ai',
    ])
  })

  it('does not confuse modules sharing a name prefix', () => {
    assert.deepEqual(select(['servers/self_team_allocation/main.go']).servers, [
      'self_team_allocation',
    ])
    assert.deepEqual(select(['servers/self_team_allocation/main.go']).shards, [
      'core',
      'self-team-allocation',
      'assessment',
      'certificate',
      'presentation',
      'ai',
    ])
  })

  it('runs only the owning shard for a spec change', () => {
    assert.deepEqual(
      select(['e2e/tests/interview/schedule.spec.ts', 'e2e/tests/api/matching.api.spec.ts']),
      {
        code: true,
        servers: [],
        clients: [],
        shards: ['core', 'interview', 'matching'],
      },
    )
  })
})

describe('module registry', () => {
  const registered = (part) => Object.values(MODULES).map((module) => module[part])

  it('registers every server', () => {
    const servers = readdirSync('servers', { withFileTypes: true })
      .filter((entry) => entry.isDirectory() && existsSync(`servers/${entry.name}/go.mod`))
      .map((entry) => entry.name)
    assert.deepEqual(servers.sort(), registered('server').filter(Boolean).sort())
  })

  it('registers every client', () => {
    const clients = readdirSync('clients', { withFileTypes: true })
      .filter(
        (entry) =>
          entry.isDirectory() && (entry.name === 'core' || entry.name.endsWith('_component')),
      )
      .map((entry) => entry.name)
    assert.deepEqual(clients.sort(), registered('client').filter(Boolean).sort())
  })

  it('maps every e2e shard to exactly one module', () => {
    const shards = loadShardPatterns().map(({ name }) => name)
    assert.deepEqual(shards.sort(), registered('shard').filter(Boolean).sort())
  })

  it('only depends on registered modules', () => {
    for (const module of Object.values(MODULES)) {
      for (const dep of module.dependsOn ?? []) assert.ok(MODULES[dep], dep)
    }
  })
})
