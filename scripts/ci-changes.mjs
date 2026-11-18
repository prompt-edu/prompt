#!/usr/bin/env node
// Picks the servers, clients and e2e shards CI runs for a set of changed files.
// Usage: git diff --name-only --no-renames A B | node scripts/ci-changes.mjs
//        node scripts/ci-changes.mjs --all
// Prints GITHUB_OUTPUT lines: code, servers, clients, shards.

import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Every module depends on core. `dependsOn` lists the modules whose data a module's e2e shard
// reads at runtime, so a change there reruns that shard too.
export const MODULES = {
  core: { server: 'core', client: 'core', shard: 'core' },
  interview: { server: 'interview', client: 'interview_component', shard: 'interview' },
  matching: { client: 'matching_component', shard: 'matching' },
  team_allocation: {
    server: 'team_allocation',
    client: 'team_allocation_component',
    shard: 'team-allocation',
  },
  self_team_allocation: {
    server: 'self_team_allocation',
    client: 'self_team_allocation_component',
    shard: 'self-team-allocation',
    dependsOn: ['team_allocation'],
  },
  assessment: {
    server: 'assessment',
    client: 'assessment_component',
    shard: 'assessment',
    dependsOn: ['team_allocation', 'self_team_allocation', 'matching'],
  },
  certificate: {
    server: 'certificate',
    client: 'certificate_component',
    shard: 'certificate',
    dependsOn: ['team_allocation', 'self_team_allocation'],
  },
  presentation: { server: 'presentation', client: 'presentation_component', shard: 'presentation' },
  infrastructure_setup: {
    server: 'infrastructure_setup',
    client: 'infrastructure_setup_component',
    shard: 'infrastructure-setup',
  },
  ai: { server: 'ai', client: 'ai_component', shard: 'ai', dependsOn: ['assessment', 'example'] },
  example: { server: 'example_server', client: 'example_component', shard: 'example' },
}

// Keep in sync with the push trigger's paths-ignore in dev.yml.
const IGNORED = [
  (file) => file.endsWith('.md'),
  (file) => file.startsWith('docs/'),
  (file) => file.startsWith('.github/ISSUE_TEMPLATE/'),
  (file) =>
    [
      '.gitignore',
      'LICENSE',
      '.github/workflows/build-docs.yml',
      '.github/workflows/discord-notifications.yml',
    ].includes(file),
]

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

export function loadShardPatterns() {
  const { shards } = JSON.parse(readFileSync(path.join(repoRoot, 'e2e/shards.json'), 'utf8'))
  return shards.map(({ name, paths }) => ({ name, patterns: paths.map((p) => new RegExp(p)) }))
}

const all = (part) =>
  Object.values(MODULES)
    .map((module) => module[part])
    .filter(Boolean)

export function selectAll() {
  return { code: true, servers: all('server'), clients: all('client'), shards: all('shard') }
}

function dependents(names) {
  const affected = new Set(names)
  let grew = true
  while (grew) {
    grew = false
    for (const [name, module] of Object.entries(MODULES)) {
      if (!affected.has(name) && (module.dependsOn ?? []).some((dep) => affected.has(dep))) {
        affected.add(name)
        grew = true
      }
    }
  }
  return affected
}

function owner(file, shardPatterns) {
  for (const [name, module] of Object.entries(MODULES)) {
    if (module.server && file.startsWith(`servers/${module.server}/`))
      return { modules: [name], part: 'server' }
    if (module.client && file.startsWith(`clients/${module.client}/`))
      return { modules: [name], part: 'client' }
  }
  // Shared files of one side (the lockfile, a Dockerfile, clients/shared) reach all its modules.
  if (file.startsWith('servers/')) return { modules: Object.keys(MODULES), part: 'server' }
  if (file.startsWith('clients/')) return { modules: Object.keys(MODULES), part: 'client' }
  if (file.startsWith('e2e/tests/')) {
    const spec = `/${file.slice('e2e/'.length)}`
    const shard = shardPatterns.find(({ patterns }) => patterns.some((p) => p.test(spec)))
    if (shard) return { shard: shard.name }
  }
  return null
}

export function select(files, shardPatterns = loadShardPatterns()) {
  const code = files.filter((file) => !IGNORED.some((ignored) => ignored(file)))
  if (code.length === 0) return { code: false, servers: [], clients: [], shards: [] }

  const servers = new Set()
  const clients = new Set()
  const changed = new Set()
  const shards = new Set(['core'])

  for (const file of code) {
    const found = owner(file, shardPatterns)
    if (!found) return selectAll()
    if (found.shard) shards.add(found.shard)
    for (const name of found.modules ?? []) {
      ;(found.part === 'server' ? servers : clients).add(name)
      changed.add(name)
    }
  }

  const affected = changed.has('core') ? Object.keys(MODULES) : dependents(changed)
  for (const name of affected) shards.add(MODULES[name].shard)

  return {
    code: true,
    servers: pick(servers, 'server'),
    clients: pick(clients, 'client'),
    shards: inOrder(shards),
  }
}

const pick = (names, part) =>
  Object.entries(MODULES)
    .filter(([name, module]) => names.has(name) && module[part])
    .map(([, module]) => module[part])

const inOrder = (shards) => all('shard').filter((shard) => shards.has(shard))

export function toOutput({ code, servers, clients, shards }) {
  return [
    `code=${code}`,
    `servers=${JSON.stringify(servers)}`,
    `clients=${JSON.stringify(clients)}`,
    `shards=${JSON.stringify(shards)}`,
  ].join('\n')
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const selection = process.argv.includes('--all')
    ? selectAll()
    : select(readFileSync(0, 'utf8').split('\n').filter(Boolean))
  console.log(toOutput(selection))
}
