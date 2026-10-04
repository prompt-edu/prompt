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
    })
  })
})
