import { createRequire } from 'node:module'
import { describe, expect, it } from 'vitest'
import packageJson from '../../package.json'
import { federatedDependencies } from './federatedDependencies.mjs'

const rootVersion = (name: keyof typeof packageJson.dependencies) => ({
  singleton: true,
  requiredVersion: packageJson.dependencies[name],
})

describe('federatedDependencies', () => {
  it('shares every package the remotes must not duplicate, at the root version', () => {
    expect(federatedDependencies()).toEqual({
      react: rootVersion('react'),
      'react-dom': rootVersion('react-dom'),
      'react-router-dom': rootVersion('react-router-dom'),
      '@tanstack/react-query': rootVersion('@tanstack/react-query'),
      '@tumaet/prompt-shared-state': rootVersion('@tumaet/prompt-shared-state'),
      '@tumaet/prompt-ui-components': rootVersion('@tumaet/prompt-ui-components'),
      recharts: rootVersion('recharts'),
    })
  })

  // ChartContainer and the charts composed into it must share one recharts module. If the UI
  // library ever needs a recharts range the root does not satisfy, yarn nests a second copy under
  // it and every build ships two versions again.
  it('resolves the UI library to the root recharts copy', () => {
    const rootRequire = createRequire(import.meta.url)
    const uiLibraryRequire = createRequire(rootRequire.resolve('@tumaet/prompt-ui-components'))

    expect(uiLibraryRequire.resolve('recharts')).toBe(rootRequire.resolve('recharts'))
  })
})
