#!/usr/bin/env node
// Reads e2e/shards.json: check | matrix | paths <name> | names.

import { readFileSync, readdirSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const e2eDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const testsDir = path.join(e2eDir, 'tests')

// Mirrors Playwright's forceRegExp (packages/playwright/src/util.ts).
function forceRegExp(pattern) {
  const match = pattern.match(/^\/(.*)\/([gi]*)$/)
  return match ? new RegExp(match[1], match[2]) : new RegExp(pattern, 'gi')
}

// Leading-slash relative paths stand in for the absolute ones Playwright matches.
function specPaths() {
  const walk = (dir) =>
    readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
      const full = path.join(dir, entry.name)
      if (entry.isDirectory()) return walk(full)
      return entry.isFile() && entry.name.endsWith('.spec.ts') ? [full] : []
    })

  return walk(testsDir)
    .map((full) => `/${path.relative(e2eDir, full).split(path.sep).join('/')}`)
    .sort()
}

function loadShards() {
  const { shards } = JSON.parse(readFileSync(path.join(e2eDir, 'shards.json'), 'utf8'))
  return shards
}

function check() {
  const shards = loadShards()
  const specs = specPaths()
  const errors = []

  if (specs.length === 0) errors.push(`no *.spec.ts files found under ${testsDir}`)

  // Every spec must be claimed by exactly one shard.
  for (const spec of specs) {
    const owners = shards
      .filter((shard) => shard.paths.some((pattern) => forceRegExp(pattern).test(spec)))
      .map((shard) => shard.name)

    if (owners.length === 0) errors.push(`${spec} is not covered by any shard`)
    else if (owners.length > 1) errors.push(`${spec} is claimed by ${owners.join(', ')}`)
  }

  // And every pattern must still claim something, so stale entries surface.
  for (const shard of shards) {
    for (const pattern of shard.paths) {
      if (!specs.some((spec) => forceRegExp(pattern).test(spec)))
        errors.push(`shard "${shard.name}" pattern ${pattern} matches no spec`)
    }
  }

  const writers = shards.filter((shard) => shard.cacheWriter).map((shard) => shard.name)
  if (writers.length !== 1)
    errors.push(`exactly one shard must set cacheWriter, found ${writers.length}`)

  if (errors.length > 0) {
    console.error('Shard partition is invalid:')
    for (const error of errors) console.error(`  - ${error}`)
    console.error('\nFix e2e/shards.json so every spec is claimed exactly once.')
    process.exit(1)
  }

  console.log(`Shard partition OK: ${specs.length} specs across ${shards.length} shards.`)
  for (const shard of shards) {
    const owned = specs.filter((spec) =>
      shard.paths.some((pattern) => forceRegExp(pattern).test(spec)),
    ).length
    console.log(`  ${shard.name.padEnd(22)} ${owned} specs`)
  }
}

function matrix() {
  const include = loadShards().map((shard) => ({
    name: shard.name,
    paths: shard.paths.join(' '),
    cacheWriter: Boolean(shard.cacheWriter),
  }))
  console.log(JSON.stringify({ include }))
}

function paths(name) {
  const shard = loadShards().find((candidate) => candidate.name === name)
  if (!shard) {
    console.error(
      `Unknown shard "${name}". Available: ${loadShards()
        .map((candidate) => candidate.name)
        .join(', ')}`,
    )
    process.exit(1)
  }
  console.log(shard.paths.join(' '))
}

const [command, argument] = process.argv.slice(2)

switch (command) {
  case 'check':
    check()
    break
  case 'matrix':
    matrix()
    break
  case 'paths':
    paths(argument)
    break
  case 'names':
    console.log(
      loadShards()
        .map((shard) => shard.name)
        .join('\n'),
    )
    break
  default:
    console.error('Usage: node scripts/shards.mjs <check|matrix|paths <name>|names>')
    process.exit(2)
}
